# LanGive Android 客户端 — 开发说明书

> 这份说明给负责实现 Android 端 UI 与 SAF 集成的工程师阅读。Go 服务层（设备发现、配置、文件传输、安全、自动更新）已完成、跨平台，不需要在 Android 端重写。本目录的目标是产出一个 Android 工程，把已编译好的 `langive.aar` 嵌进去，并在原生 UI 上接通。

---

## 1. 项目背景

LanGive 是一个跨平台局域网文件传输工具。桌面端基于 Wails v2（Go + Vue 3），代码在仓库根目录。Android 端复用同一个 Go 内核：

- `internal/config` — JSON 配置（设备名、UUID、下载路径、扫描间隔等）
- `internal/mdns` — `_langive._tcp.local` 注册与发现
- `internal/transfer` — Gin HTTP 服务（端口 5566），实现文件接收/发送、断点续传、二段式传输请求 + 一次性 token 鉴权
- `internal/security` — 会话与 token 管理
- `internal/updater` — 自动更新（Android 端**不要**调用，桌面专属）

这五个包通过 `gomobile bind -target=android` 打包成 `langive.aar`，Android 工程通过 JNI 直接调用其中的导出函数。

---

## 2. 必须落地的工程结构

请在 `mobile/android/` 下生成完整 Gradle 工程，参考结构：

```
mobile/android/
├── build.gradle.kts                # 根 build script
├── settings.gradle.kts
├── gradle.properties
├── gradle/wrapper/
├── gradlew / gradlew.bat
└── app/
    ├── build.gradle.kts            # 模块 build script，dependencies 引入 :langive
    ├── proguard-rules.pro
    ├── libs/
    │   └── langive.aar             # 由 CI 的 gomobile 步骤注入，本地开发先放占位
    └── src/main/
        ├── AndroidManifest.xml
        ├── java/com/langive/app/
        │   ├── MainActivity.kt
        │   ├── ui/                 # Compose 屏幕
        │   ├── service/            # 前台服务（持续 mDNS 与接收）
        │   └── bridge/             # 与 langive.aar 的 Kotlin wrapper
        └── res/
            └── values/strings.xml  # 中文为主
```

CI 已经准备好这一段（`.github/workflows/release.yml` 的 `build-android` job），它会执行：

```bash
gomobile bind -target=android -androidapi 21 \
  -o mobile/android/app/libs/langive.aar \
  ./internal/mdns ./internal/transfer ./internal/config ./internal/security
```

随后 `cd mobile/android && ./gradlew assembleRelease`。所以你要确保 Gradle 工程在没有 `app/libs/langive.aar` 时能 sync（用 `compileOnly` 占位 + CI 注入），但在 release 构建时必须有真实 AAR。

---

## 3. AAR 暴露的 API（重点）

`gomobile bind` 会把每个 Go 包变成一个 Java 包。命名规则：`<modulePrefix>.<packageName>`，默认 `go.<package>` → 但常见做法是 `mobile init` 后用 `-javapkg=com.langive.bridge`，得到：

- `com.langive.bridge.config.Config`
- `com.langive.bridge.mdns.Service / DeviceInfo`
- `com.langive.bridge.transfer.Service / TransferStatus / IncomingRequest`
- `com.langive.bridge.security.Manager`

请在 CI 的 `gomobile bind` 命令上加 `-javapkg=com.langive.bridge`，并在文档与 wrapper 中保持一致。

### 关键函数（导出名以 Go 大写为准，转 Java 后首字母小写）

**配置 (config)**

```kotlin
val cfg = Config.load()              // *config.Config
cfg.deviceName                       // String，gomobile 自动生成 getter/setter
cfg.deviceUUID
cfg.downloadPath                     // 默认值由 Go 端给出
cfg.port
cfg.privacyMode                      // 隐私模式 = 不广播到公共会话
cfg.sessionID
cfg.scanInterval                     // 秒
cfg.save()                           // 持久化到 JSON
cfg.setPrivacyMode(true)             // 切换隐私模式会换 SessionID
cfg.getScanInterval()                // 返回 long(纳秒)
cfg.getDeviceTimeout()               // 返回 long(纳秒)，约定 = ScanInterval × 3
```

**mDNS 发现**

```kotlin
val mdns = Mdns.newService(
    cfg.deviceName, cfg.deviceUUID, cfg.port,
    "1.0.0", cfg.sessionID, cfg.privacyMode
)
mdns.start()
mdns.startCleanupRoutine(cfg.scanInterval, cfg.deviceTimeout)

// 周期取列表
val list: DeviceInfoSlice = mdns.getPublicDevices()  // 自定义 wrapper 转成 Kotlin List
for (i in 0 until list.size()) {
    val d: DeviceInfo = list.get(i)
    // d.id / d.name / d.address / d.port / d.platform / d.uuid / d.privacy / d.isPublic
}

mdns.setDeviceName("新名字")     // 内部会重启 mDNS
mdns.setPrivacy(true)
mdns.stop()
```

`DeviceInfo` 是结构体，gomobile 给字段生成 getter（`getName()`, `getAddress()` …）。`gomobile` 不直接支持返回 Go slice，因此 `GetPublicDevices() []DeviceInfo` 在 Java 侧表现为一个不太友好的代理对象。**请在 Kotlin wrapper 里做一次封装**，对外暴露 `List<Device>`。

**文件传输**

接收端只需要把服务起来，监听端口 5566：

```kotlin
val ts = Transfer.newService(cfg.downloadPath, cfg.port)
ts.setOnIncomingRequest { req: IncomingRequest ->
    // 在 UI 线程弹接收对话框（详见 §5）
}
ts.start()
```

**二段式握手是强制的**：

1. 发送端调 `POST /transfer/request` → 接收端 `setOnIncomingRequest` 回调触发，UI 弹窗
2. UI 选择"接收" → 调 `ts.approveIncoming(id)` 拿到一次性 token（5 分钟有效）
3. 发送端拿 token 才能 `POST /upload`，否则后端 401

发送时反过来：

```kotlin
ts.sendFilesAs(device.address, cfg.deviceName, listOf("/storage/emulated/0/Download/foo.jpg"))
ts.sendFolderAs(device.address, cfg.deviceName, "/storage/.../some-dir")
```

> **重点：路径在 Android 上是难题。** Android 11+ 对外部存储基本只能用 SAF（Storage Access Framework）拿 `Uri`，没有真实文件路径。建议：
> 1. 用户选择文件 → SAF 给你 `Uri`
> 2. Kotlin 层 `contentResolver.openInputStream(uri)` 读出来，写到 app 私有目录的临时文件
> 3. 把临时文件路径传给 `sendFilesAs`
> 4. 发送结束后清理临时文件
>
> 接收端反过来：Go 服务把文件落到 `cfg.downloadPath`（必须是 app 可写目录，比如 `getExternalFilesDir(null)` 或 `filesDir`）。要让用户在系统文件应用看到，可以再用 `MediaStore` API 复制一份过去，或要求用户用 SAF 自己挑收件夹然后把那条路径写进 `cfg.downloadPath`。

**传输列表 / 取消 / 清空**

```kotlin
ts.getTransfers()                    // 返回 TransferStatus 列表（同样要 wrapper）
ts.cancelTransfer(id)
ts.clearCompleted()
ts.pendingRequests()                 // 启动时拉一次，避免错过未处理的待确认请求
```

`TransferStatus` 字段：`id, type("send"/"receive"), fileName, totalSize, sentSize, progress (0-100), status, peerAddr, error`。

`status` 取值：`pending | transferring | completed | failed | cancelled | paused`。

---

## 4. AndroidManifest 必备项

```xml
<!-- 网络权限 -->
<uses-permission android:name="android.permission.INTERNET" />
<uses-permission android:name="android.permission.ACCESS_NETWORK_STATE" />
<uses-permission android:name="android.permission.ACCESS_WIFI_STATE" />
<uses-permission android:name="android.permission.CHANGE_WIFI_MULTICAST_STATE" />

<!-- 前台服务（让 mDNS / 接收持续运行） -->
<uses-permission android:name="android.permission.FOREGROUND_SERVICE" />
<uses-permission android:name="android.permission.FOREGROUND_SERVICE_DATA_SYNC" />
<!-- Android 13+ 通知权限（前台服务通知 + 接收请求弹窗） -->
<uses-permission android:name="android.permission.POST_NOTIFICATIONS" />

<application
    android:label="LanGive"
    android:usesCleartextTraffic="true"  <!-- 局域网 HTTP 必需 -->
    android:networkSecurityConfig="@xml/network_security_config">
    <service
        android:name=".service.LanGiveService"
        android:foregroundServiceType="dataSync"
        android:exported="false" />
</application>
```

`network_security_config.xml` 显式允许 cleartext 到本地局域网（10/172.16/192.168 段）即可，不要全局开放。

mDNS 多播：进入扫描页时取 `WifiManager.MulticastLock`，离开时释放，否则部分机型 Wi-Fi 节能会丢包。

---

## 5. UI 要求（与桌面版对齐）

桌面端用了 Tailwind + lucide 做了 4 个屏幕。Android 端用 **Jetpack Compose + Material 3**，配色与桌面对齐：主色 `#3a64ff`，深色背景 `#0b1020`。屏幕：

1. **首页** — 在线设备数 / 传输记录数 / 当前版本，"打开下载文件夹"按钮（用 `Intent.ACTION_OPEN_DOCUMENT_TREE` 让用户挑文件夹查看）
2. **设备** — 卡片列表，平台图标。点击 → 系统选文件器（`ACTION_OPEN_DOCUMENT` / `ACTION_OPEN_DOCUMENT_TREE`）
3. **传输** — 进度条 + 状态徽章；transferring 状态可以取消；底部"清除已完成"按钮调 `clearCompleted()`
4. **设置** — 设备名 / 下载路径（SAF tree URI 持久化授权）/ 端口 / 扫描间隔 / 隐私模式开关

### 接收端弹窗（必须做）

服务启动后注册 `setOnIncomingRequest`。回调来自 Go 协程，**不能直接更新 Compose 状态**，要 hop 到 main thread：

```kotlin
ts.setOnIncomingRequest { req ->
    mainHandler.post {
        // 推到 ViewModel 的 SnackBar/Dialog flow
        viewModel.enqueueIncoming(req)
    }
}
```

UI 给两个按钮：
- 接收 → `ts.approveIncoming(req.id)`
- 拒绝 → `ts.rejectIncoming(req.id)`

未在 60 秒内做出选择，Go 端会自动按拒绝处理（这是 transfer.go 已实现的行为，不要在 Android 端再做超时）。

启动 App 时主动调 `ts.pendingRequests()` 拉一次，把没来得及处理的请求补上。

---

## 6. 前台服务

接收端必须在前台服务里跑 mDNS + transfer，不然系统会杀进程。

服务里做：
1. 创建 `notification channel`，发一条"LanGive 正在监听局域网传输"的常驻通知
2. 启动 `mdns.Service` 与 `transfer.Service`
3. 注册 `setOnIncomingRequest`，把请求广播给前台 Activity（用 `LocalBroadcastManager` 或 `MutableSharedFlow`）
4. `onDestroy` 调 `mdns.stop()` / `ts.stop()`

后台时把 `cfg.setBackgroundMode(true)`，扫描间隔会自动切到 30s，省电。

---

## 7. 自动更新

Android 端**不要**调 `internal/updater`。Android 走 Google Play / GitHub Releases APK 直链，由用户手动更新。请在「设置 → 关于」里写明这一点。

---

## 8. 构建链路对照

桌面端跑通的 release 流程：

1. `scripts/release.sh patch`（或 minor/major）
2. 同步 `internal/config/config.go` 中 `Version` 常量、`frontend/package.json`、`wails.json`
3. 生成 `CHANGELOG.md` + `.release/notes.md`
4. `git tag -a vX.Y.Z` 推送
5. GitHub Actions `release.yml` 跑 Windows/macOS/Linux/Android 矩阵
6. release notes 显示 changelog 内容

**Android 工程必须保证 CI 命令 `cd mobile/android && ./gradlew assembleRelease` 能成功**，并把 `app/build/outputs/apk/release/*.apk` 作为 artifact 上传。这是当前 CI 失败的主因——目前 `mobile/android/` 是空目录。

---

## 9. 验收清单

- [ ] `./gradlew assembleDebug` 在没有 `langive.aar` 的情况下能 sync（仅给开发者用 `compileOnly` 占位的 stub 即可）
- [ ] CI 注入 `langive.aar` 后 `./gradlew assembleRelease` 成功，输出 APK
- [ ] 装上 APK 后，与 Linux/Windows/macOS 桌面端可以互发文件
- [ ] 收到文件前显示弹窗，拒绝/超时不会写盘
- [ ] 隐私模式下设备不出现在桌面端「设备」列表
- [ ] 锁屏后再亮屏，传输能继续（前台服务工作）
- [ ] 切到后台 30s 后扫描频率明显下降（看日志或 `getScanInterval()`）

---

## 10. 给 Gemini 的提示词建议

以下整段话可以直接喂给 Gemini，让它生成 Android 工程：

> 你有一个 Go 写的 LanGive 内核（mDNS 设备发现 + HTTP 文件传输 + JSON 配置 + 二段式 token 鉴权），通过 `gomobile bind -target=android -javapkg=com.langive.bridge -o mobile/android/app/libs/langive.aar ./internal/mdns ./internal/transfer ./internal/config ./internal/security` 已经打成 AAR，包路径 `com.langive.bridge.{config,mdns,transfer,security}`。请在 `mobile/android/` 下生成完整 Gradle (KTS) 工程：minSdk 21 / targetSdk 34，Kotlin + Jetpack Compose + Material 3，主色 `#3a64ff`，4 个屏幕（Home / Devices / Transfers / Settings）+ 接收请求弹窗 + 前台服务（dataSync 类型常驻）。文件选择走 SAF（`ACTION_OPEN_DOCUMENT[_TREE]`），把 `Uri` 转成临时文件后传给 `transfer.SendFilesAs(deviceAddress, deviceName, listOf(tempPath))`。接收回调通过 `transfer.SetOnIncomingRequest` 注册，弹窗后调 `ApproveIncoming(id)` / `RejectIncoming(id)`。设置页要写下载路径（持久化 SAF tree URI）、设备名、端口、扫描间隔、隐私模式开关、自动更新提示「Android 不支持自动更新，请到 GitHub Releases 下载新版」。所有 UI 文案使用中文。AndroidManifest 必须包含 INTERNET / ACCESS_WIFI_STATE / CHANGE_WIFI_MULTICAST_STATE / FOREGROUND_SERVICE / FOREGROUND_SERVICE_DATA_SYNC / POST_NOTIFICATIONS，`usesCleartextTraffic=true` 并提供 `network_security_config.xml` 仅放行 RFC1918 段。读取 `langive.aar` 中的列表/结构体时用 Kotlin wrapper 转成 `List<DeviceInfo>` 等友好类型。你不需要修改任何 Go 代码或 `.github/workflows/release.yml`，只产出 Android 工程。

Gemini 输出工程后，把 `mobile/android/` 整个落到本仓库，再触发一次 release tag 即可。

---

*文档版本：v1.0.0 — 2026-05-25*

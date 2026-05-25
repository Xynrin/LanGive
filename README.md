# LanGive

一款基于 mDNS 协议的跨平台局域网文件传输工具，支持 Windows、macOS、Linux、iOS 和 Android。

![License](https://img.shields.io/badge/license-GPLv3-blue.svg)
![Version](https://img.shields.io/badge/version-1.0.0-green.svg)
![Go](https://img.shields.io/badge/Go-1.23-blue.svg)
![Wails](https://img.shields.io/badge/Wails-v2-orange.svg)

## 特性

- 🔍 **自动发现** - 基于 mDNS 协议自动发现局域网内的设备
- 🔒 **会话隔离** - 公共会话与隐私模式，保护你的传输安全
- 📁 **文件传输** - 支持文件和文件夹的快速传输
- 🖥️ **跨平台** - 支持 Windows、macOS、Linux、iOS 和 Android
- 🔄 **实时进度** - 传输进度实时显示
- 📝 **自定义设备名** - 支持自定义在局域网中显示的设备名称
- 🆕 **自动更新** - 支持版本检查和自动更新
- ⚡ **会话续接** - 后台运行时自动降低扫描频率，节省资源

## 系统要求

### 桌面端
- Windows 10/11 (amd64, x86, arm64)
- macOS 10.15+ (Intel/Apple Silicon)
- Linux (amd64, x86, arm64)

### 移动端
- iOS 14+ (arm64)
- Android 7.0+ (arm64, x86, x86_64)

## 快速开始

### 安装

前往 [Releases](https://github.com/Xynrin/LanGive/releases) 页面下载对应平台的安装包：

**桌面端**
| 平台 | 架构 | 下载格式 |
|------|------|----------|
| Windows | amd64/x86/arm64 | `.exe` (NSIS) |
| macOS | Intel/Apple Silicon | `.app` / `.zip` |
| Linux | amd64/x86/arm64 | AppImage / deb / RPM |

**移动端**
- **Android**: 下载 `LanGive-android-*.apk`
- **iOS**: 通过 TestFlight 安装（即将推出）

### 使用

1. 在需要传输文件的设备上启动 LanGive
2. 在"设备"页面查看发现的设备
3. 选择目标设备并点击"发送"
4. 选择要发送的文件或文件夹
5. 等待传输完成

## 开发

### 环境要求

- Go 1.23+
- Node.js 20+
- Wails CLI v2

### 安装依赖

```bash
# 安装 Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 安装前端依赖
cd frontend
npm install
```

### 开发模式

```bash
wails dev
```

### 构建

```bash
# 构建当前平台
wails build

# 构建特定平台
wails build -platform windows/amd64
wails build -platform darwin/arm64
wails build -platform linux/amd64
```

## 项目结构

```
LanGive/
├── main.go                    # 应用入口
├── app.go                     # Wails 应用逻辑
├── wails.json                # Wails 配置
├── go.mod                    # Go 模块
│
├── internal/                  # 内部包
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── mdns/                 # mDNS 服务发现
│   │   └── mdns.go
│   ├── transfer/             # 文件传输
│   │   └── transfer.go
│   ├── security/              # 安全与会话
│   │   ├── security.go
│   │   └── token.go
│   └── updater/              # 自动更新
│       └── updater.go
│
├── frontend/                  # 前端代码
│   ├── src/
│   │   ├── components/       # Vue 组件
│   │   │   └── Sidebar.vue
│   │   ├── views/            # 页面视图
│   │   │   ├── Home.vue
│   │   │   ├── Devices.vue
│   │   │   ├── Transfers.vue
│   │   │   └── Settings.vue
│   │   ├── router/           # 路由配置
│   │   │   └── index.js
│   │   ├── App.vue
│   │   ├── main.js
│   │   └── style.css
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
│
├── mobile/                    # 移动端代码
│   ├── android/              # Android 项目
│   └── ios/                  # iOS 项目
│
├── assets/                    # 静态资源
│   └── logo.jpg
│
├── docs/                     # 文档
│   └── SPEC.md               # 详细设计文档
│
├── build/                    # 构建输出
│   └── bin/
│
└── .github/
    └── workflows/            # CI/CD 配置
        └── release.yml
```

## 技术栈

| 类别 | 技术 |
|------|------|
| 后端 | Go |
| 前端 | Vue 3 + Vue Router + Pinia |
| UI 框架 | Wails v2 |
| 服务发现 | mDNS (hashicorp/mdns) |
| 文件传输 | TCP + HTTP |
| 构建 | GitHub Actions |

## 架构设计

### 模块划分

1. **设备发现模块** (`internal/mdns/`)
   - mDNS 服务注册与发现
   - 设备状态管理
   - TXT 记录解析

2. **传输模块** (`internal/transfer/`)
   - TCP 连接管理
   - 文件分段传输
   - ZIP 打包发送

3. **安全模块** (`internal/security/`)
   - 会话管理
   - Token 验证
   - 隐私模式

4. **配置模块** (`internal/config/`)
   - JSON 配置读写
   - 跨平台路径处理

5. **更新模块** (`internal/updater/`)
   - Git 版本检查
   - 下载与安装

## 隐私与会话

### 公共会话
- 默认加入公共会话
- 所有未开启隐私模式的设备可见
- 适合家庭、办公等信任环境

### 隐私模式
- 不在公共会话中广播
- 只能通过 IP 直接连接
- 适合公共 WiFi 等敏感环境

## License

本项目采用 [GPL-V3](LICENSE) 协议开源。

## 致谢

- [Wails](https://wails.io/) - 跨平台桌面应用框架
- [hashicorp/mdns](https://github.com/hashicorp/mdns) - mDNS 库
- [Vue.js](https://vuejs.org/) - 前端框架
- [Gin](https://gin-gonic.com/) - HTTP web 框架

---

*LanGive - 让局域网文件传输变得简单*

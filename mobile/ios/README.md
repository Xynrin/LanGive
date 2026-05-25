# LanGive iOS 客户端 — 占位说明

iOS 端目前**未规划交付**。原因：

1. `gomobile bind -target=ios` 产出 `Langive.xcframework`，但 iOS 沙盒对 mDNS（`_langive._tcp`）有更严格限制（iOS 14+ 需要 `NSLocalNetworkUsageDescription` + `NSBonjourServices` Info.plist 声明 + 用户手动允许「本地网络」权限），且 14.5 后还要求 `NSAdvertisingAttributionReportEndpoint` 等额外配置；
2. 后台保活策略与 Android 完全不同，前台服务/常驻通知模型不通用；
3. 文件系统沙盒：iOS 没有"下载文件夹"概念，需要走 `UIDocumentPicker` + Files App 集成；
4. CI 需要 macOS runner + 苹果开发者账号签名才能产出 IPA。

短期不规划。如果未来要做，建议另开 Issue 详细论证。

桌面端（macOS / Windows / Linux）+ Android 已覆盖目标用户群。

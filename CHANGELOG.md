# Changelog

## v1.0.13 — 2026-05-25

- fix(ci): 在 npm run build 前先 wails generate module


## v1.0.12 — 2026-05-25

- fix(ci): 显式构建前端，避免 dist 为空导致 embed 失败


## v1.0.11 — 2026-05-25

- fix(linux-pkg): 修复包装内缺图标/desktop/运行时依赖问题


## v1.0.10 — 2026-05-25



## v1.0.9 — 2026-05-25

- fix(android): bridge Config 改名为 Settings，修 Kotlin 编译错误


## v1.0.8 — 2026-05-25

- fix(android): manifest 移除尚未提供的 mipmap 图标引用


## v1.0.7 — 2026-05-25

- fix(android): bridge Config 字段改为不导出，避免 gomobile 重定义


## v1.0.6 — 2026-05-25

- fix(android): 新增 mobile/bridge/ 公开包以适配 gomobile bind


## v1.0.5 — 2026-05-25

- fix(android): 添加 golang.org/x/mobile 依赖


## v1.0.4 — 2026-05-25

- fix(ci): 修复 AppImage libfuse 依赖 + 接入 Android 工程


## v1.0.3 — 2026-05-25

- fix(ci): Linux 改用原生 runner + 强制 webkit2_41


## v1.0.2 — 2026-05-25

- feat: 完成审计报告全部修复 + 前端重写 + 发布工具链


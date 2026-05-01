# cloud-cli

高扩展、微内核架构的网盘 CLI 工具。

## 架构

```
cloud-cli/
├── core/          # 微内核：统一 Storage 接口 + 驱动注册表 + 配置管理
├── backends/      # 可插拔后端驱动（quark, baidu, aliyun...）
├── cmd/           # Cobra CLI 命令层
├── utils/         # 工具：并发池、进度条
└── main.go        # 入口：空白导入驱动，自动注册
```

## 特性

- 🏗️ **微内核设计** — 主程序只通过 `core.Storage` 接口发号施令，完全不在乎操作的是哪个网盘
- 🔌 **插件式驱动** — 新增网盘只需实现接口 + `init()` 注册，零侵入主程序
- 📦 **双输出模式** — 所有命令支持 `--json`，便于 Shell/jq 管道自动化
- 🔒 **安全存储** — Cookie/Token 持久化到 `~/.config/cloud-cli/config.json`，权限 600
- ⚡ **零外部依赖** — 编译后 12MB 单二进制，无需安装 Python/Node 运行时

## 安装

```bash
# 从源码编译
go build -o cloud-cli .
sudo mv cloud-cli /usr/local/bin/

# 或直接下载预编译二进制
```

## 快速开始

```bash
# 登录（粘贴 Cookie）
cloud-cli login quark

# 列出根目录
cloud-cli ls /
cloud-cli ls / --json

# 上传文件
cloud-cli upload local.txt /remote/dir/

# 下载文件
cloud-cli download /remote/file.txt ./local.txt

# 创建目录
cloud-cli mkdir /new/folder

# 删除
cloud-cli rm /unwanted/file.txt
```

## 支持的网盘

| 网盘 | 状态 | 功能 |
|------|------|------|
| 夸克网盘 | ✅ 可用 | ls/upload/download/mkdir/rm/login |
| 百度网盘 | 🚧 待开发 | — |
| 115网盘 | 🚧 待开发 | — |
| 阿里云盘 | 🚧 待开发 | — |

## 开发新驱动

1. 创建 `backends/<provider>/driver.go`
2. 实现 `core.Storage` 接口
3. 在 `init()` 中调用 `core.Register("<provider>", NewDriver)`
4. 在 `main.go` 中空白导入 `_ "backends/<provider>"`

## License

MIT

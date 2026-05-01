# ☁️ cloud-cli — 通用网盘命令行工具

**高性能 · 微内核架构 · 多网盘统一接口 · 跨盘同步传输**

**High-Performance · Microkernel Architecture · Universal Cloud Storage CLI · Cross-Cloud Sync**

---

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Ab-code520/cloud-cli?include_prereleases&sort=semver)](https://github.com/Ab-code520/cloud-cli/releases)
[![Build Status](https://github.com/Ab-code520/cloud-cli/actions/workflows/release.yml/badge.svg)](https://github.com/Ab-code520/cloud-cli/actions)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-informational)](https://github.com/Ab-code520/cloud-cli/releases)

---

## 🌟 为什么选择 cloud-cli？ / Why cloud-cli?

| 特性 / Feature | cloud-cli | rclone | 单网盘 CLI / Single-Provider CLI |
|:---|:---:|:---:|:---:|
| **通用接口 + 专属能力** | ✅ 分层架构，通用底座+扩展接口 | ✅ 功能丰富但配置复杂 | ❌ 功能绑定特定网盘 |
| **夸克网盘支持** | ✅ 完整支持（秒传、分片、断点） | ❌ 不支持 | ✅ 仅夸克 |
| **断点续传** | ✅ 崩溃自动恢复 | ✅ | ❌ 通常不支持 |
| **秒传检测** | ✅ 自动 SHA1 预检 | ❌ | ❌ |
| **跨盘对拷** | ✅ 夸克 ↔ 百度 ↔ 本地 | ✅ | ❌ |
| **目录同步** | ✅ `--dry-run` + `--delete` | ✅ 功能多但学习成本高 | ❌ |
| **配置复杂度** | 🟢 极简，一条命令登录 | 🔴 需配置多个 remote | 🟢 简单 |
| **多账户管理** | ✅ YAML 统一管理 | ✅ 配置文件中管理 | ❌ 通常单账户 |
| **CI/CD 发布** | ✅ Goreleaser 自动构建多平台 | ✅ | ❌ |

### 🏆 核心优势 / Core Advantages

1. **🔌 插件化驱动架构** — 每个网盘只需实现接口即可接入，未来可扩展百度、115、阿里、OneDrive 等。
2. **🧠 智能能力检测** — CLI 自动检测当前网盘支持的功能，不支持的功能优雅降级，不报错。
3. **🔄 跨盘同步引擎** — 基于流式 I/O 的同步引擎，本地与云端、云端与云端之间无缝传输。
4. **⚡ 生产级上传** — 并发分片 + 断点续传 + 秒传检测 + SSRF 防护，专为大文件优化。
5. **📋 极简配置** — 扫码登录 / Cookie 导入，一条命令完成，无需复杂配置文件。

---

## 📦 安装 / Installation

### 方式一：下载预编译二进制 / Download Pre-built Binary

前往 [Releases](https://github.com/Ab-code520/cloud-cli/releases) 下载对应平台的二进制文件：

```bash
# Linux amd64
curl -LO https://github.com/Ab-code520/cloud-cli/releases/download/v1.4.0/cloud-cli_1.4.0_linux_amd64.tar.gz
tar -xzf cloud-cli_1.4.0_linux_amd64.tar.gz
sudo mv cloud-cli /usr/local/bin/
```

### 方式二：从源码编译 / Build from Source

```bash
git clone https://github.com/Ab-code520/cloud-cli.git
cd cloud-cli
go build -o cloud-cli .
sudo mv cloud-cli /usr/local/bin/
```

### 方式三：使用 Homebrew (macOS/Linux) / Use Homebrew

```bash
# Add our tap
brew tap Ab-code520/tap

# Install
brew install cloud-cli

# Upgrade
brew upgrade cloud-cli
```

### 方式四：使用 Go install / Go Install

```bash
go install github.com/Ab-code520/cloud-cli@latest
```

---

## 🚀 快速开始 / Quick Start

### 1️⃣ 登录 / Login

```bash
# 扫码登录夸克（推荐）/ QR Scan Login (Recommended)
cloud-cli login quark

# 或使用 Cookie 登录 / Login with Cookie
cloud-cli login quark --cookie "your_cookie_string_here"
```

### 2️⃣ 浏览文件 / Browse Files

```bash
# 列出根目录 / List root directory
cloud-cli ls

# 列出指定目录 / List specific directory
cloud-cli ls folder_id

# 递归列出 / List recursively (future)
```

### 3️⃣ 上传文件 / Upload

```bash
# 上传单个文件（自动秒传检测 + 断点续传）
cloud-cli upload ./movie.mp4 /videos/

# 使用指定策略 / With upload policy
cloud-cli upload ./movie.mp4 /videos/ --policy overwrite   # 覆盖已有文件
cloud-cli upload ./movie.mp4 /videos/ --policy skip        # 跳过已存在（默认）
cloud-cli upload ./movie.mp4 /videos/ --policy rsync       # 仅在修改后上传
```

### 4️⃣ 下载文件 / Download

```bash
# 下载文件 / Download file
cloud-cli download /videos/movie.mp4 ./local_movie.mp4

# 断点续传下载（中断后重新运行即可恢复）/ Resume on interrupt
cloud-cli download /videos/big_movie.mp4 ./big.mp4
```

### 5️⃣ 同步目录 / Sync (✨ v1.4.0 新增)

```bash
# 本地 → 夸克（仅同步新增/修改的文件）
cloud-cli sync ./photos quark:/backup/photos

# 预览同步内容（不会实际执行）/ Dry Run
cloud-cli sync ./data quark:/backup --dry-run

# 同步并删除目标端多余文件 / Sync + Delete Extra
cloud-cli sync ./important quark:/important --delete
```

### 6️⃣ 更多命令 / More Commands

```bash
# 文件操作 / File Operations
cloud-cli copy <file_id> <dest_id>      # 复制文件
cloud-cli move <file_id> <dest_id>      # 移动文件
cloud-cli rename <file_id> "新名称"       # 重命名
cloud-cli info <file_id>                # 查看文件详情
cloud-cli delete <file_id>              # 删除文件
cloud-cli mkdir <path>                  # 创建目录

# 高级功能 / Advanced Features
cloud-cli search "关键词"                # 搜索文件
cloud-cli share create <id1> <id2>      # 创建分享
cloud-cli share list                    # 查看分享列表
cloud-cli share delete <share_id>       # 删除分享
cloud-cli quota                         # 查看空间使用情况
cloud-cli recycle list                  # 查看回收站
cloud-cli recycle restore <fid>         # 恢复文件
cloud-cli recycle delete <fid>          # 彻底删除

# 账户管理 / Account Management
cloud-cli user                          # 查看当前账户信息
```

---

## 🏗️ 架构设计 / Architecture

### 分层接口设计 / Layered Interface Design

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Commands                         │
│   login  ls  upload  download  sync  copy  move  search ... │
├─────────────────────────────────────────────────────────────┤
│                    通用引擎 / Core Engines                   │
│  ┌─────────────┐  ┌──────────────┐  ┌────────────────────┐ │
│  │ Sync Engine │  │  Transfer    │  │ Config Manager     │ │
│  │  (同步引擎)  │  │  (传输引擎)   │  │  (配置管理器)       │ │
│  └─────────────┘  └──────────────┘  └────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│              通用底座接口 / Base Storage Interface            │
│     Name()  Init()  List()  Open()  Put()  Delete() ...     │
├─────────────────────────────────────────────────────────────┤
│              扩展能力接口 / Extension Interfaces              │
│  ┌──────────┐ ┌──────────┐ ┌───────────┐ ┌───────────────┐ │
│  │ Sharable │ │Searchable│ │QuotaProv. │ │  RecycleBin   │ │
│  │ (分享)    │ │ (搜索)    │ │ (空间)     │ │  (回收站)      │ │
│  └──────────┘ └──────────┘ └───────────┘ └───────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                  驱动实现 / Driver Backends                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │ Quark Drive │  │ Local FS    │  │ Baidu / 115 / Ali... │ │
│  │ (夸克网盘)   │  │ (本地文件系统) │  │ (未来扩展)            │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

### 目录结构 / Directory Structure

```
cloud-cli/
├── core/                  # 微内核核心
│   ├── fs.go              #   基础接口 + 扩展接口定义
│   ├── config.go          #   YAML 多账户配置管理
│   ├── sync.go            #   目录同步引擎
│   ├── transfer.go        #   跨盘流式传输引擎
│   └── registry.go        #   驱动注册表
├── backends/              # 网盘驱动
│   ├── quark/             #   夸克网盘（完整实现）
│   └── local/             #   本地文件系统
├── cmd/                   # CLI 命令（Cobra）
├── utils/                 # 工具库
│   ├── pool.go            #   并发工作池（带熔断）
│   └── hash.go            #   流式 SHA1 计算
├── .github/workflows/     # CI/CD
│   └── release.yml        #   Goreleaser 自动构建
└── main.go                # 入口（导入所有驱动）
```

---

## 📋 完整功能列表 / Full Feature List

### 基础操作 / Basic Operations
- ✅ `login` — 扫码登录 / Cookie 登录
- ✅ `ls` — 列出目录
- ✅ `upload` — 上传文件（分片并发 + 断点续传 + 秒传）
- ✅ `download` — 下载文件（支持断点续传）
- ✅ `delete` — 删除文件/文件夹
- ✅ `mkdir` — 创建目录

### 文件管理 / File Management
- ✅ `copy` — 复制文件
- ✅ `move` — 移动文件
- ✅ `rename` — 重命名
- ✅ `info` — 查看文件详情（大小、类型、Hash、时间）

### 高级功能 / Advanced Features
- ✅ `search` — 网盘内文件搜索
- ✅ `share` — 分享管理（创建/列表/删除）
- ✅ `quota` — 空间使用情况（含可视化进度条）
- ✅ `recycle` — 回收站管理（列表/恢复/彻底删除）

### 同步与传输 / Sync & Transfer (v1.4.0+)
- ✅ `sync` — 目录同步（支持 `--dry-run` 和 `--delete`）
- ✅ 本地 ↔ 云端同步
- ✅ 跨盘流式传输引擎（预留）

### 工程化 / Engineering
- ✅ YAML 多账户配置
- ✅ Goreleaser 跨平台构建（Linux/macOS/Windows × amd64/arm64）
- ✅ GitHub Actions CI/CD
- ✅ SSRF 安全防护

---

## 🔧 高级用法 / Advanced Usage

### 上传策略 / Upload Policy

```bash
# 跳过已存在文件（默认）/ Skip existing
cloud-cli upload ./file.txt /dest/ --policy skip

# 覆盖已存在文件 / Overwrite existing
cloud-cli upload ./file.txt /dest/ --policy overwrite

# 仅在文件有变化时上传 / Upload only if modified
cloud-cli upload ./file.txt /dest/ --policy rsync
```

### JSON 输出 / JSON Output

所有列表和查询命令都支持 `--output json` (或 `-o json`) 参数，方便与其他脚本或工具集成：

```bash
# 以 JSON 格式列出文件 / List files as JSON
cloud-cli ls -o json

# 输出示例：
# [
#   {
#     "id": "file_id_1",
#     "name": "document.pdf",
#     "size": 1048576,
#     "is_dir": false,
#     "mod_time": "2026-01-01T12:00:00Z",
#     "path": "/document.pdf"
#   }
# ]

# 结合 jq 使用 / Pipe to jq
cloud-cli ls -o json | jq '.[].name'

# 查看空间使用情况 / Check quota as JSON
cloud-cli quota -o json
```

### 同步预览 / Sync Dry Run

在执行实际同步之前，可以先预览会做什么操作：

```bash
cloud-cli sync ./photos quark:/backup/photos --dry-run
```

输出示例：
```
🔄 Syncing ./photos -> quark:/backup/photos ...
👀 This is a dry run, no files will be modified.
⬆️ upload IMG_001.jpg [DRY RUN]
⬆️ upload IMG_002.jpg [DRY RUN]
🔄 update edited_photo.png [DRY RUN]
✅ Sync completed.
```

### 配置管理 / Configuration

配置文件位于 `~/.config/cloud-cli/config.yaml`：

```yaml
default: quark-default
accounts:
  quark-default:
    type: quark
    cookie:
      cookie: "your_cookie_here"
settings:
  speed_limit: "0"
  retries: 3
  timeout: 60
```

---

## 🌍 支持的平台 / Supported Platforms

| 平台 / OS | amd64 | arm64 |
|:---|:---:|:---:|
| **Linux** | ✅ | ✅ |
| **macOS** | ✅ | ✅ |
| **Windows** | ✅ | ✅ |

---

## 🔌 如何添加新网盘驱动？ / How to Add a New Driver?

1. 创建驱动目录：`backends/<provider>/driver.go` + `api.go`
2. 实现 `core.Storage` 接口（至少实现 CRUD 方法）
3. 可选：实现扩展接口（`Sharable`、`Searchable` 等）
4. 在 `init()` 中注册：`core.Register("provider", NewDriver)`
5. 在 `main.go` 中添加空白导入：`_ "github.com/Ab-code520/cloud-cli/backends/<provider>"`

**就是这么简单！** 驱动接入后即可享受：同步引擎、传输引擎、CLI 命令等全部基础设施。

---

## 📝 License

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## 🤝 贡献 / Contributing

欢迎提交 Issue 和 Pull Request！

Issues and Pull Requests are welcome!

---

<div align="center">

**Made with ❤️ by [Abu](https://github.com/Ab-code520)**

[⭐ Star this repo](https://github.com/Ab-code520/cloud-cli) · [📖 Report Bug](https://github.com/Ab-code520/cloud-cli/issues) · [💡 Request Feature](https://github.com/Ab-code520/cloud-cli/issues)

</div>

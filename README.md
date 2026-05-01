# ☁️ cloud-cli

**High-Performance, Microkernel-Architecture CLI for Cloud Storage.**

Supports multiple cloud providers (Quark, Baidu, 115, AliDrive...) with a unified interface. Built with Go, designed for automation, concurrency, and extensibility.

## ✨ Features (v1.0)

- 🚀 **Microkernel Architecture**: `core/fs.go` interface + `backends` plugin pattern. Add new providers by just implementing an interface.
- 🔥 **Quark Drive Support**: Full support for Upload (Multipart), Download, List, Delete.
- ⚡ **Concurrent Engine**: `utils/pool.go` manages worker pools with **Context Cancellation** (Ctrl+C instantly stops all requests).
- 🔒 **SSRF Protection**: Validates all upload/download URLs against private/reserved IP ranges.
- 💾 **Resumable Uploads**: State machine persists to `~/.cache/cloud-cli/upload_states/`. Crash? Just run upload again to resume.
- ⚖️ **Instant Upload (秒传)**: Automatic SHA1 calculation and pre-check before upload.
- 📦 **Streaming I/O**: `Open` / `Put` interfaces support `io.Reader` / `io.ReadCloser` for piping.
- 🌍 **Cross-Platform**: Compiles to Linux, macOS, Windows (amd64/arm64).

## 📦 Installation

```bash
# Download binary from Releases
# Or compile from source:
go build -o cloud-cli .
sudo mv cloud-cli /usr/local/bin/
```

## 🚀 Quick Start

```bash
# 1. Login (Save Cookie to config.json)
cloud-cli login quark

# 2. List Files
cloud-cli ls /

# 3. Upload File (Supports Resume & Instant Upload)
cloud-cli upload ./movie.mp4 /video/

# 4. Download File
cloud-cli download /video/movie.mp4 ./local.mp4

# 5. Delete File
cloud-cli delete /video/movie.mp4
```

## 🏗️ Architecture

```
cloud-cli/
├── core/                # Microkernel (Storage Interface, Registry)
├── backends/            # Drivers (quark/, baidu/, ...)
│   └── quark/           # Full implementation: Multipart, Resume, Auth
├── cmd/                 # Cobra Commands (ls, upload, download...)
├── utils/               # Pool (Concurrency), Hash, Security (SSRF)
└── main.go              # Entry Point
```

## 🔌 Adding a New Provider

1. Create `backends/<provider>/driver.go`
2. Implement `core.Storage` interface
3. Register in `init()`: `core.Register("name", NewDriver)`
4. Add blank import in `main.go`

## ⚙️ CI/CD & Releases

- Uses **GoReleaser** for automated cross-compilation.
- GitHub Actions automatically build and release on git tag push.

## 📝 License

MIT

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [v1.4.0] — 2026-01-XX

### 🎉 New Features
- **Universal Architecture**: Layered interface design (`Storage` base + extension interfaces: `Sharable`, `Searchable`, `QuotaProvider`, `RecycleBin`, `OfflineDownloader`)
- **Local Filesystem Driver**: Treat local disk as a cloud backend (`backends/local/`)
- **Sync Engine**: `cloud-cli sync <source> <dest>` for directory synchronization with `--dry-run` and `--delete` support
- **Transfer Engine**: Cross-cloud streaming transfer (`core/transfer.go`)
- **YAML Multi-Account Config**: Unified config management with `core/config.go`
- **Dynamic Capability Detection**: CLI automatically checks if a driver supports a feature before executing

### 🛠️ Improvements
- Migrated config format from JSON to YAML
- Driver registration via `init()` + registry pattern
- Graceful fallback for unsupported features with clear error messages

### 🐛 Bug Fixes
- Fixed interface signature mismatches for share/quota/recycle methods
- Fixed Windows path detection in `parseResourcePath`

---

## [v1.3.0]

### 🎉 New Features
- **Move**: `cloud-cli move <src> <dest>` — Move files/folders
- **Rename**: `cloud-cli rename <id> <new_name>` — Rename files/folders
- **Quota**: `cloud-cli quota` — Visual storage usage display
- **Recycle Bin**: `cloud-cli recycle list/restore/delete` — Manage deleted files
- **Upload Policy**: `--policy skip|overwrite|rsync` — Control upload behavior
- **Batch Operations**: Multi-file support for delete/upload/copy/move

---

## [v1.2.0]

### 🎉 New Features
- **Share Management**: `cloud-cli share create/list/delete`
- **File Search**: `cloud-cli search <query>` with directory filter

---

## [v1.1.0]

### 🎉 New Features
- **File Info**: `cloud-cli info <file_id>` — View detailed metadata
- **Copy**: `cloud-cli copy <src> <dest>` — Copy files across folders

---

## [v1.0.0]

### 🎉 Initial Release
- Quark Drive support with full CRUD operations
- Concurrent multipart upload with resumable state machine
- Instant upload (秒传) via SHA1 pre-check
- SSRF protection for upload/download URLs
- QR code login support
- Goreleaser CI/CD pipeline for cross-platform builds
- Cobra-based CLI with help system

<div align="center">

# rdl

**Fast CLI for downloading files through [Real-Debrid](https://real-debrid.com)**

Unrestrict hoster links and download with optimized multi-connection [aria2c](https://aria2.github.io/) transfers.

[![Go](https://img.shields.io/github/go-mod/go-version/joeyfurness/rdl)](https://go.dev/)
[![License](https://img.shields.io/github/license/joeyfurness/rdl)](LICENSE)
[![CI](https://github.com/joeyfurness/rdl/actions/workflows/ci.yml/badge.svg)](https://github.com/joeyfurness/rdl/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/joeyfurness/rdl)](https://github.com/joeyfurness/rdl/releases)

[Install](#install) · [Quick Start](#quick-start) · [Commands](#commands) · [Configuration](#configuration)

</div>

---

## Features

- **One command to download** — `rdl <links>` handles unrestriction and downloading in one step
- **Multiple input methods** — positional args, file (`-f`), clipboard (`--clip`), or stdin pipe
- **Parallel multi-connection downloads** — powered by aria2c with tuned speed tier parameters
- **Persistent queue** — accumulate links with `rdl queue add`, download later with `rdl queue run`
- **Download history** — track downloads, retry failures with `--retry-failed`
- **Torrent support** — add magnet links, select files, download when ready
- **OAuth device flow** — one-time browser login, automatic token refresh
- **Smart output** — interactive progress in terminal, NDJSON when piped, quiet mode for scripts

## Install

### Prerequisites

- [aria2](https://aria2.github.io/) — `brew install aria2`
- A [Real-Debrid](https://real-debrid.com) premium account

<details open>
<summary><strong>From source</strong></summary>

```bash
# Requires Go 1.22+
git clone https://github.com/joeyfurness/rdl.git
cd rdl
make build
# Optionally: sudo mv rdl /usr/local/bin/
```

</details>

<details>
<summary><strong>From GitHub release</strong></summary>

Download the latest binary from [Releases](https://github.com/joeyfurness/rdl/releases):

```bash
# macOS (Apple Silicon)
curl -Lo rdl.tar.gz https://github.com/joeyfurness/rdl/releases/latest/download/rdl_darwin_arm64.tar.gz
tar xzf rdl.tar.gz
sudo mv rdl /usr/local/bin/

# macOS (Intel)
curl -Lo rdl.tar.gz https://github.com/joeyfurness/rdl/releases/latest/download/rdl_darwin_amd64.tar.gz
tar xzf rdl.tar.gz
sudo mv rdl /usr/local/bin/

# Linux (x86_64)
curl -Lo rdl.tar.gz https://github.com/joeyfurness/rdl/releases/latest/download/rdl_linux_amd64.tar.gz
tar xzf rdl.tar.gz
sudo mv rdl /usr/local/bin/
```

</details>

## Quick Start

```bash
# First run triggers OAuth login automatically
rdl https://rapidgator.net/file/abc123

# Multiple links
rdl https://host.com/file1 https://host.com/file2 https://host.com/file3

# From a file (one link per line, # comments supported)
rdl -f links.txt

# From clipboard
rdl --clip

# Pipe from stdin
cat links.txt | rdl

# Preview without downloading
rdl --dry-run https://host.com/file1

# Download to a specific directory
rdl --to ~/Downloads/MyFolder https://host.com/file1

# Retry any previously failed downloads
rdl --retry-failed
```

## Commands

### Download (default)

| Command | Description |
|---------|-------------|
| `rdl [links...]` | Unrestrict and download links |
| `rdl -f links.txt` | Download links from file |
| `rdl --clip` | Download links from clipboard |
| `rdl --retry-failed` | Retry previously failed downloads |

### Auth

| Command | Description |
|---------|-------------|
| `rdl auth login` | Authenticate with Real-Debrid |
| `rdl auth logout` | Remove stored tokens |
| `rdl auth status` | Show authentication state |

### Queue

| Command | Description |
|---------|-------------|
| `rdl queue add <links...>` | Add links to persistent queue |
| `rdl queue list` | Show queued links |
| `rdl queue clear` | Clear the queue |
| `rdl queue run` | Download all queued links |

### Torrent

| Command | Description |
|---------|-------------|
| `rdl torrent add <magnet>` | Add a magnet link |
| `rdl torrent list` | List your torrents |
| `rdl torrent select <id> [all]` | Select files from a torrent |
| `rdl torrent download <id>` | Download a completed torrent |

### Utility

| Command | Description |
|---------|-------------|
| `rdl history` | Show download history |
| `rdl history --failed` | Show only failed downloads |
| `rdl config get <key>` | Get a config value |
| `rdl config edit` | Open config in `$EDITOR` |
| `rdl config path` | Show config file path |
| `rdl completions <shell>` | Generate shell completions (bash/zsh/fish) |

## Flags

| Flag | Description |
|------|-------------|
| `--to <dir>` | Output directory |
| `-f, --file <path>` | Read links from file |
| `--clip` | Read links from clipboard |
| `--dry-run` | Preview without downloading |
| `--retry-failed` | Retry previously failed downloads |
| `--json` | Force JSON output |
| `-q, --quiet` | Errors only |
| `--fast` | Fast speed tier (2 files, 8 connections each) |
| `--slow` | Standard speed tier (3 files, 4 connections each) |
| `-v, --verbose` | Increase verbosity |
| `--no-color` | Disable colored output |

## Output Modes

rdl auto-detects the best output mode:

| Mode | When | Description |
|------|------|-------------|
| Interactive | Terminal (TTY) | Progress bars and status |
| JSON | Piped / `--json` | Newline-delimited JSON events |
| Quiet | `-q` flag | Errors to stderr only |

```bash
# Pipe to jq for structured output
rdl https://host.com/file | jq 'select(.event == "summary")'
```

## Configuration

Config file: `~/.config/rdl/config.toml`

```toml
[download]
directory = "~/Downloads"
speed_tier = "auto"        # auto | fast | standard
max_retries = 3
retry_delay = "5s"

[output]
mode = "auto"              # auto | interactive | json | quiet
color = true

[behavior]
open_after = false
overwrite = "resume"       # resume | ask | always | never
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `RDL_TOKEN` | Skip OAuth, use this API token directly |
| `RDL_OUTPUT_DIR` | Override download directory |
| `RDL_MODE` | Override output mode (json/quiet/interactive) |

## Speed Tiers

| Setting | Fast (500Mbps+) | Standard (100-500Mbps) |
|---------|-----------------|------------------------|
| Concurrent files | 2 | 3 |
| Connections per file | 8 | 4 |
| Piece size | 8M | 4M |

Auto-detected by default. Override with `--fast` or `--slow`.

## Shell Completions

```bash
# Zsh
rdl completions zsh > ~/.zsh/completions/_rdl

# Bash
rdl completions bash > /usr/local/etc/bash_completion.d/rdl

# Fish
rdl completions fish > ~/.config/fish/completions/rdl.fish
```

## Contributing

Found a bug or have an idea? [Open an issue](https://github.com/joeyfurness/rdl/issues).

## License

[MIT](LICENSE)

# rdl

A fast command-line tool for downloading files through [Real-Debrid](https://real-debrid.com). Paste your hoster links, and `rdl` unrestricts them via the Real-Debrid API and downloads with optimized [aria2c](https://aria2.github.io/) multi-connection transfers.

## Features

- **One command to download** — `rdl <links>` handles unrestriction and downloading in one step
- **Multiple input methods** — positional args, file (`-f`), clipboard (`--clip`), or stdin pipe
- **Parallel multi-connection downloads** — powered by aria2c with tuned parameters
- **Persistent queue** — accumulate links over time with `rdl queue add`, download later with `rdl queue run`
- **Download history** — track completed and failed downloads, retry failures with `--retry-failed`
- **Torrent support** — add magnet links, select files, download when ready
- **OAuth device flow** — one-time setup, automatic token refresh
- **Smart output** — interactive progress in terminal, NDJSON when piped, quiet mode for scripts

## Install

### Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [aria2](https://aria2.github.io/) — `brew install aria2`
- A [Real-Debrid](https://real-debrid.com) premium account

### Build from source

```bash
git clone https://github.com/joeyfurness/rdl.git
cd rdl
go build -o rdl .
```

Optionally move the binary to your PATH:

```bash
mv rdl /usr/local/bin/
```

## Quick Start

```bash
# First run triggers OAuth login
rdl https://rapidgator.net/file/abc123

# Multiple links at once
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
```

## Commands

```
rdl [links...]                    Download links (default command)
rdl auth login                    Authenticate with Real-Debrid
rdl auth logout                   Remove stored tokens
rdl auth status                   Show authentication state
rdl queue add <links...>          Add links to persistent queue
rdl queue list                    Show queued links
rdl queue clear                   Clear the queue
rdl queue run                     Download all queued links
rdl torrent add <magnet>          Add a magnet link
rdl torrent list                  List your torrents
rdl torrent select <id> [all]     Select files from a torrent
rdl torrent download <id>         Download a completed torrent
rdl history                       Show download history
rdl history --failed              Show only failed downloads
rdl config get <key>              Get a config value
rdl config edit                   Open config in $EDITOR
rdl config path                   Print config file location
rdl completions <bash|zsh|fish>   Generate shell completions
```

## Flags

| Flag | Description |
|------|-------------|
| `--to <dir>` | Output directory |
| `-f, --file <path>` | Read links from file |
| `--clip` | Read links from clipboard |
| `--dry-run` | Show what would download without downloading |
| `--retry-failed` | Retry previously failed downloads |
| `--json` | Force JSON output |
| `-q, --quiet` | Suppress non-essential output |
| `--fast` | Use fast speed tier (fewer concurrent files, more connections each) |
| `--slow` | Use standard speed tier |
| `-v, --verbose` | Increase verbosity |
| `--no-color` | Disable colored output |

## Output Modes

`rdl` auto-detects the best output mode:

- **Interactive** (terminal) — progress bars and status updates
- **JSON** (piped/redirected) — newline-delimited JSON events, ideal for `jq`
- **Quiet** (`-q`) — errors only, for scripts that just need the exit code

```bash
# JSON events when piped
rdl https://host.com/file | jq '.event'

# Force JSON in terminal
rdl --json https://host.com/file
```

## Configuration

Config lives at `~/.config/rdl/config.toml`:

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

Override with environment variables:

```bash
RDL_TOKEN=...         # Skip OAuth, use this API token
RDL_OUTPUT_DIR=...    # Override download directory
RDL_MODE=json         # Override output mode
```

## Speed Tiers

| Setting | Fast (500Mbps+) | Standard (100-500Mbps) |
|---------|-----------------|------------------------|
| Concurrent files | 2 | 3 |
| Connections per file | 8 | 4 |
| Piece size | 8M | 4M |

Auto-detected by default, override with `--fast` or `--slow`.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All downloads succeeded |
| 1 | Partial or total failure |

## Shell Completions

```bash
# Zsh
rdl completions zsh > ~/.zsh/completions/_rdl

# Bash
rdl completions bash > /usr/local/etc/bash_completion.d/rdl

# Fish
rdl completions fish > ~/.config/fish/completions/rdl.fish
```

## License

MIT

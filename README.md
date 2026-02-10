# Whosthere

[![Go Report Card](https://goreportcard.com/badge/github.com/ramonvermeulen/whosthere)](https://goreportcard.com/report/github.com/ramonvermeulen/whosthere)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ramonvermeulen/whosthere)](https://go.dev/doc/devel/release)
[![License](https://img.shields.io/github/license/ramonvermeulen/whosthere)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/ramonvermeulen/whosthere)](https://github.com/ramonvermeulen/whosthere/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/ramonvermeulen/whosthere)](https://github.com/ramonvermeulen/whosthere)

Local Area Network discovery tool with an interactive Terminal User Interface (TUI) written in Go.
Discover, explore, and understand your LAN in an intuitive way.

Whosthere performs **unprivileged, concurrent scans** using [**mDNS**](https://en.wikipedia.org/wiki/Multicast_DNS)
and [**SSDP**](https://en.wikipedia.org/wiki/Simple_Service_Discovery_Protocol) scanners. Additionally, it sweeps the
local subnet by attempting TCP/UDP connections to trigger ARP resolution, then reads the
[**ARP cache**](https://en.wikipedia.org/wiki/Address_Resolution_Protocol) to identify devices on your Local Area Network.
This technique populates the ARP cache without requiring elevated privileges. All discovered devices are enhanced with
[**OUI**](https://standards-oui.ieee.org/) lookups to display manufacturers when available.

Whosthere provides a friendly, intuitive way to answer the question every network administrator asks: "Who's there on my network?"

![demo gif](.github/assets/demo.gif)

## Features

- **Interactive TUI:** Navigate and explore discovered devices intuitively.
- **Fast & Concurrent:** Leverages multiple discovery methods simultaneously.
- **No Elevated Privileges Required:** Runs entirely in user-space.
- **Device Enrichment:** Uses [**OUI**](https://standards-oui.ieee.org/) lookup to show device manufacturers.
- **Integrated Port Scanner:** Optional service discovery on found hosts (only scan devices with permission!).
- **Daemon Mode with HTTP API:** Run in the background and integrate with other tools.
- **Theming & Configuration:** Personalize the look and behavior via YAML configuration.

## Installation

Via [**Homebrew**](https://brew.sh/) with `brew`:

```bash
brew install whosthere
```

On [**NixOS**](https://nixos.org/) with `nix`:

```bash
nix profile install nixpkgs#whosthere
```

On [**Arch Linux**](https://archlinux.org/) with `yay`:

```bash
yay -S whosthere-bin
```

If your package manager is not listed you can always install with [**Go**](https://go.dev/):

```bash
go install github.com/ramonvermeulen/whosthere@latest
```

Or build from source:

```bash
git clone https://github.com/ramonvermeulen/whosthere.git
cd whosthere
make build
```

Additionally, you can download pre-built binaries from the
[**releases page**](https://github.com/ramonvermeulen/whosthere/releases).

## Usage

Run the TUI for interactive discovery:

```bash
whosthere
```

Run as cli to do a single scan and output results:

```bash
whosthere scan -t 5
```

Run as a daemon with HTTP API:

```bash
whosthere daemon --port 8080
```

Additional command line options can be found by running:

```bash
whosthere --help
```

## Platforms

Whosthere is supported on the following platforms:

- [x] Linux
- [x] macOS
- [x] Windows

## Key bindings (TUI)

| Key                | Action                      |
| ------------------ | --------------------------- |
| `/`                | Start regex search          |
| `k`                | Up                          |
| `j`                | Down                        |
| `g`                | Go to top                   |
| `G`                | Go to bottom                |
| `y`                | Copy IP of selected device  |
| `Y`                | Copy MAC of selected device |
| `enter`            | Show device details         |
| `CTRL+t`           | Toggle theme selector       |
| `CTRL+c`           | Stop application            |
| `ESC`              | Clear search / Go back      |
| `p` (details view) | Start port scan on device   |
| `tab` (modal view) | Switch button selection     |

## Environment Variables

### General Environment Variables

| Variable           | Description                                                                     |
| ------------------ | ------------------------------------------------------------------------------- |
| `WHOSTHERE_CONFIG` | Path to the configuration file, to be able to overwrite the default location.   |
| `WHOSTHERE_LOG`    | Set the log level (e.g., `debug`, `info`, `warn`, `error`). Defaults to `info`. |
| `NO_COLOR`         | Disable ANSI colors in the TUI.                                                 |

### Configuration via Environment Variables

Any configuration option that is available in the YAML configuration, can be set via environment variables using the
`WHOSTHERE__` prefix (note the double underscore). Nested configuration keys are separated by double underscores, and
keys are case-insensitive.

Examples:

- `WHOSTHERE__SPLASH__ENABLED=false` - Disable the splash screen, equivalent to `splash.enabled: false` in the YAML config
- `WHOSTHERE__SPLASH__DELAY=2s` - Set splash screen delay to 2 seconds, equivalent to `splash.delay: 2s` in the YAML config
- `WHOSTHERE__SCAN_INTERVAL=30s` - Set scan interval to 30 seconds, equivalent to `scan_interval: 30s` in the YAML config
- `WHOSTHERE__SCANNERS__MDNS__ENABLED=false` - Disable mDNS scanner, equivalent to `scanners.mdns.enabled: false` in the YAML config
- `WHOSTHERE__PORT_SCANNER__TCP=80,443,8080` - Set custom TCP ports to scan, equivalent to `port_scanner.tcp: [80, 443, 8080]` in the YAML config
- `WHOSTHERE__THEME__NAME=cyberpunk` - Set theme to cyberpunk, equivalent to `theme.name: cyberpunk` in the YAML config

## Configuration

Whosthere supports multiple configuration methods with the following precedence (highest to lowest):

1. **Command line flags** - Highest priority
1. **Environment variables** - Using `WHOSTHERE__` prefix
1. **Configuration file** - YAML format
1. **Default values** - [**Built-in defaults**](https://github.com/ramonvermeulen/whosthere/blob/main/internal/core/config/config.go)

### Configuration File

By default, Whosthere looks for configuration in this order:

1. Path specified in `WHOSTHERE_CONFIG` environment variable (if set)
1. `$XDG_CONFIG_HOME/whosthere/config.yaml` (if `XDG_CONFIG_HOME` is set)
1. `~/.config/whosthere/config.yaml` (fallback)

**Example configuration:**

```yaml
# Uncomment the next line to configure a specific network interface - uses OS default if not set
# network_interface: eth0

# How often to run discovery scans
scan_interval: 20s

# Maximum timeout for each scan
scan_timeout: 10s

scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: true
  arp:
    enabled: true

sweeper:
  enabled: true
  interval: 5m
  timeout: 20s

port_scanner:
  timeout: 5s
  # List of TCP ports to scan on discovered devices
  tcp: [21, 22, 23, 25, 80, 110, 135, 139, 143, 389, 443, 445, 993, 995, 1433, 1521, 3306, 3389, 5432, 5900, 8080, 8443, 9000, 9090, 9200, 9300, 10000, 27017]

splash:
  enabled: true
  delay: 1s

theme:
  # When disabled, the TUI will use the terminal it's default ANSI colors
  # Also see the NO_COLOR environment variable to completely disable ANSI colors
  enabled: true
  # See the complete list of available themes at https://github.com/ramonvermeulen/whosthere/tree/main/internal/ui/theme/theme.go
  # Set name to "custom" to use the custom colors below
  # For any color that is not configured it will take the default theme value as fallback
  name: default

  # Custom theme colors (uncomment and set name: custom to use)
  # primitive_background_color: "#000a1a"
  # contrast_background_color: "#001a33"
  # more_contrast_background_color: "#003366"
  # border_color: "#0088ff"
  # title_color: "#00ffff"
  # graphics_color: "#00ffaa"
  # primary_text_color: "#cceeff"
  # secondary_text_color: "#6699ff"
  # tertiary_text_color: "#ffaa00"
  # inverse_text_color: "#000a1a"
  # contrast_secondary_text_color: "#88ddff"
```

## Daemon mode HTTP API

When running Whosthere in daemon mode, it exposes an very simplistic HTTP API with the following endpoints:

| Method | Endpoint       | Description                        |
| ------ | -------------- | ---------------------------------- |
| GET    | `/devices`     | Get list of all discovered devices |
| GET    | `/device/{ip}` | Get details of a specific device   |
| GET    | `/health`      | Health check                       |

## Themes

Theme can be configured via the configuration file, or at runtime via the `CTRL+t` key binding.
A complete list of available themes can be found [**here**](https://github.com/ramonvermeulen/whosthere/blob/main/internal/ui/theme/theme.go), feel free to open a PR to add your own theme!

Example of theme configuration:

```yaml
theme:
  enabled: true
  name: cyberpunk
```

When the `name` is set to `custom`, the other color options can be used to create your own custom theme.
When the `enabled` option is set to `false`, the TUI will use the terminal's default ANSI colors.
When `NO_COLOR` environment variable is set, all ANSI colors will be disabled.

## Logging

Logs are written to the application's state directory:

- `$XDG_STATE_HOME/whosthere/app.log` (if `XDG_STATE_HOME` is set)
- `~/.local/state/whosthere/app.log` (fallback Linux/MacOS)
- `%LOCALAPPDATA%/whosthere/app.log` (fallback Windows)

When not running in TUI mode, logs are also written to the console.

## Known Issues

For clipboard functionality to work, a [**fork of go-clipboard**](https://github.com/dece2183/go-clipboard) is used.
Ensure you have the appropriate copy tool installed for your OS:

| OS                                     | Supported copy tools                         |
| -------------------------------------- | -------------------------------------------- |
| Darwin                                 | `pbcopy`                                     |
| Windows                                | `clip.exe`                                   |
| Linux/FreeBSD/NetBSD/OpenBSD/Dragonfly | X11: `xsel`, `xclip` <br> Wayland: `wl-copy` |

## Disclaimer

Whosthere is intended for use on networks where you have permission to perform network discovery and scanning,
such as your own home network. Unauthorized scanning of networks may be illegal and unethical.
Always obtain proper authorization before using this tool on any network.

This tool was created primarily for educational purposes and home network exploration. While functional
and useful, it should not be considered a professional-grade network monitoring solution. Results are
based on unprivileged scanning techniques and may not be comprehensive. For critical network analysis
or security assessments, consider using established professional tools with formal support.

## Contributing

Contributions and suggestions such as feature requests, bug reports, or improvements are welcome!
Feel free to open issues or submit pull requests on the GitHub repository.
Please make sure to discuss any major changes on a Github issue before implementing them.

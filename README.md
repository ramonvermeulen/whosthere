# Whosthere

[![Go Report Card](https://goreportcard.com/badge/github.com/ramonvermeulen/whosthere)](https://goreportcard.com/report/github.com/ramonvermeulen/whosthere)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ramonvermeulen/whosthere)](https://go.dev/doc/devel/release)
[![License](https://img.shields.io/github/license/ramonvermeulen/whosthere)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/ramonvermeulen/whosthere)](https://github.com/ramonvermeulen/whosthere/releases)
[![GitHub Repo stars](https://img.shields.io/github/stars/ramonvermeulen/whosthere)](https://github.com/ramonvermeulen/whosthere)

Local area network discovery tool with a modern Terminal User Interface (TUI) written in Go.
Discover, explore, and understand your LAN in an intuitive way.

Whosthere performs **privilege-less, concurrent scans** using [**mDNS**](https://en.wikipedia.org/wiki/Multicast_DNS) and
[**SSDP**](https://en.wikipedia.org/wiki/Simple_Service_Discovery_Protocol) scanners. Aside from that it sweeps
the local subnet with TCP/UDP connections and reads out the [**ARP Cache**](https://en.wikipedia.org/wiki/ARP_cache) to
quickly find and identify devices on your Local Area Network. This is a technique that under-the-hood triggers ARP
requests to all IP addresses in the network interface it's subnet, without requiring elevated privileges.
All discovered devices are enhanced with [**OUI**](https://standards-oui.ieee.org/) lookup to show manufacturers if available.

![demo gif](.github/assets/demo.gif)

## Features

- **Modern TUI:** Navigate and explore discovered devices intuitively.
- **Fast & Concurrent:** Leverages multiple discovery methods simultaneously.
- **No Elevated Privileges Required:** Runs entirely in user-space.
- **Device Enrichment:** Uses [**OUI**](https://standards-oui.ieee.org/) lookup to show device manufacturers.
- **Integrated Port Scanner:** Optional service discovery on found hosts (only scan devices with permission!).
- **Daemon Mode with HTTP API:** Run in the background and integrate with other tools.
- **Theming & Configuration:** Personalize the look and behavior via YAML configuration.

## Installation

```bash
brew tap ramonvermeulen/whosthere
brew install whosthere
```

Or with Go:

```bash
go install github.com/ramonvermeulen/whosthere@latest
```

Or build from source:

```bash
git clone github.com/ramonvermeulen/whosthere.git
cd whosthere
make build
```

## Usage

Run the TUI for interactive discovery:

```bash
whosthere
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
- [ ] Windows (maybe in the future, contributions welcome!)

## Key bindings (TUI)

| Key                | Action                     |
| ------------------ | -------------------------- |
| `/`                | Start regex search         |
| `k`                | Up                         |
| `j`                | Down                       |
| `g`                | Go to top                  |
| `G`                | Go to bottom               |
| `y`                | Copy IP of selected device |
| `enter`            | Show device details        |
| `CTRL+t`           | Toggle theme selector      |
| `CTRL+c`           | Stop application           |
| `ESC`              | Clear search / Go back     |
| `p` (details view) | Start port scan on device  |
| `tab` (modal view) | Switch button selection    |

## Environment Variables

| Variable           | Description                                                                     |
| ------------------ | ------------------------------------------------------------------------------- |
| `WHOSTHERE_CONFIG` | Path to the configuration file, to be able to overwrite the default location.   |
| `WHOSTHERE_LOG`    | Set the log level (e.g., `debug`, `info`, `warn`, `error`). Defaults to `info`. |

## Configuration

Whosthere can be configured via a YAML configuration file.
By default, it looks for the configuration file in the following order:

- Path specified in the `WHOSTHERE_CONFIG` environment variable (if set)
- `$XDG_CONFIG_HOME/whosthere/whosthere.yaml` (if `XDG_CONFIG_HOME` is set)
- `~/.config/whosthere/whosthere.yaml` (otherwise)

When not running in TUI mode, logs are also written to the console output.

Example of the default configuration file:

```yaml
# How often to run discovery scans
scan_interval: 20s

# Maximum duration for each scan
# If you set this too low some scanners or the sweeper might not complete in time
scan_duration: 10s

# Splash screen configuration
splash:
  enabled: true
  delay: 1s

# Theme configuration
theme:
  # Configure the theme to use for the TUI, complete list of available themes at:
  # https://github.com/ramonvermeulen/whosthere/tree/main/internal/ui/theme/theme.go
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

# Scanner configuration
scanners:
  mdns:
    enabled: true
  ssdp:
    enabled: true
  arp:
    enabled: true

# Port scanner configuration
port_scanner:
  timeout: 5s
  # List of TCP ports to scan on discovered devices
  tcp: [21, 22, 23, 25, 80, 110, 135, 139, 143, 389, 443, 445, 993, 995, 1433, 1521, 3306, 3389, 5432, 5900, 8080, 8443, 9000, 9090, 9200, 9300, 10000, 27017]

# Uncomment the next line to configure a specific network interface - uses OS default if not set
# network_interface: lo0
```

## Daemon mode HTTP API

When running Whosthere in daemon mode, it exposes an very simplistic HTTP API with the following endpoints:

| Method | Endpoint       | Description                        |
| ------ | -------------- | ---------------------------------- |
| GET    | `/devices`     | Get list of all discovered devices |
| GET    | `/device/{ip}` | Get details of a specific device   |
| GET    | `/healtz`      | Health check                       |

## Themes

Theme can be configured via the configuration file, or at runtime via the `CTRL+t` key binding.
A complete list of available themes can be found [**here**](<>), feel free to open a PR to add your own theme!

Example of theme configuration:

```yaml
theme:
  name: cyberpunk
```

When the `name` is set to `custom`, the other color options can be used to create your own custom theme.

## Logging

Logs are written to the application's state directory:

- `$XDG_STATE_HOME/whosthere/app.log` (if `XDG_STATE_HOME` is set)
- `~/.local/state/whosthere/app.log` (otherwise)

When not running in TUI mode, logs are also output to the console.

## Disclaimer

Whosthere is intended for use on networks where you have permission to perform network discovery and scanning,
such as your own home network. Unauthorized scanning of networks may be illegal and unethical.
Always obtain proper authorization before using this tool on any network.

## Contributing

Contributions and suggestions such as feature requests, bug reports, or improvements are welcome!
Feel free to open issues or submit pull requests on the GitHub repository.
Please make sure to discuss any major changes on a Github issue before implementing them.

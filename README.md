# whosthere

Whosthere is a TUI application that discovers devices and services on your local network, built in Go for fast, 
concurrent scanning and a clean terminal interface. I'm building this primarily for myself to deepen my understanding
of Golang and networking fundamentals. Feel free to use it, contribute to it, or suggest features! I'm open to all kinds
of suggestions and feedback.

## Configuration
A lot of behavior within whosthere can be configured to your liking. By default, whosthere will try to look for a configuration
file at `$XDG_CONFIG_HOME/whosthere/config.yaml`, or `~/.config/whosthere/config.yaml` if the [**XDG Base Directory**](https://specifications.freedesktop.org/basedir/latest/#basics)
environment variables are not set. If no configuration file is found, whosthere will create one with default values on first run.
The location of the configuration file can be overridden by providing the `--config` (`-c`) flag when starting whosthere, 
or the `WHOSTHERE_CONFIG` environment variable.

Here is an example configuration file with all available options and their default values:

```yaml
splash:
  enabled: true # Show splash screen on startup
  delay: 1 # Delay in seconds for the splash screen
```

## Logging
Whosthere supports logging to a file for debugging and monitoring purposes. By default, logs are written to
`$XDG_STATE_HOME/whosthere/whosthere.log`, or `~/.local/state/whosthere/whosthere.log` if the [**XDG Base Directory**](https://specifications.freedesktop.org/basedir/latest/#basics)
environment variables are not set. The log level is set to `info` by default, but can be changed via the `WHOSTHERE_LOG`
environment variable. 

For example, to set the log level to `debug`, you can start whosthere with the following command:

```bash
WHOSTHERE_LOG=debug whosthere
```


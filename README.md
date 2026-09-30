# netscope

**Live network mapper in Go.** Single binary, no dependencies.

Sweeps a subnet, finds live hosts, fingerprints open ports, and renders the results in a live terminal dashboard — or exports to JSON/HTML.

> by [cortex](https://github.com/zasmastro)

## Install

    go install github.com/zasmastro/netscope@latest

Or download a prebuilt binary from the [Releases](../../releases) page.

## Usage

    netscope tui <cidr> [-o report.(json|html)]   live dashboard + optional export
    netscope sweep <cidr> [--json]                find live hosts (text or JSON)
    netscope scan <host> [--json]                 scan a host's common ports
    netscope version                              print version
    netscope help                                 this message

## Examples

Live dashboard:

    netscope tui 192.168.1.0/24

Dashboard with JSON export:

    netscope tui 192.168.1.0/24 -o report.json

Dashboard with HTML report:

    netscope tui 192.168.1.0/24 -o report.html

Quick sweep, text output:

    netscope sweep 192.168.1.0/24

    [*] sweeping 254 hosts on 192.168.1.0/24

    [+] 192.168.1.1
    [+] 192.168.1.8

    [*] 2 live in 4.2s

JSON output for scripts:

    netscope scan 192.168.1.1 --json

    {
      "tool": {"name": "netscope", "version": "0.3.2", ...},
      "scanned_at": "2026-09-30T17:01:41Z",
      "target": "192.168.1.1",
      "hosts": [
        {
          "ip": "192.168.1.1",
          "status": "up",
          "os_guess": "network device (router / switch)",
          "ports": [
            {"port": 53, "service": "dns"},
            {"port": 80, "service": "http"}
          ]
        }
      ]
    }

## Features

- **Live dashboard** — hosts and ports appear in real time as they're found
- **Host detail view** — press Enter on any host for full port + banner info
- **OS guess** — ports + banners scored against Windows / Linux / macOS / network-device signatures
- **Concurrent scanning** — 128 workers sweep a /24 in seconds
- **Service labels** — ports show as `ssh`, `http`, `smb`, not raw numbers
- **JSON export** — machine-readable output for pipelines
- **HTML report** — standalone styled report, opens in any browser
- **Text mode** — `sweep` and `scan` for scripting, no UI
- **Zero dependencies** — one Go binary, no runtime, no installs

## Build from source

    git clone https://github.com/zasmastro/netscope
    cd netscope
    go build -o netscope.exe .

Requires Go 1.21+.

## Status

v0.3.2 — dashboard, detail view, OS guess, JSON + HTML export. History, diff, config file coming.

## Legal

Use only on networks you own or have written permission to test.

## License

MIT
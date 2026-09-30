# netscope

Live network mapper in Go. Single binary, no dependencies.

> by [cortex](https://github.com/zasmastro)

## Install

    go install github.com/zasmastro/netscope@latest

Or build from source:

    git clone https://github.com/zasmastro/netscope
    cd netscope
    go build -o netscope.exe .

## Usage

    netscope sweep <cidr>    find live hosts on a subnet
    netscope scan <host>     scan a single host's common ports
    netscope version         print version
    netscope help            this message

## Examples

    netscope sweep 192.168.1.0/24
    netscope scan 192.168.1.1

## Status

v0.1.0 — working scanner. Live TUI dashboard coming.

## License

MIT
package main

import (
	"fmt"
	"os"
	"time"
)

const (
	name    = "netscope"
	version = "0.2.0"
	author  = "cortex"
	repo    = "https://github.com/zasmastro/netscope"
)

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	switch os.Args[1] {
	case "sweep":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope sweep <cidr>")
			os.Exit(1)
		}
		runSweep(os.Args[2])

	case "scan":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope scan <host>")
			os.Exit(1)
		}
		runScan(os.Args[2])

	case "tui":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope tui <cidr>")
			os.Exit(1)
		}
		runTUI(os.Args[2])

	case "version", "-v", "--version":
		fmt.Printf("%s %s\nby %s — %s\n", name, version, author, repo)

	case "help", "-h", "--help":
		help()

	default:
		fmt.Printf("unknown command: %s\n\n", os.Args[1])
		help()
		os.Exit(1)
	}
}

func help() {
	fmt.Printf(`%s %s — live network mapper
by %s · %s

Usage:
  %s sweep <cidr>    find live hosts on a subnet (text output)
  %s scan <host>     scan a single host's common ports (text output)
  %s tui <cidr>      live dashboard — sweeps and scans with a UI
  %s version         print version
  %s help            this message

Examples:
  %s sweep 192.168.1.0/24
  %s scan 192.168.1.1
  %s tui 192.168.1.0/24
`, name, version, author, repo, name, name, name, name, name, name, name, name)
}

func runSweep(cidr string) {
	hosts, err := expandCIDR(cidr)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s — by %s\n", name, version, author)
	fmt.Printf("[*] sweeping %d hosts on %s\n\n", len(hosts), cidr)
	start := time.Now()

	results := sweep(hosts)

	elapsed := time.Since(start)
	fmt.Printf("\n[*] %d live in %s\n", len(results), elapsed.Round(time.Millisecond))
	fmt.Printf("    %s by %s · %s\n", name, author, repo)
}

func runScan(host string) {
	fmt.Printf("%s %s — by %s\n", name, version, author)
	fmt.Printf("[*] scanning %s\n\n", host)
	start := time.Now()

	ports := scanHost(host)

	elapsed := time.Since(start)
	fmt.Printf("\n[*] %d open in %s\n", len(ports), elapsed.Round(time.Millisecond))
	fmt.Printf("    %s by %s · %s\n", name, author, repo)
}
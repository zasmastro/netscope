package main

import (
	"fmt"
	"os"
	"time"
)

const (
	name    = "netscope"
	version = "0.3.1"
	author  = "cortex"
	repo    = "https://github.com/zasmastro/netscope"
)

func main() {
	if len(os.Args) < 2 {
		help()
		return
	}

	args := os.Args[2:]
	jsonOut := false
	var outFile string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonOut = true
		case "-o":
			if i+1 < len(args) {
				outFile = args[i+1]
				i++
			}
		}
	}

	switch os.Args[1] {
	case "sweep":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope sweep <cidr> [--json]")
			os.Exit(1)
		}
		runSweep(os.Args[2], jsonOut)

	case "scan":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope scan <host> [--json]")
			os.Exit(1)
		}
		runScan(os.Args[2], jsonOut)

	case "tui":
		if len(os.Args) < 3 {
			fmt.Println("usage: netscope tui <cidr> [-o report.json]")
			os.Exit(1)
		}
		runTUI(os.Args[2], outFile)

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
  %s sweep <cidr> [--json]         find live hosts (text or JSON)
  %s scan <host> [--json]          scan a host's common ports
  %s tui <cidr> [-o report.json]   live dashboard, optional JSON export
  %s version                       print version
  %s help                          this message

Examples:
  %s sweep 192.168.1.0/24
  %s sweep 192.168.1.0/24 --json
  %s scan 192.168.1.1 --json
  %s tui 192.168.1.0/24 -o report.json
`, name, version, author, repo, name, name, name, name, name, name, name, name, name)
}

func runSweep(cidr string, jsonOut bool) {
	hosts, err := expandCIDR(cidr)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if !jsonOut {
		fmt.Printf("%s %s — by %s\n", name, version, author)
		fmt.Printf("[*] sweeping %d hosts on %s\n\n", len(hosts), cidr)
	}

	start := time.Now()
	results := sweep(hosts)
	elapsed := time.Since(start)

	if jsonOut {
		entries := []hostEntry{}
		for _, ip := range results {
			entries = append(entries, hostEntry{ip: ip, ports: []int{}})
		}
		fmt.Println(buildJSON(cidr, entries))
		return
	}

	fmt.Printf("\n[*] %d live in %s\n", len(results), elapsed.Round(time.Millisecond))
	fmt.Printf("    %s by %s · %s\n", name, author, repo)
}

func runScan(host string, jsonOut bool) {
	if !jsonOut {
		fmt.Printf("%s %s — by %s\n", name, version, author)
		fmt.Printf("[*] scanning %s\n\n", host)
	}

	start := time.Now()
	ports := scanHost(host)
	elapsed := time.Since(start)

	if jsonOut {
		h := hostEntry{ip: host, ports: ports, banners: make(map[int]string)}
		fmt.Println(buildJSON(host, []hostEntry{h}))
		return
	}

	fmt.Printf("\n[*] %d open in %s\n", len(ports), elapsed.Round(time.Millisecond))
	fmt.Printf("    %s by %s · %s\n", name, author, repo)
}
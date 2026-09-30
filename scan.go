package main

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

var topPorts = []int{
	21, 22, 23, 25, 53, 80, 110, 135, 139, 143,
	443, 445, 993, 995, 1433, 1521, 3306, 3389,
	5432, 5900, 5985, 6379, 8080, 8443, 9000, 27017,
}

func expandCIDR(cidr string) ([]string, error) {
	ip, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR: %w", err)
	}
	_ = ip

	var hosts []string
	for cur := network.IP.Mask(network.Mask); network.Contains(cur); incIP(cur) {
		if cur.Equal(network.IP) {
			continue
		}
		bcast := make(net.IP, len(cur))
		copy(bcast, cur)
		for i := range bcast {
			bcast[i] |= ^network.Mask[i]
		}
		if cur.Equal(bcast) {
			continue
		}
		hosts = append(hosts, cur.String())
	}
	return hosts, nil
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func sweep(hosts []string) []string {
	const workers = 128
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var live []string

	for _, h := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(host string) {
			defer wg.Done()
			defer func() { <-sem }()
			if isAlive(host) {
				mu.Lock()
				live = append(live, host)
				mu.Unlock()
				fmt.Printf("[+] %s\n", host)
			}
		}(h)
	}
	wg.Wait()

	sort.Slice(live, func(i, j int) bool {
		return ipLess(live[i], live[j])
	})
	return live
}

func isAlive(host string) bool {
	probes := []int{445, 139, 22, 80, 443, 3389, 8080}
	for _, p := range probes {
		if tryConnect(host, p, 300*time.Millisecond) {
			return true
		}
	}
	return false
}

func scanHost(host string) []int {
	var mu sync.Mutex
	var open []int
	var wg sync.WaitGroup

	for _, p := range topPorts {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			if tryConnect(host, port, 800*time.Millisecond) {
				banner := grabBanner(host, port)
				mu.Lock()
				open = append(open, port)
				if banner != "" {
					fmt.Printf("[+] %s:%-5d open  | %s\n", host, port, banner)
				} else {
					fmt.Printf("[+] %s:%-5d open\n", host, port)
				}
				mu.Unlock()
			}
		}(p)
	}
	wg.Wait()

	sort.Ints(open)
	return open
}

func tryConnect(host string, port int, timeout time.Duration) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func grabBanner(host string, port int) string {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 1500*time.Millisecond)
	if err != nil {
		return ""
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(1200 * time.Millisecond))
	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		return ""
	}
	s := cleanBanner(string(buf[:n]))
	if len(s) > 80 {
		s = s[:80] + "..."
	}
	return s
}

func cleanBanner(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\r' || c == '\n' || c == '\t' {
			out = append(out, ' ')
		} else if c >= 32 && c < 127 {
			out = append(out, c)
		}
	}
	result := ""
	prevSpace := false
	for _, c := range string(out) {
		if c == ' ' {
			if !prevSpace {
				result += " "
			}
			prevSpace = true
		} else {
			result += string(c)
			prevSpace = false
		}
	}
	return trim(result)
}

func ipLess(a, b string) bool {
	ia := net.ParseIP(a).To4()
	ib := net.ParseIP(b).To4()
	if ia == nil || ib == nil {
		return a < b
	}
	for i := 0; i < 4; i++ {
		if ia[i] != ib[i] {
			return ia[i] < ib[i]
		}
	}
	return false
}

func trim(s string) string {
	start, end := 0, len(s)
	for start < end && s[start] == ' ' {
		start++
	}
	for end > start && s[end-1] == ' ' {
		end--
	}
	return s[start:end]
}

func sweepStream(hosts []string, onFound func(string)) {
	const workers = 128
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, h := range hosts {
		wg.Add(1)
		sem <- struct{}{}
		go func(host string) {
			defer wg.Done()
			defer func() { <-sem }()
			if isAlive(host) {
				mu.Lock()
				onFound(host)
				mu.Unlock()
			}
		}(h)
	}
	wg.Wait()
}

func scanHostStream(host string, onPort func(port int, banner string)) {
	var wg sync.WaitGroup

	for _, p := range topPorts {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			if tryConnect(host, port, 800*time.Millisecond) {
				banner := grabBanner(host, port)
				onPort(port, banner)
			}
		}(p)
	}
	wg.Wait()
}

func escapeJSON(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 32 {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

func buildJSON(target string, hosts []hostEntry) string {
	var b strings.Builder

	b.WriteString("{\n")
	b.WriteString(fmt.Sprintf("  \"tool\": {\"name\": %q, \"version\": %q, \"author\": %q, \"repo\": %q},\n",
		name, version, author, repo))
	b.WriteString(fmt.Sprintf("  \"scanned_at\": %q,\n", time.Now().UTC().Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("  \"target\": %q,\n", target))
	b.WriteString("  \"hosts\": [\n")

	for i, h := range hosts {
		b.WriteString("    {\n")
		b.WriteString(fmt.Sprintf("      \"ip\": %q,\n", h.ip))
		b.WriteString("      \"status\": \"up\",\n")

		guess := guessOS(&h)
		if guess != "" {
			b.WriteString(fmt.Sprintf("      \"os_guess\": %q,\n", guess))
		}

		b.WriteString("      \"ports\": [")
		if len(h.ports) == 0 {
			b.WriteString("]\n")
		} else {
			b.WriteString("\n")
			for j, p := range h.ports {
				svc := portNames[p]
				banner := h.banners[p]
				b.WriteString("        {")
				b.WriteString(fmt.Sprintf("\"port\": %d", p))
				if svc != "" {
					b.WriteString(fmt.Sprintf(", \"service\": %q", svc))
				}
				if banner != "" {
					b.WriteString(fmt.Sprintf(", \"banner\": %q", escapeJSON(banner)))
				}
				b.WriteString("}")
				if j+1 < len(h.ports) {
					b.WriteString(",")
				}
				b.WriteString("\n")
			}
			b.WriteString("      ]\n")
		}

		b.WriteString("    }")
		if i+1 < len(hosts) {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}

	b.WriteString("  ]\n")
	b.WriteString("}\n")
	return b.String()
}
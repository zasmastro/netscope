package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	styleTitle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	styleCredit   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleHeader   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	styleUp       = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	styleSelected = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226"))
	stylePort     = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleStatus   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleArrow    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	styleBar      = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	styleOS       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213"))
	styleBanner   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
)

var portNames = map[int]string{
	21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp", 53: "dns",
	80: "http", 110: "pop3", 135: "rpc", 139: "netbios", 143: "imap",
	443: "https", 445: "smb", 993: "imaps", 995: "pop3s",
	1433: "mssql", 1521: "oracle", 3306: "mysql", 3389: "rdp",
	5432: "postgres", 5900: "vnc", 5985: "winrm", 6379: "redis",
	8080: "http-alt", 8443: "https-alt", 9000: "http-alt", 27017: "mongo",
}

type hostFoundMsg struct{ host string }
type hostDoneMsg struct{ host string }
type portsFoundMsg struct {
	host   string
	port   int
	banner string
}
type scanDoneMsg struct{}

type hostEntry struct {
	ip        string
	ports     []int
	banners   map[int]string
	done      bool
	firstSeen time.Time
	scanTime  time.Duration
}

type model struct {
	cidr       string
	total      int
	scanned    int
	hosts      []*hostEntry
	byIP       map[string]*hostEntry
	selected   int
	start      time.Time
	elapsed    time.Duration
	done       bool
	width      int
	detailView bool
	detailHost *hostEntry
	outFile    string
}

func newModel(cidr string, total int, outFile string) model {
	return model{
		cidr:    cidr,
		total:   total,
		byIP:    make(map[string]*hostEntry),
		start:   time.Now(),
		outFile: outFile,
	}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		if m.detailView {
			switch msg.String() {
			case "esc", "backspace", "left":
				m.detailView = false
				m.detailHost = nil
			case "q", "ctrl+c":
				m.writeJSONIfRequested()
				return m, tea.Quit
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.writeJSONIfRequested()
			return m, tea.Quit
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.hosts)-1 {
				m.selected++
			}
		case "enter":
			if m.selected < len(m.hosts) {
				m.detailHost = m.hosts[m.selected]
				m.detailView = true
			}
		}

	case hostFoundMsg:
		h := &hostEntry{
			ip:        msg.host,
			banners:   make(map[int]string),
			firstSeen: time.Now(),
		}
		m.hosts = append(m.hosts, h)
		m.byIP[msg.host] = h

	case portsFoundMsg:
		if h, ok := m.byIP[msg.host]; ok {
			h.ports = append(h.ports, msg.port)
			if msg.banner != "" {
				h.banners[msg.port] = msg.banner
			}
		}

	case hostDoneMsg:
		if h, ok := m.byIP[msg.host]; ok {
			h.done = true
			h.scanTime = time.Since(h.firstSeen)
			m.scanned++
		}

	case scanDoneMsg:
		m.done = true
		for _, h := range m.hosts {
			h.done = true
		}
	}

	m.elapsed = time.Since(m.start)
	return m, nil
}

func (m *model) writeJSONIfRequested() {
	if m.outFile == "" {
		return
	}
	entries := make([]hostEntry, 0, len(m.hosts))
	for _, h := range m.hosts {
		entries = append(entries, *h)
	}
	data := buildJSON(m.cidr, entries)
	if err := os.WriteFile(m.outFile, []byte(data), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", m.outFile, err)
	}
}

func (m model) View() string {
	if m.detailView && m.detailHost != nil {
		return m.detailViewRender()
	}
	return m.tableViewRender()
}

func (m model) tableViewRender() string {
	var b strings.Builder

	b.WriteString(styleTitle.Render("netscope") + " " + styleDim.Render(version) +
		"  " + styleCredit.Render("by "+author+" · "+repo) + "\n")

	var progress string
	if m.done {
		progress = styleUp.Render("● done")
	} else {
		pct := 0.0
		if m.total > 0 {
			pct = float64(m.scanned) / float64(m.total)
		}
		filled := int(pct * 20)
		bar := styleBar.Render(strings.Repeat("█", filled)) +
			styleDim.Render(strings.Repeat("░", 20-filled))
		progress = bar + " " + fmt.Sprintf("%3d%%", int(pct*100))
	}
	b.WriteString("  " + styleDim.Render("scanning "+m.cidr) + "  " + progress + "\n\n")

	b.WriteString(fmt.Sprintf("  %-18s %-9s %-24s %s\n",
		styleHeader.Render("HOST"),
		styleHeader.Render("STATUS"),
		styleHeader.Render("PORTS"),
		styleHeader.Render("SERVICES")))
	b.WriteString("  " + styleDim.Render(strings.Repeat("─", 74)) + "\n")

	if len(m.hosts) == 0 {
		b.WriteString(styleDim.Render("  waiting for hosts…") + "\n")
	}

	for i, h := range m.hosts {
		arrow := "  "
		if i == m.selected {
			arrow = styleArrow.Render("> ")
		}

		status := stylePort.Render("◐ scan")
		if h.done {
			status = styleUp.Render("● up")
		}

		portStrs := make([]string, len(h.ports))
		for j, p := range h.ports {
			portStrs[j] = fmt.Sprintf("%d", p)
		}
		ports := strings.Join(portStrs, ", ")
		if len(ports) > 22 {
			ports = ports[:22] + "…"
		}
		if ports == "" {
			if h.done {
				ports = styleDim.Render("—")
			} else {
				ports = styleDim.Render("…")
			}
		}

		svcStrs := []string{}
		for _, p := range h.ports {
			if name, ok := portNames[p]; ok {
				svcStrs = append(svcStrs, name)
			}
		}
		services := strings.Join(svcStrs, ", ")

		row := fmt.Sprintf("%s%-18s %-9s %-24s %s", arrow, h.ip, status, ports, services)
		if i == m.selected {
			row = styleSelected.Render(row)
		}
		b.WriteString(row + "\n")
	}

	b.WriteString("\n")

	portCount := 0
	for _, h := range m.hosts {
		portCount += len(h.ports)
	}
	b.WriteString(styleStatus.Render(fmt.Sprintf("  %d hosts · %d open ports · %.1fs",
		len(m.hosts), portCount, m.elapsed.Seconds())) + "\n")

	hint := "[↑↓] navigate   [enter] details   [q] quit"
	if m.outFile != "" {
		hint = "[↑↓] navigate   [enter] details   [q] quit + save → " + m.outFile
	}
	b.WriteString("  " + styleDim.Render(hint) + "\n")

	return b.String()
}

func (m model) detailViewRender() string {
	h := m.detailHost
	var b strings.Builder

	b.WriteString(styleTitle.Render("netscope") + " " + styleDim.Render(version) +
		"  " + styleCredit.Render("by "+author+" · "+repo) + "\n\n")

	b.WriteString(styleOS.Render("  "+h.ip) + "  " + styleUp.Render("● up") + "\n")

	b.WriteString(styleDim.Render(fmt.Sprintf("  first seen: %s   scanned in: %s",
		h.firstSeen.Format("15:04:05"), h.scanTime.Round(time.Millisecond))) + "\n")

	guess := guessOS(h)
	if guess != "" {
		b.WriteString(styleOS.Render("  likely OS: ") + guess + "\n")
	}
	b.WriteString("\n")

	b.WriteString(fmt.Sprintf("  %-8s %-14s %s\n",
		styleHeader.Render("PORT"),
		styleHeader.Render("SERVICE"),
		styleHeader.Render("BANNER")))
	b.WriteString("  " + styleDim.Render(strings.Repeat("─", 74)) + "\n")

	if len(h.ports) == 0 {
		b.WriteString(styleDim.Render("  no open ports") + "\n")
	}

	for _, p := range h.ports {
		svc := portNames[p]
		if svc == "" {
			svc = "—"
		}
		banner := h.banners[p]
		if banner == "" {
			banner = "—"
		}
		if len(banner) > 44 {
			banner = banner[:44] + "…"
		}
		b.WriteString(fmt.Sprintf("  %-8d %-14s %s\n",
			p,
			stylePort.Render(svc),
			styleBanner.Render(banner)))
	}

	b.WriteString("\n")
	b.WriteString("  " + styleDim.Render("[esc] back   [q] quit") + "\n")

	return b.String()
}

func guessOS(h *hostEntry) string {
	ports := make(map[int]bool)
	for _, p := range h.ports {
		ports[p] = true
	}

	banners := strings.ToLower(strings.Join(bannerValues(h), " "))

	windowsScore := 0
	if ports[135] {
		windowsScore += 2
	}
	if ports[139] {
		windowsScore += 2
	}
	if ports[445] {
		windowsScore += 3
	}
	if ports[5985] {
		windowsScore += 3
	}
	if ports[3389] {
		windowsScore += 2
	}

	linuxScore := 0
	if ports[22] && strings.Contains(banners, "openssh") {
		linuxScore += 4
	}
	if strings.Contains(banners, "ubuntu") {
		linuxScore += 5
	}
	if strings.Contains(banners, "debian") {
		linuxScore += 5
	}
	if strings.Contains(banners, "centos") {
		linuxScore += 5
	}
	if strings.Contains(banners, "nginx") {
		linuxScore += 2
	}
	if strings.Contains(banners, "apache") {
		linuxScore += 1
	}

	macScore := 0
	if strings.Contains(banners, "darwin") {
		macScore += 5
	}
	if strings.Contains(banners, "mac os") {
		macScore += 5
	}

	if ports[53] && !ports[22] {
		return "network device (router / switch)"
	}

	max := func(a, b, c int) (string, int) {
		if a >= b && a >= c {
			return "Windows", a
		}
		if b >= c {
			return "Linux / Unix", b
		}
		return "macOS", c
	}
	label, score := max(windowsScore, linuxScore, macScore)

	if score == 0 {
		return ""
	}
	return label
}

func bannerValues(h *hostEntry) []string {
	out := make([]string, 0, len(h.banners))
	for _, v := range h.banners {
		out = append(out, v)
	}
	return out
}

func runTUI(cidr string, outFile string) {
	hosts, err := expandCIDR(cidr)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	m := newModel(cidr, len(hosts), outFile)
	p := tea.NewProgram(m, tea.WithAltScreen())

	go func() {
		found := make(chan string, len(hosts))

		go func() {
			sweepStream(hosts, func(host string) {
				found <- host
			})
			close(found)
		}()

		var live []string
		for h := range found {
			live = append(live, h)
			p.Send(hostFoundMsg{host: h})
		}

		portCh := make(chan portsFoundMsg, 4096)
		doneCh := make(chan string, len(live))

		for _, h := range live {
			go func(host string) {
				scanHostStream(host, func(port int, banner string) {
					portCh <- portsFoundMsg{host: host, port: port, banner: banner}
				})
				doneCh <- host
			}(h)
		}

		doneCount := 0
		for doneCount < len(live) {
			select {
			case pr := <-portCh:
				p.Send(pr)
			case host := <-doneCh:
				p.Send(hostDoneMsg{host: host})
				doneCount++
			}
		}

		p.Send(scanDoneMsg{})
	}()

	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
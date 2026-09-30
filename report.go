package main

import (
	"fmt"
	"html"
	"sort"
	"strings"
	"time"
)

func buildHTML(target string, hosts []hostEntry) string {
	var b strings.Builder

	totalPorts := 0
	for _, h := range hosts {
		totalPorts += len(h.ports)
	}

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>netscope — ` + html.EscapeString(target) + `</title>
<style>
:root {
  --bg: #0e1116;
  --panel: #161b22;
  --border: #30363d;
  --text: #e6edf3;
  --dim: #8b949e;
  --accent: #d2a8ff;
  --green: #3fb950;
  --blue: #58a6ff;
  --orange: #d29922;
}
* { box-sizing: border-box; }
body {
  margin: 0;
  padding: 40px 20px;
  background: var(--bg);
  color: var(--text);
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, monospace;
  font-size: 14px;
  line-height: 1.5;
}
.wrap { max-width: 960px; margin: 0 auto; }
h1 {
  font-size: 28px;
  margin: 0 0 4px 0;
  color: var(--accent);
}
.sub { color: var(--dim); margin-bottom: 24px; }
.sub a { color: var(--blue); text-decoration: none; }
.summary {
  display: flex;
  gap: 24px;
  padding: 16px 20px;
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 24px;
}
.summary .stat { display: flex; flex-direction: column; }
.summary .num { font-size: 24px; font-weight: 600; }
.summary .label { color: var(--dim); font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
.host {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 8px;
  margin-bottom: 16px;
  overflow: hidden;
}
.host-head {
  padding: 14px 20px;
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  gap: 16px;
}
.host-ip { font-size: 18px; font-weight: 600; font-family: monospace; }
.badge {
  font-size: 11px;
  padding: 3px 10px;
  border-radius: 12px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 600;
}
.badge-up { background: rgba(63,185,80,0.15); color: var(--green); }
.badge-os { background: rgba(210,168,255,0.15); color: var(--accent); }
table { width: 100%; border-collapse: collapse; }
th, td { text-align: left; padding: 10px 20px; border-bottom: 1px solid var(--border); }
th { color: var(--dim); font-weight: 500; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em; }
td.port { font-family: monospace; color: var(--orange); }
td.svc { color: var(--blue); font-family: monospace; }
td.banner { color: var(--dim); font-family: monospace; font-size: 12px; word-break: break-all; }
tr:last-child td { border-bottom: none; }
.footer {
  margin-top: 40px;
  padding-top: 20px;
  border-top: 1px solid var(--border);
  color: var(--dim);
  font-size: 12px;
  text-align: center;
}
.footer a { color: var(--blue); text-decoration: none; }
</style>
</head>
<body>
<div class="wrap">
`)

	// header
	b.WriteString(`<h1>netscope</h1>` + "\n")
	b.WriteString(`<div class="sub">v` + version + ` · by <a href="https://github.com/zasmastro">cortex</a> · <a href="` + repo + `">` + repo + `</a></div>` + "\n")
	b.WriteString(`<div class="sub">target: <strong>` + html.EscapeString(target) + `</strong> · scanned at ` + time.Now().Format("2006-01-02 15:04:05") + `</div>` + "\n")

	// summary
	b.WriteString(`<div class="summary">` + "\n")
	b.WriteString(`  <div class="stat"><span class="num">` + fmt.Sprintf("%d", len(hosts)) + `</span><span class="label">hosts</span></div>` + "\n")
	b.WriteString(`  <div class="stat"><span class="num">` + fmt.Sprintf("%d", totalPorts) + `</span><span class="label">open ports</span></div>` + "\n")
	b.WriteString(`</div>` + "\n")

	// sort hosts by IP
	sorted := make([]hostEntry, len(hosts))
	copy(sorted, hosts)
	sort.Slice(sorted, func(i, j int) bool {
		return ipLess(sorted[i].ip, sorted[j].ip)
	})

	for _, h := range sorted {
		b.WriteString(`<div class="host">` + "\n")
		b.WriteString(`  <div class="host-head">` + "\n")
		b.WriteString(`    <span class="host-ip">` + html.EscapeString(h.ip) + `</span>` + "\n")
		b.WriteString(`    <span class="badge badge-up">● up</span>` + "\n")
		guess := guessOS(&h)
		if guess != "" {
			b.WriteString(`    <span class="badge badge-os">` + html.EscapeString(guess) + `</span>` + "\n")
		}
		b.WriteString(`  </div>` + "\n")

		if len(h.ports) == 0 {
			b.WriteString(`  <table><tr><td colspan="3" style="padding:14px 20px;color:var(--dim)">no open ports</td></tr></table>` + "\n")
		} else {
			b.WriteString(`  <table>` + "\n")
			b.WriteString(`    <thead><tr><th>Port</th><th>Service</th><th>Banner</th></tr></thead>` + "\n")
			b.WriteString(`    <tbody>` + "\n")
			for _, p := range h.ports {
				svc := portNames[p]
				if svc == "" {
					svc = "—"
				}
				banner := h.banners[p]
				if banner == "" {
					banner = "—"
				}
				b.WriteString(`      <tr>`)
				b.WriteString(`<td class="port">` + fmt.Sprintf("%d", p) + `</td>`)
				b.WriteString(`<td class="svc">` + html.EscapeString(svc) + `</td>`)
				b.WriteString(`<td class="banner">` + html.EscapeString(banner) + `</td>`)
				b.WriteString(`</tr>` + "\n")
			}
			b.WriteString(`    </tbody>` + "\n")
			b.WriteString(`  </table>` + "\n")
		}

		b.WriteString(`</div>` + "\n")
	}

	// footer
	b.WriteString(`<div class="footer">Generated by netscope v` + version + ` — <a href="` + repo + `">` + repo + `</a></div>` + "\n")
	b.WriteString(`</div>` + "\n")
	b.WriteString(`</body>` + "\n")
	b.WriteString(`</html>` + "\n")

	return b.String()
}
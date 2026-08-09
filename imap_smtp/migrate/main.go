package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/joho/godotenv"
)

func detectProvider(imapHost string) string {
	hosts := map[string]string{
		"imap.163.com": "163", "imap.vip.163.com": "vip.163",
		"imap.126.com": "126", "imap.vip.126.com": "vip.126",
		"imap.188.com": "188", "imap.vip.188.com": "vip.188",
		"imap.yeah.net": "yeah", "imap.gmail.com": "gmail",
		"outlook.office365.com": "outlook", "imap.qq.com": "qq",
		"imap.exmail.qq.com": "exmail.qq", "imap.mail.me.com": "icloud",
		"imap.fastmail.com": "fastmail", "imap.qiye.163.com": "netease-enterprise-north",
	}
	if name, ok := hosts[imapHost]; ok {
		return name
	}
	return "custom"
}

func quote(s string) string {
	if s == "" {
		return ""
	}
	if matched, _ := regexp.MatchString(`\s`, s); matched || strings.ContainsAny(s, `#"$'\\`) {
		return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
	}
	return s
}

func expandHome(paths string) string {
	home, _ := os.UserHomeDir()
	var parts []string
	for _, s := range strings.Split(paths, ",") {
		s = strings.TrimSpace(s)
		if strings.HasPrefix(s, "~/") {
			s = filepath.Join(home, s[2:])
		} else if s == "~" {
			s = home
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, ",")
}

func accountBlock(env map[string]string, prefix, name string) []string {
	p := prefix
	imapHost := env[p+"IMAP_HOST"]
	user := env[p+"IMAP_USER"]
	pass := env[p+"IMAP_PASS"]
	provider := detectProvider(imapHost)

	var lines []string
	lines = append(lines, "")
	lines = append(lines, "# "+name+" account")

	if provider == "custom" {
		lines = append(lines, p+"PROVIDER=custom")
		lines = append(lines, p+"USERNAME="+quote(user))
		lines = append(lines, p+"PASSWORD="+quote(pass))
		lines = append(lines, p+"IMAP_HOST="+quote(imapHost))
		if v, ok := env[p+"IMAP_PORT"]; ok && v != "" {
			lines = append(lines, p+"IMAP_PORT="+v)
		} else {
			lines = append(lines, p+"IMAP_PORT=993")
		}
		if v, ok := env[p+"IMAP_TLS"]; ok && v != "" {
			lines = append(lines, p+"IMAP_TLS="+v)
		} else {
			lines = append(lines, p+"IMAP_TLS=true")
		}
		lines = append(lines, p+"SMTP_HOST="+quote(env[p+"SMTP_HOST"]))
		if v, ok := env[p+"SMTP_PORT"]; ok && v != "" {
			lines = append(lines, p+"SMTP_PORT="+v)
		} else {
			lines = append(lines, p+"SMTP_PORT=587")
		}
		if v, ok := env[p+"SMTP_SECURE"]; ok && v != "" {
			lines = append(lines, p+"SMTP_SECURE="+v)
		} else {
			lines = append(lines, p+"SMTP_SECURE=false")
		}
	} else {
		lines = append(lines, p+"PROVIDER="+provider)
		lines = append(lines, p+"USERNAME="+quote(user))
		lines = append(lines, p+"PASSWORD="+quote(pass))
	}

	if v, ok := env[p+"IMAP_REJECT_UNAUTHORIZED"]; ok && v == "false" {
		lines = append(lines, p+"IMAP_REJECT_UNAUTHORIZED=false")
	}
	if v, ok := env[p+"SMTP_REJECT_UNAUTHORIZED"]; ok && v == "false" {
		lines = append(lines, p+"SMTP_REJECT_UNAUTHORIZED=false")
	}

	return lines
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: migrate-legacy-config <legacy-env-path>")
		os.Exit(1)
	}

	legacyPath := os.Args[1]
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Legacy config not found: %s\n", legacyPath)
		os.Exit(1)
	}

	env, err := godotenv.Read(legacyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse legacy config: %v\n", err)
		os.Exit(1)
	}

	var blocks []string
	seen := make(map[string]bool)

	if _, ok := env["IMAP_HOST"]; ok {
		blocks = append(blocks, accountBlock(env, "", "default")...)
		seen["default"] = true
	}

	// Find prefixed accounts
	prefixes := make(map[string]bool)
	for key := range env {
		if matched, _ := regexp.MatchString(`^[A-Z0-9]+_IMAP_HOST$`, key); matched {
			prefix := strings.TrimSuffix(key, "_IMAP_HOST")
			prefixes[prefix] = true
		}
	}
	var sortedPrefixes []string
	for p := range prefixes {
		sortedPrefixes = append(sortedPrefixes, p)
	}
	sort.Strings(sortedPrefixes)

	for _, prefix := range sortedPrefixes {
		name := strings.ToLower(prefix)
		if seen[name] {
			continue
		}
		seen[name] = true
		blocks = append(blocks, accountBlock(env, prefix+"_", name)...)
	}

	// File access whitelist
	readDirs := expandHome(env["ALLOWED_READ_DIRS"])
	if readDirs == "" {
		readDirs = expandHome("~/Downloads,~/Documents")
	}
	writeDirs := expandHome(env["ALLOWED_WRITE_DIRS"])
	if writeDirs == "" {
		writeDirs = expandHome("~/Downloads")
	}

	blocks = append(blocks, "")
	blocks = append(blocks, "# File access whitelist (security)")
	blocks = append(blocks, "ALLOWED_READ_DIRS="+readDirs)
	blocks = append(blocks, "ALLOWED_WRITE_DIRS="+writeDirs)

	output := strings.Join(blocks, "\n")
	output = strings.TrimPrefix(output, "\n") // remove leading blank line
	fmt.Print(output + "\n")
}

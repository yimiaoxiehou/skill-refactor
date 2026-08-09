package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/charset"
	gomessage "github.com/emersion/go-message/mail"
	"github.com/joho/godotenv"
)

func init() {
	imap.CharsetReader = charset.Reader
}

// ---- Provider Presets ----

type imapPreset struct {
	Host              string
	Port              int
	TLS               bool
	RejectUnauthorized bool
}

type smtpPreset struct {
	Host              string
	Port              int
	Secure            bool
	RejectUnauthorized bool
}

type providerCfg struct {
	IMAP imapPreset
	SMTP smtpPreset
}

var providers = map[string]providerCfg{
	"163":          {IMAP: imapPreset{"imap.163.com", 993, true, true}, SMTP: smtpPreset{"smtp.163.com", 465, true, true}},
	"vip.163":      {IMAP: imapPreset{"imap.vip.163.com", 993, true, true}, SMTP: smtpPreset{"smtp.vip.163.com", 465, true, true}},
	"126":          {IMAP: imapPreset{"imap.126.com", 993, true, true}, SMTP: smtpPreset{"smtp.126.com", 465, true, true}},
	"vip.126":      {IMAP: imapPreset{"imap.vip.126.com", 993, true, true}, SMTP: smtpPreset{"smtp.vip.126.com", 465, true, true}},
	"188":          {IMAP: imapPreset{"imap.188.com", 993, true, true}, SMTP: smtpPreset{"smtp.188.com", 465, true, true}},
	"vip.188":      {IMAP: imapPreset{"imap.vip.188.com", 993, true, true}, SMTP: smtpPreset{"smtp.vip.188.com", 465, true, true}},
	"yeah":         {IMAP: imapPreset{"imap.yeah.net", 993, true, true}, SMTP: smtpPreset{"smtp.yeah.net", 465, true, true}},
	"gmail":        {IMAP: imapPreset{"imap.gmail.com", 993, true, true}, SMTP: smtpPreset{"smtp.gmail.com", 587, false, true}},
	"outlook":      {IMAP: imapPreset{"outlook.office365.com", 993, true, true}, SMTP: smtpPreset{"smtp.office365.com", 587, false, true}},
	"qq":           {IMAP: imapPreset{"imap.qq.com", 993, true, true}, SMTP: smtpPreset{"smtp.qq.com", 587, false, true}},
	"exmail.qq":    {IMAP: imapPreset{"imap.exmail.qq.com", 993, true, true}, SMTP: smtpPreset{"smtp.exmail.qq.com", 465, true, true}},
	"icloud":       {IMAP: imapPreset{"imap.mail.me.com", 993, true, true}, SMTP: smtpPreset{"smtp.mail.me.com", 587, false, true}},
	"fastmail":     {IMAP: imapPreset{"imap.fastmail.com", 993, true, true}, SMTP: smtpPreset{"smtp.fastmail.com", 465, true, true}},
	"netease-enterprise-north": {IMAP: imapPreset{"imap.qiye.163.com", 993, true, true}, SMTP: smtpPreset{"smtp.qiye.163.com", 465, true, true}},
	"netease-enterprise-east":  {IMAP: imapPreset{"imap.qiye.163.com", 993, true, true}, SMTP: smtpPreset{"smtp.qiye.163.com", 465, true, true}},
}

func detectProvider(imapHost string) string {
	for name, p := range providers {
		if p.IMAP.Host == imapHost {
			return name
		}
	}
	return ""
}

// ---- Config ----

type imapConfig struct {
	Host              string
	Port              int
	User              string
	Pass              string
	TLS               bool
	RejectUnauthorized bool
	Mailbox           string
}

type smtpConfig struct {
	Host              string
	Port              int
	User              string
	Pass              string
	Secure            bool
	RejectUnauthorized bool
	From              string
}

type config struct {
	IMAP             imapConfig
	SMTP             smtpConfig
	AllowedReadDirs  []string
	AllowedWriteDirs []string
}

func loadConfig() *config {
	home, _ := os.UserHomeDir()
	legacyPath := filepath.Join(home, ".config", "imap-smtp-email", ".env")
	sharedPath := filepath.Join(home, ".config", "mail-skills", ".env")
	skillPath, _ := filepath.Abs(filepath.Join(filepath.Dir(os.Args[0]), "..", ".env"))

	var cfgPath, cfgType string
	if _, err := os.Stat(legacyPath); err == nil {
		cfgPath = legacyPath
		cfgType = "legacy"
	} else if _, err := os.Stat(sharedPath); err == nil {
		cfgPath = sharedPath
		cfgType = "shared"
	} else if _, err := os.Stat(skillPath); err == nil {
		cfgPath = skillPath
		cfgType = "legacy"
	}

	if cfgPath == "" {
		fmt.Fprintln(os.Stderr, "Error: No email configuration found. Run setup.")
		os.Exit(1)
	}

	accountName := ""
	for i, arg := range os.Args {
		if arg == "--account" && i+1 < len(os.Args) {
			accountName = os.Args[i+1]
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			break
		}
	}

	envMap, err := godotenv.Read(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	prefix := ""
	if accountName != "" {
		prefix = strings.ToUpper(accountName) + "_"
	}

	var cfg *config
	if cfgType == "shared" {
		cfg = buildFromShared(envMap, prefix)
	} else {
		cfg = buildFromLegacy(envMap, prefix)
	}

	if cfg == nil {
		fmt.Fprintf(os.Stderr, "Error: Account not found or incomplete in %s\n", cfgPath)
		os.Exit(1)
	}

	return cfg
}

func buildFromLegacy(env map[string]string, prefix string) *config {
	p := prefix

	imapTLS := true
	if v, ok := env[p+"IMAP_TLS"]; ok {
		imapTLS = v == "true"
	}
	imapRejectUnauth := true
	if v, ok := env[p+"IMAP_REJECT_UNAUTHORIZED"]; ok {
		imapRejectUnauth = v != "false"
	}
	smtpSecure := false
	if v, ok := env[p+"SMTP_SECURE"]; ok {
		smtpSecure = v == "true"
	}
	smtpRejectUnauth := true
	if v, ok := env[p+"SMTP_REJECT_UNAUTHORIZED"]; ok {
		smtpRejectUnauth = v != "false"
	}
	imapMailbox := "INBOX"
	if v, ok := env[p+"IMAP_MAILBOX"]; ok {
		imapMailbox = v
	}
	from := env[p+"SMTP_USER"]
	if v, ok := env[p+"SMTP_FROM"]; ok && v != "" {
		from = v
	}

	cfg := &config{
		IMAP: imapConfig{
			Host: env[p+"IMAP_HOST"], Port: atoi(env[p+"IMAP_PORT"], 993),
			User: env[p+"IMAP_USER"], Pass: env[p+"IMAP_PASS"],
			TLS: imapTLS, RejectUnauthorized: imapRejectUnauth, Mailbox: imapMailbox,
		},
		SMTP: smtpConfig{
			Host: env[p+"SMTP_HOST"], Port: atoi(env[p+"SMTP_PORT"], 587),
			User: env[p+"SMTP_USER"], Pass: env[p+"SMTP_PASS"],
			Secure: smtpSecure, RejectUnauthorized: smtpRejectUnauth, From: from,
		},
	}

	if rd, ok := env["ALLOWED_READ_DIRS"]; ok {
		for _, d := range strings.Split(rd, ",") {
			if t := strings.TrimSpace(d); t != "" {
				cfg.AllowedReadDirs = append(cfg.AllowedReadDirs, expandHome(t))
			}
		}
	}
	if wd, ok := env["ALLOWED_WRITE_DIRS"]; ok {
		for _, d := range strings.Split(wd, ",") {
			if t := strings.TrimSpace(d); t != "" {
				cfg.AllowedWriteDirs = append(cfg.AllowedWriteDirs, expandHome(t))
			}
		}
	}
	return cfg
}

func buildFromShared(env map[string]string, prefix string) *config {
	p := prefix
	providerKey := env[p+"PROVIDER"]
	if providerKey == "" {
		return nil
	}
	username := env[p+"USERNAME"]
	password := env[p+"PASSWORD"]
	if username == "" || password == "" {
		return nil
	}

	var imapP imapPreset
	var smtpP smtpPreset

	if providerKey == "custom" {
		imapTLS := true
		if v := env[p+"IMAP_TLS"]; v == "false" {
			imapTLS = false
		}
		imapRA := true
		if v := env[p+"IMAP_REJECT_UNAUTHORIZED"]; v == "false" {
			imapRA = false
		}
		imapP = imapPreset{Host: env[p+"IMAP_HOST"], Port: atoi(env[p+"IMAP_PORT"], 993), TLS: imapTLS, RejectUnauthorized: imapRA}

		smtpSecure := false
		if v := env[p+"SMTP_SECURE"]; v == "true" {
			smtpSecure = true
		}
		smtpRA := true
		if v := env[p+"SMTP_REJECT_UNAUTHORIZED"]; v == "false" {
			smtpRA = false
		}
		smtpP = smtpPreset{Host: env[p+"SMTP_HOST"], Port: atoi(env[p+"SMTP_PORT"], 587), Secure: smtpSecure, RejectUnauthorized: smtpRA}
	} else {
		preset, ok := providers[providerKey]
		if !ok || preset.IMAP.Host == "" {
			return nil
		}
		imapP = preset.IMAP
		smtpP = preset.SMTP
		if v := env[p+"IMAP_REJECT_UNAUTHORIZED"]; v == "false" {
			imapP.RejectUnauthorized = false
		}
		if v := env[p+"SMTP_REJECT_UNAUTHORIZED"]; v == "false" {
			smtpP.RejectUnauthorized = false
		}
	}

	imapMailbox := "INBOX"
	if v, ok := env[p+"IMAP_MAILBOX"]; ok {
		imapMailbox = v
	}
	from := username
	if v, ok := env[p+"SMTP_FROM"]; ok && v != "" {
		from = v
	}

	cfg := &config{
		IMAP: imapConfig{
			Host: imapP.Host, Port: imapP.Port,
			User: username, Pass: password,
			TLS: imapP.TLS, RejectUnauthorized: imapP.RejectUnauthorized,
			Mailbox: imapMailbox,
		},
		SMTP: smtpConfig{
			Host: smtpP.Host, Port: smtpP.Port,
			User: username, Pass: password,
			Secure: smtpP.Secure, RejectUnauthorized: smtpP.RejectUnauthorized,
			From: from,
		},
	}

	if rd, ok := env["ALLOWED_READ_DIRS"]; ok {
		for _, d := range strings.Split(rd, ",") {
			if t := strings.TrimSpace(d); t != "" {
				cfg.AllowedReadDirs = append(cfg.AllowedReadDirs, expandHome(t))
			}
		}
	} else {
		cfg.AllowedReadDirs = []string{expandHome("~/Downloads"), expandHome("~/Documents")}
	}
	if wd, ok := env["ALLOWED_WRITE_DIRS"]; ok {
		for _, d := range strings.Split(wd, ",") {
			if t := strings.TrimSpace(d); t != "" {
				cfg.AllowedWriteDirs = append(cfg.AllowedWriteDirs, expandHome(t))
			}
		}
	} else {
		cfg.AllowedWriteDirs = []string{expandHome("~/Downloads")}
	}
	return cfg
}

func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}

func expandHome(p string) string {
	p = strings.TrimSpace(p)
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	if p == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	return p
}

// ---- IMAP Connection ----

func imapConnect(cfg *config) (*client.Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.IMAP.Host, cfg.IMAP.Port)

	var c *client.Client
	var err error
	if cfg.IMAP.TLS {
		tlsCfg := &tls.Config{ServerName: cfg.IMAP.Host, InsecureSkipVerify: !cfg.IMAP.RejectUnauthorized}
		c, err = client.DialTLS(addr, tlsCfg)
	} else {
		c, err = client.Dial(addr)
	}
	if err != nil {
		return nil, fmt.Errorf("IMAP connection failed: %v", err)
	}

	if err := c.Login(cfg.IMAP.User, cfg.IMAP.Pass); err != nil {
		c.Logout()
		return nil, fmt.Errorf("IMAP login failed: %v", err)
	}

	return c, nil
}

// ---- Mail Parsing ----

func parseEmailBody(body []byte) (map[string]interface{}, error) {
	mr, err := gomessage.CreateReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}

	header := mr.Header
	from := header.Get("From")
	to := header.Get("To")
	subj := header.Get("Subject")
	if subj == "" {
		subj = "(no subject)"
	}
	dt := header.Get("Date")

	var text, html string
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		switch h := p.Header.(type) {
		case *gomessage.InlineHeader:
			ct := h.Get("Content-Type")
			bodyBytes, _ := io.ReadAll(p.Body)
			if strings.Contains(ct, "text/plain") {
				text = string(bodyBytes)
			} else if strings.Contains(ct, "text/html") {
				html = string(bodyBytes)
			}
		}
	}

	snippet := text
	if snippet == "" && html != "" {
		re := regexp.MustCompile(`<[^>]*>`)
		snippet = strings.TrimSpace(re.ReplaceAllString(html, ""))
	}
	if len([]rune(snippet)) > 200 {
		snippet = string([]rune(snippet)[:200])
	}

	return map[string]interface{}{
		"from": from, "to": to, "subject": subj,
		"headerDate": dt, "text": text, "html": html, "snippet": snippet,
	}, nil
}

// ---- IMAP Commands ----

var imapSeqSet = new(imap.SeqSet)

func imapCheck(cfg *config, mailbox string, limit int, recentTime string, unreadOnly bool) {
	c, err := imapConnect(cfg)
	if err != nil {
		exitErr(err)
	}
	defer c.Logout()

	mboxStatus, err := c.Select(mailbox, true)
	if err != nil {
		exitErr(err)
	}
	_ = mboxStatus

	criteria := imap.NewSearchCriteria()
	if recentTime != "" {
		criteria.Since = parseRelativeTime(recentTime)
	}

	var seqSet *imap.SeqSet
	var uids []uint32

	if unreadOnly {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}

	uids, err = c.UidSearch(criteria)
	if err != nil {
		// fallback: search recent messages
		seqNums, err := c.Search(criteria)
		if err != nil {
			exitErr(err)
		}
		if len(seqNums) == 0 {
			outputJSON(map[string]interface{}{"results": []interface{}{}, "meta": map[string]interface{}{"fallbackUsed": false, "returned": 0}})
			return
		}
		if len(seqNums) > limit {
			seqNums = seqNums[len(seqNums)-limit:]
		}
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(seqNums...)
	} else {
		if len(uids) == 0 {
			outputJSON(map[string]interface{}{"results": []interface{}{}, "meta": map[string]interface{}{"fallbackUsed": false, "returned": 0}})
			return
		}
		if len(uids) > limit {
			uids = uids[len(uids)-limit:]
		}
		seqSet = new(imap.SeqSet)
		seqSet.AddNum(uids...)
	}

	items := make(chan *imap.Message, 10)
	fetchOpts := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchUid, imap.FetchBodyStructure}
	fetchOpts = append(fetchOpts, "BODY[]")

	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqSet, fetchOpts, items)
	}()

	var msgs []*imap.Message
	for msg := range items {
		msgs = append(msgs, msg)
	}
	if err := <-done; err != nil {
		exitErr(err)
	}

	// Reverse for newest first
	var results []map[string]interface{}
	for i := len(msgs) - 1; i >= 0; i-- {
		msg := msgs[i]
		body := getBody(msg)
		if body == nil {
			continue
		}
		parsed, err := parseEmailBody(body)
		if err != nil {
			continue
		}
		parsed["uid"] = msg.Uid
		parsed["date"] = msg.InternalDate.Format(time.RFC3339)
		parsed["flags"] = msg.Flags
		results = append(results, parsed)
	}

	outputJSON(map[string]interface{}{
		"results": results,
		"meta":    map[string]interface{}{"fallbackUsed": false, "provider": detectProvider(cfg.IMAP.Host), "returned": len(results)},
	})
}

func getBody(msg *imap.Message) []byte {
	for _, literal := range msg.Body {
		data, err := io.ReadAll(literal)
		if err == nil && len(data) > 0 {
			return data
		}
	}
	return nil
}

func imapSearch(cfg *config, mailbox string, from, subject, recent, since, before string, unseen, seen bool, limit int) {
	c, err := imapConnect(cfg)
	if err != nil {
		exitErr(err)
	}
	defer c.Logout()

	_, err = c.Select(mailbox, true)
	if err != nil {
		exitErr(err)
	}

	if unseen && seen {
		fmt.Fprintln(os.Stderr, "Error: --unseen and --seen cannot be used together")
		os.Exit(1)
	}

	criteria := imap.NewSearchCriteria()
	if recent != "" {
		criteria.Since = parseRelativeTime(recent)
	} else {
		if since != "" {
			t, _ := time.Parse("2006-01-02", since)
			criteria.Since = t
		}
		if before != "" {
			t, _ := time.Parse("2006-01-02", before)
			criteria.Before = t
		}
	}
	if from != "" {
		criteria.Header.Add("FROM", from)
	}
	if subject != "" {
		criteria.Header.Add("SUBJECT", subject)
	}
	if unseen {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}
	if seen {
		criteria.WithFlags = []string{imap.SeenFlag}
	}

	uids, err := c.UidSearch(criteria)
	if err != nil {
		exitErr(err)
	}

	if len(uids) == 0 {
		outputJSON(map[string]interface{}{"results": []interface{}{}, "meta": map[string]interface{}{"fallbackUsed": false, "returned": 0}})
		return
	}

	if len(uids) > limit {
		uids = uids[len(uids)-limit:]
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	items := make(chan *imap.Message, 10)
	fetchOpts := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchUid}
	fetchOpts = append(fetchOpts, "BODY[]")

	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqSet, fetchOpts, items)
	}()

	var msgs []*imap.Message
	for msg := range items {
		msgs = append(msgs, msg)
	}
	if err := <-done; err != nil {
		exitErr(err)
	}

	var results []map[string]interface{}
	for _, msg := range msgs {
		body := getBody(msg)
		if body == nil {
			continue
		}
		parsed, _ := parseEmailBody(body)
		parsed["uid"] = msg.Uid
		parsed["date"] = msg.InternalDate.Format(time.RFC3339)
		parsed["flags"] = msg.Flags
		results = append(results, parsed)
	}

	sort.Slice(results, func(i, j int) bool {
		di := results[i]["date"].(string)
		dj := results[j]["date"].(string)
		return di > dj
	})
	if len(results) > limit {
		results = results[:limit]
	}

	outputJSON(map[string]interface{}{
		"results": results,
		"meta":    map[string]interface{}{"fallbackUsed": false, "provider": detectProvider(cfg.IMAP.Host), "returned": len(results)},
	})
}

func imapFetch(cfg *config, uid uint32, mailbox string) {
	c, err := imapConnect(cfg)
	if err != nil {
		exitErr(err)
	}
	defer c.Logout()

	_, err = c.Select(mailbox, true)
	if err != nil {
		exitErr(err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid)

	items := make(chan *imap.Message, 1)
	fetchOpts := []imap.FetchItem{imap.FetchEnvelope, imap.FetchFlags, imap.FetchInternalDate, imap.FetchUid}
	fetchOpts = append(fetchOpts, "BODY[]")

	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqSet, fetchOpts, items)
	}()

	var msgs []*imap.Message
	for msg := range items {
		msgs = append(msgs, msg)
	}
	if err := <-done; err != nil {
		exitErr(err)
	}

	if len(msgs) == 0 {
		fmt.Fprintf(os.Stderr, "Message UID %d not found\n", uid)
		os.Exit(1)
	}

	msg := msgs[0]
	body := getBody(msg)
	if body == nil {
		fmt.Fprintf(os.Stderr, "Message body empty\n")
		os.Exit(1)
	}

	parsed, _ := parseEmailBody(body)
	parsed["uid"] = msg.Uid
	parsed["date"] = msg.InternalDate.Format(time.RFC3339)
	parsed["flags"] = msg.Flags
	outputJSON(parsed)
}

func imapMarkRead(cfg *config, uids []uint32, mailbox string, read bool) {
	c, err := imapConnect(cfg)
	if err != nil {
		exitErr(err)
	}
	defer c.Logout()

	_, err = c.Select(mailbox, false)
	if err != nil {
		exitErr(err)
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uids...)

	flags := []interface{}{imap.SeenFlag}
	if read {
		err = c.UidStore(seqSet, imap.FormatFlagsOp(imap.AddFlags, true), flags, nil)
	} else {
		err = c.UidStore(seqSet, imap.FormatFlagsOp(imap.RemoveFlags, true), flags, nil)
	}
	if err != nil {
		exitErr(err)
	}

	action := "marked as read"
	if !read {
		action = "marked as unread"
	}
	outputJSON(map[string]interface{}{"success": true, "uids": uids, "action": action})
}

func imapListMailboxes(cfg *config) {
	c, err := imapConnect(cfg)
	if err != nil {
		exitErr(err)
	}
	defer c.Logout()

	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var result []map[string]interface{}
	for mb := range mailboxes {
		result = append(result, map[string]interface{}{
			"name":       mb.Name,
			"delimiter":  mb.Delimiter,
			"attributes": mb.Attributes,
		})
	}
	if err := <-done; err != nil {
		exitErr(err)
	}

	outputJSON(result)
}

// ---- SMTP Commands ----

func smtpSend(cfg *config, to, cc, bcc, subject, body, htmlBody, from string) {
	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)

	var auth smtp.Auth
	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = cfg.SMTP.Host
	}
	if cfg.SMTP.User != "" {
		auth = smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, host)
	}

	sender := cfg.SMTP.From
	if from != "" {
		sender = from
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", sender))
	recipients := splitAndTrim(to)
	sb.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ", ")))
	if cc != "" {
		sb.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")

	if htmlBody != "" {
		sb.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
		sb.WriteString("\r\n" + htmlBody)
	} else {
		sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
		sb.WriteString("\r\n" + body)
	}

	allRcpt := append(recipients, splitAndTrim(cc)...)
	allRcpt = append(allRcpt, splitAndTrim(bcc)...)

	if len(allRcpt) == 0 {
		fmt.Fprintln(os.Stderr, "No recipients")
		os.Exit(1)
	}

	var client *smtp.Client
	var smtpConnErr error
	if cfg.SMTP.Secure {
		tlsCfg := &tls.Config{ServerName: cfg.SMTP.Host, InsecureSkipVerify: !cfg.SMTP.RejectUnauthorized}
		tlsConn, tlsErr := tls.Dial("tcp", addr, tlsCfg)
		if tlsErr != nil {
			exitErr(fmt.Errorf("SMTP connection failed: %v", tlsErr))
		}
		client, smtpConnErr = smtp.NewClient(tlsConn, host)
	} else {
		client, smtpConnErr = smtp.Dial(addr)
	}
	if smtpConnErr != nil {
		exitErr(fmt.Errorf("SMTP connection failed: %v", smtpConnErr))
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			exitErr(fmt.Errorf("SMTP auth failed: %v", err))
		}
	}

	if err := client.Mail(sender); err != nil {
		exitErr(fmt.Errorf("SMTP MAIL FROM failed: %v", err))
	}
	for _, r := range allRcpt {
		if err := client.Rcpt(r); err != nil {
			exitErr(fmt.Errorf("SMTP RCPT TO failed for %s: %v", r, err))
		}
	}

	wc, err := client.Data()
	if err != nil {
		exitErr(fmt.Errorf("SMTP DATA failed: %v", err))
	}
	_, err = wc.Write([]byte(sb.String()))
	if err != nil {
		exitErr(fmt.Errorf("SMTP write failed: %v", err))
	}
	wc.Close()

	outputJSON(map[string]interface{}{"success": true, "to": to, "subject": subject})
}

func smtpTest(cfg *config) {
	addr := fmt.Sprintf("%s:%d", cfg.SMTP.Host, cfg.SMTP.Port)

	var auth smtp.Auth
	host, _, _ := net.SplitHostPort(addr)
	if host == "" {
		host = cfg.SMTP.Host
	}
	if cfg.SMTP.User != "" {
		auth = smtp.PlainAuth("", cfg.SMTP.User, cfg.SMTP.Pass, host)
	}

	sender := cfg.SMTP.From
	if sender == "" {
		sender = cfg.SMTP.User
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: SMTP Connection Test\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nThis is a test email from the IMAP/SMTP email skill.\r\n", sender, cfg.SMTP.User)

	var client *smtp.Client
	var err error
	if cfg.SMTP.Secure {
		tlsCfg := &tls.Config{ServerName: cfg.SMTP.Host, InsecureSkipVerify: !cfg.SMTP.RejectUnauthorized}
		tlsConn, tlsErr := tls.Dial("tcp", addr, tlsCfg)
		if tlsErr != nil {
			exitErr(fmt.Errorf("SMTP test failed: %v", tlsErr))
		}
		client, err = smtp.NewClient(tlsConn, host)
	} else {
		client, err = smtp.Dial(addr)
	}
	if err != nil {
		exitErr(fmt.Errorf("SMTP test failed: %v", err))
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			exitErr(fmt.Errorf("SMTP test auth failed: %v", err))
		}
	}

	if err := client.Mail(sender); err != nil {
		exitErr(fmt.Errorf("SMTP test MAIL FROM failed: %v", err))
	}
	if err := client.Rcpt(cfg.SMTP.User); err != nil {
		exitErr(fmt.Errorf("SMTP test RCPT TO failed: %v", err))
	}
	wc, err := client.Data()
	if err != nil {
		exitErr(fmt.Errorf("SMTP test DATA failed: %v", err))
	}
	wc.Write([]byte(msg))
	wc.Close()

	outputJSON(map[string]interface{}{"success": true, "message": "SMTP connection successful"})
}

// ---- Utilities ----

func parseRelativeTime(s string) time.Time {
	re := regexp.MustCompile(`^(\d+)(m|h|d)$`)
	matches := re.FindStringSubmatch(s)
	if matches == nil {
		return time.Time{}
	}
	val, _ := strconv.Atoi(matches[1])
	unit := matches[2]
	now := time.Now()
	switch unit {
	case "m":
		return now.Add(-time.Duration(val) * time.Minute)
	case "h":
		return now.Add(-time.Duration(val) * time.Hour)
	case "d":
		return now.Add(-time.Duration(val) * 24 * time.Hour)
	}
	return time.Time{}
}

func splitAndTrim(s string) []string {
	var result []string
	for _, item := range strings.Split(s, ",") {
		if t := strings.TrimSpace(item); t != "" {
			result = append(result, t)
		}
	}
	return result
}

func outputJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		exitErr(err)
	}
	fmt.Println(string(data))
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// ---- CLI Entry ----

func main() {
	cfg := loadConfig()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: imap-smtp <command> [options]")
		fmt.Fprintln(os.Stderr, "Commands: check, search, fetch, mark-read, mark-unread, list-mailboxes, send, test")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "check", "imap-check":
		mailbox := cfg.IMAP.Mailbox
		limit := 10
		var recent string
		unseen := false
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--mailbox":
				if i+1 < len(args) {
					mailbox = args[i+1]
					i++
				}
			case "--limit":
				if i+1 < len(args) {
					limit, _ = strconv.Atoi(args[i+1])
					i++
				}
			case "--recent":
				if i+1 < len(args) {
					recent = args[i+1]
					i++
				}
			case "--unseen":
				unseen = true
			}
		}
		imapCheck(cfg, mailbox, limit, recent, unseen)

	case "search", "imap-search":
		mailbox := cfg.IMAP.Mailbox
		limit := 20
		var from, subject, recent, since, before string
		var unseen, seen bool
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--mailbox":
				if i+1 < len(args) {
					mailbox = args[i+1]
					i++
				}
			case "--limit":
				if i+1 < len(args) {
					limit, _ = strconv.Atoi(args[i+1])
					i++
				}
			case "--from":
				if i+1 < len(args) {
					from = args[i+1]
					i++
				}
			case "--subject":
				if i+1 < len(args) {
					subject = args[i+1]
					i++
				}
			case "--recent":
				if i+1 < len(args) {
					recent = args[i+1]
					i++
				}
			case "--since":
				if i+1 < len(args) {
					since = args[i+1]
					i++
				}
			case "--before":
				if i+1 < len(args) {
					before = args[i+1]
					i++
				}
			case "--unseen":
				unseen = true
			case "--seen":
				seen = true
			}
		}
		imapSearch(cfg, mailbox, from, subject, recent, since, before, unseen, seen, limit)

	case "fetch", "imap-fetch":
		mailbox := cfg.IMAP.Mailbox
		var uid uint32
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--mailbox":
				if i+1 < len(args) {
					mailbox = args[i+1]
					i++
				}
			default:
				if v, err := strconv.ParseUint(args[i], 10, 32); err == nil {
					uid = uint32(v)
				}
			}
		}
		if uid == 0 {
			fmt.Fprintln(os.Stderr, "Error: UID required")
			os.Exit(1)
		}
		imapFetch(cfg, uid, mailbox)

	case "mark-read", "imap-mark-read":
		mailbox := cfg.IMAP.Mailbox
		var uids []uint32
		for _, a := range args {
			if strings.HasPrefix(a, "--") {
				continue
			}
			if v, err := strconv.ParseUint(a, 10, 32); err == nil {
				uids = append(uids, uint32(v))
			}
		}
		if len(uids) == 0 {
			fmt.Fprintln(os.Stderr, "Error: UID(s) required")
			os.Exit(1)
		}
		imapMarkRead(cfg, uids, mailbox, true)

	case "mark-unread", "imap-mark-unread":
		mailbox := cfg.IMAP.Mailbox
		var uids []uint32
		for _, a := range args {
			if strings.HasPrefix(a, "--") {
				continue
			}
			if v, err := strconv.ParseUint(a, 10, 32); err == nil {
				uids = append(uids, uint32(v))
			}
		}
		if len(uids) == 0 {
			fmt.Fprintln(os.Stderr, "Error: UID(s) required")
			os.Exit(1)
		}
		imapMarkRead(cfg, uids, mailbox, false)

	case "list-mailboxes", "imap-list-mailboxes":
		imapListMailboxes(cfg)

	case "send", "smtp-send":
		var to, cc, bcc, subject, body, htmlBody, from string
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "--to":
				if i+1 < len(args) {
					to = args[i+1]
					i++
				}
			case "--cc":
				if i+1 < len(args) {
					cc = args[i+1]
					i++
				}
			case "--bcc":
				if i+1 < len(args) {
					bcc = args[i+1]
					i++
				}
			case "--subject":
				if i+1 < len(args) {
					subject = args[i+1]
					i++
				}
			case "--body":
				if i+1 < len(args) {
					body = args[i+1]
					i++
				}
			case "--html":
				if i+1 < len(args) {
					htmlBody = args[i+1]
					i++
				}
			case "--from":
				if i+1 < len(args) {
					from = args[i+1]
					i++
				}
			}
		}
		if to == "" || subject == "" {
			fmt.Fprintln(os.Stderr, "Error: --to and --subject required")
			os.Exit(1)
		}
		smtpSend(cfg, to, cc, bcc, subject, body, htmlBody, from)

	case "test", "smtp-test":
		smtpTest(cfg)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		fmt.Fprintln(os.Stderr, "Commands: check, search, fetch, mark-read, mark-unread, list-mailboxes, send, test")
		os.Exit(1)
	}
}

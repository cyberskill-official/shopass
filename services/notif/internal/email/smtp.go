package email

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strings"
	"time"
)

// SMTPProvider sends via plain SMTP when credentials are configured.
// Construct only via NewSMTPFromEnv — missing env must stay on LogProvider
// (fail-closed: never pretend delivery succeeded).
type SMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPFromEnv returns a live SMTP provider when SMTP_HOST + SMTP_FROM are set
// and either (a) SMTP_USERNAME+SMTP_PASSWORD or (b) SMTP_AUTH=none for local
// relay. Returns nil when incomplete so callers keep the noop LogProvider.
func NewSMTPFromEnv() *SMTPProvider {
	host := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	from := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	if host == "" || from == "" {
		return nil
	}
	port := strings.TrimSpace(os.Getenv("SMTP_PORT"))
	if port == "" {
		port = "587"
	}
	user := strings.TrimSpace(os.Getenv("SMTP_USERNAME"))
	pass := os.Getenv("SMTP_PASSWORD")
	authMode := strings.ToLower(strings.TrimSpace(os.Getenv("SMTP_AUTH")))
	if authMode != "none" && (user == "" || pass == "") {
		return nil
	}
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: user,
		password: pass,
		from:     from,
	}
}

func (p *SMTPProvider) Send(ctx context.Context, msg EmailMessage) (SendOutcome, error) {
	if p == nil {
		return SendOutcome{Result: ResultFailed}, fmt.Errorf("smtp provider nil")
	}
	to := strings.TrimSpace(msg.To)
	if to == "" {
		return SendOutcome{Result: ResultPermanent}, nil
	}

	addr := net.JoinHostPort(p.host, p.port)
	var auth smtp.Auth
	if p.username != "" {
		auth = smtp.PlainAuth("", p.username, p.password, p.host)
	}

	body := msg.TextBody
	if body == "" {
		body = stripTags(msg.HTMLBody)
	}
	payload := []byte(fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nList-Unsubscribe: <mailto:unsubscribe@%s>\r\n\r\n%s",
		p.from, to, sanitizeHeader(msg.Subject), domainOf(p.from), body,
	))

	type result struct {
		err error
	}
	ch := make(chan result, 1)
	go func() {
		ch <- result{err: smtp.SendMail(addr, auth, p.from, []string{to}, payload)}
	}()

	select {
	case <-ctx.Done():
		return SendOutcome{Result: ResultRetry, RetryAfter: 30 * time.Second}, ctx.Err()
	case r := <-ch:
		if r.err != nil {
			return SendOutcome{Result: ResultRetry, RetryAfter: 60 * time.Second}, r.err
		}
		return SendOutcome{Result: ResultSent, ProviderMessageID: "smtp"}, nil
	}
}

func sanitizeHeader(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, s)
}

func domainOf(from string) string {
	if i := strings.LastIndex(from, "@"); i >= 0 && i+1 < len(from) {
		return from[i+1:]
	}
	return "shopass.local"
}

func stripTags(html string) string {
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}

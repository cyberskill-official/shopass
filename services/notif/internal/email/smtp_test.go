package email

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSMTPFromEnv_FailClosedWithoutCreds(t *testing.T) {
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_FROM", "")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	require.Nil(t, NewSMTPFromEnv())
}

func TestNewSMTPFromEnv_RequiresAuthUnlessNone(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_FROM", "alerts@shopass.example")
	t.Setenv("SMTP_USERNAME", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_AUTH", "")
	require.Nil(t, NewSMTPFromEnv())

	t.Setenv("SMTP_AUTH", "none")
	p := NewSMTPFromEnv()
	require.NotNil(t, p)
	require.Equal(t, "smtp.example.com", p.host)
	require.Equal(t, "alerts@shopass.example", p.from)
}

func TestNewSMTPFromEnv_WithUserPass(t *testing.T) {
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "465")
	t.Setenv("SMTP_FROM", "alerts@shopass.example")
	t.Setenv("SMTP_USERNAME", "u")
	t.Setenv("SMTP_PASSWORD", "p")
	t.Setenv("SMTP_AUTH", "")
	p := NewSMTPFromEnv()
	require.NotNil(t, p)
	require.Equal(t, "465", p.port)
	require.Equal(t, "u", p.username)
}

func TestSanitizeHeader(t *testing.T) {
	require.Equal(t, "Hello world", sanitizeHeader("Hello\r\n world"))
}

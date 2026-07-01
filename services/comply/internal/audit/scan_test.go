package audit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeFixture(t *testing.T, filename, content string) string {
	dir := t.TempDir()
	path := filepath.Join(dir, filename)
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
	return dir
}

func hasRule(findings []Finding, rule string) bool {
	for _, f := range findings {
		if f.Rule == rule {
			return true
		}
	}
	return false
}

func TestScan_CatchesPlatformToken(t *testing.T) {
	dir := writeFixture(t, "repo.go", `var shopeeToken = readCookie()`) // audit:allow fixture
	f, err := Scan(dir)
	require.NoError(t, err)
	require.True(t, hasRule(f, "platform_session_token"))
}

func TestScan_CatchesCleartextPassword(t *testing.T) {
	dir := writeFixture(t, "schema.sql", `CREATE TABLE u (password TEXT);`) // audit:allow fixture
	f, err := Scan(dir)
	require.NoError(t, err)
	require.True(t, hasRule(f, "cleartext_password"))
}

func TestScan_CleanCodePasses(t *testing.T) {
	dir := writeFixture(t, "ok.sql", `CREATE TABLE u (pwd_hash TEXT NOT NULL);`)
	f, err := Scan(dir)
	require.NoError(t, err)
	require.Empty(t, f)
}

func TestScan_AllowlistSkipped(t *testing.T) {
	dir := writeFixture(t, "x.go", `var token = jwtInternal() // audit:allow JWT noi bo, khong phai token san`)
	f, err := Scan(dir)
	require.NoError(t, err)
	require.Empty(t, f)
}

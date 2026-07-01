package region

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "countries-*.yaml")
	require.NoError(t, err)
	defer f.Close()
	_, err = f.WriteString(content)
	require.NoError(t, err)
	return f.Name()
}

func TestLoad_DuplicateCountry_Errors(t *testing.T) {
	dupCountryYAML := `countries:
  - country: VN
  - country: VN`
	_, err := Load(writeTemp(t, dupCountryYAML))
	require.Error(t, err)
}

func TestLoad_BadChannel_Errors(t *testing.T) {
	badChannelYAML := `countries:
  - country: VN
    affiliateChannelsAllowed: [telegram]`
	_, err := Load(writeTemp(t, badChannelYAML)) // channel "telegram"
	require.Error(t, err)
}

func TestLoad_NonAlpha2_Errors(t *testing.T) {
	_, err := Load(writeTemp(t, `countries: [{country: Vietnam}]`))
	require.Error(t, err)
}

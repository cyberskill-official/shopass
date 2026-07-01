package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHash_DifferentEachTime(t *testing.T) {
	h1, err := Hash("p@ss12345", defaultParams)
	require.NoError(t, err)
	h2, err := Hash("p@ss12345", defaultParams)
	require.NoError(t, err)

	require.NotEqual(t, h1, h2) // salt ngẫu nhiên
	require.True(t, strings.HasPrefix(h1, "$argon2id$"))
}

func TestVerify_CorrectAndWrong(t *testing.T) {
	h, err := Hash("p@ss12345", defaultParams)
	require.NoError(t, err)

	ok, err := Verify("p@ss12345", h)
	require.NoError(t, err)
	require.True(t, ok)

	bad, err := Verify("wrong", h)
	require.NoError(t, err)
	require.False(t, bad)
}

func TestVerify_OldParamsStillWork(t *testing.T) {
	weak := Argon2Params{Time: 1, Memory: 8 * 1024, Parallelism: 1, SaltLen: 16, KeyLen: 32}
	h, err := Hash("p@ss12345", weak) // băm bằng tham số cũ/yếu
	require.NoError(t, err)

	ok, err := Verify("p@ss12345", h) // verify dùng tham số đọc từ PHC
	require.NoError(t, err)
	require.True(t, ok)
}

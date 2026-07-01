package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

type Argon2Params struct {
	Time        uint32
	Memory      uint32
	Parallelism uint8
	SaltLen     uint32
	KeyLen      uint32
}

var defaultParams = Argon2Params{
	Time:        3,
	Memory:      64 * 1024,
	Parallelism: 2,
	SaltLen:     16,
	KeyLen:      32,
}

// Hash trả chuỗi PHC: $argon2id$v=19$m=65536,t=3,p=2$<salt_b64>$<hash_b64>
func Hash(password string, p Argon2Params) (string, error) {
	salt := make([]byte, p.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, p.KeyLen)
	return encodePHC(p, salt, key), nil
}

// Verify đọc tham số từ chính chuỗi PHC -> hash cũ vẫn verify sau khi nâng mặc định (§1 #4).
func Verify(password, phc string) (bool, error) {
	p, salt, want, err := decodePHC(phc)
	if err != nil {
		return false, err
	}
	got := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func encodePHC(p Argon2Params, salt, key []byte) string {
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Key := base64.RawStdEncoding.EncodeToString(key)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.Memory, p.Time, p.Parallelism, b64Salt, b64Key)
}

func decodePHC(phc string) (p Argon2Params, salt, key []byte, err error) {
	vals := strings.Split(phc, "$")
	if len(vals) != 6 {
		return p, nil, nil, ErrInvalidHash
	}
	if vals[1] != "argon2id" {
		return p, nil, nil, ErrInvalidHash
	}
	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return p, nil, nil, err
	}
	if version != argon2.Version {
		return p, nil, nil, ErrIncompatibleVersion
	}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Time, &p.Parallelism)
	if err != nil {
		return p, nil, nil, err
	}
	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return p, nil, nil, err
	}
	p.SaltLen = uint32(len(salt))
	key, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return p, nil, nil, err
	}
	p.KeyLen = uint32(len(key))
	return p, salt, key, nil
}

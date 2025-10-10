package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argonParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type PasswordManager struct {
	params *argonParams
	pepper []byte
}

func NewPasswordManager(pepper string) *PasswordManager {
	return &PasswordManager{
		params: &argonParams{
			memory:      64 * 1024,
			iterations:  3,
			parallelism: 2,
			saltLength:  16,
			keyLength:   32,
		},
		pepper: []byte(pepper),
	}
}

func (pm *PasswordManager) Hash(password string) (string, error) {
	salt := make([]byte, pm.params.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	passwordWithPepper := []byte(password + string(pm.pepper))
	hash := argon2.IDKey(passwordWithPepper, salt, pm.params.iterations, pm.params.memory, pm.params.parallelism, pm.params.keyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=<memory>,t=<iterations>,p=<parallelism>$<salt>$<hash>
	format := "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s"
	fullHash := fmt.Sprintf(
		format,
		argon2.Version,
		pm.params.memory,
		pm.params.iterations,
		pm.params.parallelism,
		b64Salt, b64Hash,
	)

	return fullHash, nil
}

func (pm *PasswordManager) Verify(password, encodedHash string) (bool, error) {
	p, salt, hash, err := pm.decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	passwordWithPepper := []byte(password + string(pm.pepper))
	otherHash := argon2.IDKey(passwordWithPepper, salt, p.iterations, p.memory, p.parallelism, p.keyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}

	return false, nil
}

func (pm *PasswordManager) decodeHash(encodedHash string) (params *argonParams, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, errors.New("invalid encoded hash format: incorrect number of parts")
	}

	if vals[1] != "argon2id" {
		return nil, nil, nil, errors.New("invalid hash algorithm: not argon2id")
	}

	var version int
	if _, err := fmt.Sscanf(vals[2], "v=%d", &version); err != nil || version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("incompatible argon2 version: %v", err)
	}

	p := &argonParams{}
	n, err := fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.memory, &p.iterations, &p.parallelism)
	if err != nil || n != 3 {
		return nil, nil, nil, fmt.Errorf("failed to parse argon2 parameters: %v", err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode salt: %v", err)
	}
	p.saltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to decode hash: %v", err)
	}
	p.keyLength = uint32(len(hash))

	return p, salt, hash, nil
}

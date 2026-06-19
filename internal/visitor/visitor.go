package visitor

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

const devSalt = "kiko-dev-change-me"

// Hasher computes daily-rotating visitor fingerprints without storing IP or UA.
type Hasher struct {
	salt string
	now  func() time.Time
}

// NewHasher returns a hasher. Empty salt uses a dev-only default — set visitor.salt in production.
func NewHasher(salt string) Hasher {
	if salt == "" {
		salt = devSalt
	}
	return Hasher{
		salt: salt,
		now:  time.Now,
	}
}

// DevSalt reports whether the hasher is using the built-in development salt.
func (h Hasher) DevSalt() bool {
	return h.salt == devSalt
}

// Hash returns SHA-256 hex of ip + ua + salt + UTC date. IP and UA are not persisted.
func (h Hasher) Hash(ip, ua string) string {
	day := h.now().UTC().Format("2006-01-02")
	payload := ip + "\n" + ua + "\n" + h.salt + "\n" + day
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

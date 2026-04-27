package encrypt

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashSHA256(input string) string {
	hasher := sha256.New()
	hasher.Write([]byte(input))
	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)
}

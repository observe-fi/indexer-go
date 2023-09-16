package util

import (
	"crypto/sha256"
	"fmt"
)

func HashID(key string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
}

package hash

import "crypto/sha256"

func Do(val []byte) []byte {
	sum := sha256.Sum256(val)
	return sum[:]
}

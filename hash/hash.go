package hash

import (
	"encoding/binary"

	"github.com/cespare/xxhash/v2"
)

func Do(val []byte) []byte {
	// hVal := xxhash.Sum64(val)

	sum := make([]byte, 8)
	binary.LittleEndian.PutUint64(sum, xxhash.Sum64(val))

	// sum := sha256.Sum256(val)
	// return sum[:]
	return sum
}

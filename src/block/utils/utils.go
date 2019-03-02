package bcutils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"math"
	"strconv"
)

func DoubleHashSha256(data []byte) []byte {
	tmp := sha256.Sum256(data)
	tmp = sha256.Sum256(tmp[:])
	return tmp[:]
}

func GetBytesWithNonce(msg []byte, nonce uint64) []byte {
	// TODO: can be optimized
	return append(msg, []byte(strconv.FormatUint(nonce, 10))...)
}

func ComputeNonceForPowWithCancel(msg []byte, difficulty int, ctx context.Context) uint64 {
	answer := make([]byte, difficulty)
	thisCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// TODO: can be optimized
	nonce := uint64(0)
	for ; nonce < math.MaxUint64; nonce++ {
		select {
		case <-thisCtx.Done():
			return 0
		default:
			digest := DoubleHashSha256(GetBytesWithNonce(msg, nonce))
			if bytes.Equal(digest[len(digest)-difficulty:], answer) {
				return nonce
			}
			nonce++
		}
	}
	// need to fallback to different way?
	panic("failed to find nonce")
	return 0
}

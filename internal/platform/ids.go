package platform

import (
	"crypto/rand"
	"encoding/hex"
)

type RandomIDs struct{}

func (RandomIDs) NewID(prefix string) string {
	var data [12]byte
	if _, err := rand.Read(data[:]); err != nil {
		panic(err)
	}
	return prefix + "_" + hex.EncodeToString(data[:])
}

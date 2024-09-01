package bannedDB

import (
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidKey = errors.New("invalid key")
)

func UserIDToKey(userID int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, userID)
	return buf[:n]
}

func KeyToUserID(key []byte) (int64, error) {
	i, n := binary.Varint(key)
	if n == 0 {
		return 0, ErrInvalidKey
	}
	return i, nil
}

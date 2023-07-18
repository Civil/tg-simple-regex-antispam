package bannedDB

import (
	"encoding/binary"
)

func UserIDToKey(userID int64) []byte {
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutVarint(buf, userID)
	return buf[:n]
}

package dataset

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
)

func HashMD5(s string) string {
	hasher := md5.New()
	hasher.Write([]byte(s))

	return hex.EncodeToString(hasher.Sum(nil))
}

func HashMapMD5(s map[string]string) string {
	data, _ := json.Marshal(s)
	hasher := md5.New()
	hasher.Write(data)

	return hex.EncodeToString(hasher.Sum(nil))
}

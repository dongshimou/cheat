package util

import (
	"crypto/md5"
	"fmt"
)

func EncryptPassword(origin string) (target string) {
	m5 := md5.Sum([]byte(origin))
	return fmt.Sprintf("%x", m5)
}


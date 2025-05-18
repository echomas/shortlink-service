package main

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

const shortCodeLength = 6 //短链接代码的长度

func generateRandomShortCode() (string, error) {
	bytes := make([]byte, shortCodeLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := base64.URLEncoding.EncodeToString(bytes)
	code = strings.ReplaceAll(code, "_", "a") //替换非字母数字字符
	code = strings.ReplaceAll(code, "-", "b")
	return code[:shortCodeLength], nil
}

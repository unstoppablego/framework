package tool

import (
	"errors"
	"fmt"
)

// Base62 字符集
const base62Chars = "klmn89abcdefghijstuvwxPQRSopqr0123yzABCDEFGHIJKLMNO4567TUVWXYZ"
const (
	base          = 62
	encodedLength = 6
	maxID         = 56800235583 // 62^6 - 1
)

// decodingTable 用于快速查找字符对应的值
var decodingTable [128]int8

func init() {
	// 初始化解码表
	for i := range decodingTable {
		decodingTable[i] = -1
	}
	for i, c := range base62Chars {
		decodingTable[c] = int8(i)
	}
}

// 混淆密钥（简单示例，实际应用中应使用更安全的方式管理密钥）
const secretKey = 0x5A

// 混淆 ID
func obfuscate(id int64) int64 {
	return id ^ int64(secretKey)
}

// 反混淆 ID
func deobfuscate(id int64) int64 {
	return id ^ int64(secretKey)
}

// 将 ID 转换为 6 位 Base62 字符串
func Base62EncodeID(id int64) (string, error) {
	if id < 0 || id > maxID {
		return "", fmt.Errorf("ID 超出范围，必须在 0 到 %d 之间", maxID)
	}

	// 混淆 ID
	obfuscatedID := obfuscate(id)

	// 编码过程
	encoded := make([]byte, 0, encodedLength)
	for obfuscatedID > 0 {
		remainder := obfuscatedID % base
		encoded = append(encoded, base62Chars[remainder])
		obfuscatedID /= base
	}

	// 填充前导 '0' 并确保长度为6
	for len(encoded) < encodedLength {
		encoded = append(encoded, base62Chars[0])
	}

	// 反转字节数组
	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	return string(encoded), nil
}

// 将 6 位 Base62 字符串转换回 ID
func Base62DecodeID(encoded string) (int64, error) {
	if len(encoded) != encodedLength {
		return 0, fmt.Errorf("编码字符串长度必须为 %d 位", encodedLength)
	}

	var id int64
	for i := 0; i < encodedLength; i++ {
		c := encoded[i]
		if c >= 128 || decodingTable[c] == -1 {
			return 0, fmt.Errorf("编码字符串包含无效字符: %c", c)
		}
		id = id*base + int64(decodingTable[c])
	}

	// 反混淆 ID
	originalID := deobfuscate(id)

	if originalID < 0 || originalID > maxID {
		return 0, errors.New("解码后的 ID 超出有效范围")
	}

	return originalID, nil
}

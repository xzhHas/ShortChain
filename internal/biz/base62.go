package biz

import (
	"math/rand"
	"strings"
	"time"
)

var (
	globalRand *rand.Rand
)

var characters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func init() {
	// 初始化全局随机数生成器
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	// 打乱字符顺序(在使用DB自增ID的时候，防止被猜出来)
	characters = shuffleString(characters)
}

// generateShortUrl 返回短链
func generateShortUrl(id int64) string {
	return base62Encode(id)
}

// base62Encode 将十进制数字转换为62进制
func base62Encode(decimal int64) string {
	var result strings.Builder
	for decimal > 0 {
		remainder := decimal % 62
		result.WriteByte(characters[remainder])
		decimal /= 62
	}
	return result.String()
}

// shuffleString 打乱字符顺序，避免被猜测到
func shuffleString(input string) string {
	// 洗牌算法，打乱字符串的字符顺序
	chars := strings.Split(input, "")
	for i := len(chars) - 1; i > 0; i-- {
		// 获取一个0-i之间的随机索引
		j := globalRand.Intn(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return strings.Join(chars, "")
}

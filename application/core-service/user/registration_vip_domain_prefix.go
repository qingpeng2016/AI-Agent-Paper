package user

import (
	"crypto/rand"
	"math/big"
)

// 单字常用义拼音（无声调，用作子域前缀）
var registrationPinyinWords = []string{
	"fu", "lu", "shou", "xi", "cai", "he", "ping", "an", "le", "meng",
	"xing", "yun", "ji", "xiang", "mei", "hao", "long", "feng", "shui", "shan",
	"hai", "tian", "ri", "yue", "jin", "yin", "yu", "long", "hu", "ma",
	"yang", "tu", "nian", "sui", "chun", "xia", "qiu", "dong", "hong", "zi",
	"lan", "qing", "ming", "liang", "wen", "ru", "jia", "ren", "yi", "xin",
	"de", "shan", "en", "hui", "cheng", "xin", "yong", "meng", "xiang", "wang",
	"da", "xiao", "gao", "yuan", "shen", "guang", "ming", "li", "hua", "kai",
}

const (
	vipDomainPinyinMaxTries = 10
	vipDomainLetterMaxTries = 10
	vipDomainRandomLen      = 6
)

func randomPinyinPrefix() (string, error) {
	if len(registrationPinyinWords) == 0 {
		return randomAlphaLower(vipDomainRandomLen)
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(registrationPinyinWords))))
	if err != nil {
		return "", err
	}
	return registrationPinyinWords[n.Int64()], nil
}

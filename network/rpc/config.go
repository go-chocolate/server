package rpc

import (
	"strconv"
	"strings"

	"github.com/go-chocolate/server/basic"
)

type ByteText string

func (b ByteText) Value() int {
	text := strings.ToLower(string(b))
	var num []byte
	var unit string
	for i, v := range text {
		if v >= '0' && v <= '9' {
			num = append(num, byte(v))
		} else {
			unit = text[i:]
			break
		}
	}
	if len(num) == 0 {
		return 0
	}
	val, _ := strconv.Atoi(string(num))
	switch unit {
	case "", "b":
		return val
	case "k", "kb":
		return val * 1024
	case "m", "mb":
		return val * 1024 * 1024
	case "g", "gb":
		return val * 1024 * 1024 * 1024
	case "t", "tb":
		return val * 1024 * 1024 * 1024 * 1024
	case "p", "pb":
		return val * 1024 * 1024 * 1024 * 1024 * 1024
	default:
		return 0
	}
}

type Config struct {
	basic.Config
	Timeout        string
	MaxRecvMsgSize ByteText
	MaxSendMsgSize ByteText
	Logger         LoggerConfig
}

type LoggerConfig struct {
	Enable bool
}

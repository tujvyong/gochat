package utils

import (
	"bytes"
	"fmt"

	"github.com/go-redis/redis"
)

func StrJoin(dist, src string) string {
	buffer := bytes.NewBuffer(make([]byte, 0))
	buffer.WriteString(dist)
	buffer.WriteString(src)
	return buffer.String()
}

func IsExist(haystack interface{}, needles interface{}) bool {
	switch haystack.(type) {
	case []string:
		channels := haystack.([]string)
		needles = needles.(string)
		for _, v := range channels {
			if needles == v {
				return true
			}
		}
		return false
	case []*redis.PubSub:
		channels := haystack.([]*redis.PubSub)
		needles = fmt.Sprintf("PubSub(%s)", needles.(string))
		for _, v := range channels {
			if needles == v.String() {
				return true
			}
		}
		return false
	default:
		return false
	}
}

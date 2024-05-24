package wx

import (
	r2redis "github.com/open4go/db/redis"
	"github.com/open4go/log"
	"github.com/redis/go-redis/v9"
)

// GetRedisCacheHandler 获取数据库handler 这里定义一个方法
func GetRedisCacheHandler() *redis.Client {
	handler, err := r2redis.DBPool.GetHandler("cache")
	if err != nil {
		log.Log().Fatal(err)
	}
	return handler
}

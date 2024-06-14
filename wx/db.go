package wx

import (
	"context"
	r2redis "github.com/open4go/db/redis"
	"github.com/open4go/log"
	"github.com/redis/go-redis/v9"
)

// GetRedisCacheHandler 获取数据库handler 这里定义一个方法
func GetRedisCacheHandler(ctx context.Context) *redis.Client {
	handler, err := r2redis.DBPool.GetHandler("cache")
	if err != nil {
		log.Log(ctx).Fatal(err)
	}
	return handler
}

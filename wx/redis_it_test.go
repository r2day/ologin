package wx

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	r2redis "github.com/open4go/db/redis"
	"github.com/r2day/db"
	"github.com/redis/go-redis/v9"
)

// 集成测试：仅在 R2DAY_IT=1 时运行，使用本地 Redis（默认 127.0.0.1:6379，可用 R2DAY_IT_REDIS 覆盖）的 db 15，
// 只删除本测试写入的 wxa_access_token_* key。
func itRedis(t *testing.T, keys ...string) (string, *redis.Client) {
	t.Helper()
	if os.Getenv("R2DAY_IT") != "1" {
		t.Skip("集成测试：设置 R2DAY_IT=1 并启动本地 Redis 后运行")
	}
	addr := os.Getenv("R2DAY_IT_REDIS")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	c := redis.NewClient(&redis.Options{Addr: addr, DB: 15})
	if err := c.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("无法连接 Redis: %v", err)
	}
	c.Del(context.Background(), keys...)
	prev := db.RDB
	db.RDB = c
	t.Cleanup(func() {
		c.Del(context.Background(), keys...)
		db.RDB = prev
		_ = c.Close()
	})
	return addr, c
}

func TestIT_GetWxaAccessTokenWithCache(t *testing.T) {
	cases := []struct {
		name      string
		cached    string
		body      string
		wantToken string
		wantErr   bool
		wantHits  int32
		// minTTL<0 时按 go-redis 的约定精确比较：-1 表示 key 不过期，-2 表示 key 不存在
		minTTL, maxTTL time.Duration
	}{
		{"命中缓存，不请求微信", "cached-token", "", "cached-token", false, 0, -1, -1},
		{"未命中：请求微信并缓存（提前 5 分钟过期）", "", `{"access_token":"tok","expires_in":7200}`, "tok", false, 1, 6890 * time.Second, 6900 * time.Second},
		// 现状：expires_in ≤ 300 时 TTL 计算为 0 或负数，go-redis 不设置过期，token 被永久缓存
		{"expires_in=300：永久缓存（现状）", "", `{"access_token":"tok300","expires_in":300}`, "tok300", false, 1, -1, -1},
		{"expires_in 缺失：永久缓存（现状）", "", `{"access_token":"tok0"}`, "tok0", false, 1, -1, -1},
		{"微信报错：不写缓存", "", `{"errcode":40013,"errmsg":"invalid appid"}`, "", true, 3, -2, -2},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			appID := fmt.Sprintf("%s_%d", fakeAppID, i)
			key := fmt.Sprintf(wxaAccessTokenKey, appID)
			ctx := context.Background()
			_, rdb := itRedis(t, key)
			if tc.cached != "" {
				rdb.Set(ctx, key, tc.cached, 0)
			}
			var hits int32
			fakeWeChat(t, func(w http.ResponseWriter, _ *http.Request) {
				atomic.AddInt32(&hits, 1)
				_, _ = io.WriteString(w, tc.body)
			})
			got, err := GetWxaAccessTokenWithCache(ctx, appID, fakeSecret)
			if (err != nil) != tc.wantErr || got != tc.wantToken {
				t.Fatalf("got=%q err=%v", got, err)
			}
			if h := atomic.LoadInt32(&hits); h != tc.wantHits {
				t.Fatalf("请求次数=%d", h)
			}
			ttl := rdb.TTL(ctx, key).Val()
			if tc.minTTL < 0 {
				if ttl != tc.minTTL {
					t.Fatalf("TTL=%v want %v", ttl, tc.minTTL)
				}
			} else if ttl < tc.minTTL || ttl > tc.maxTTL {
				t.Fatalf("TTL=%v 不在 [%v,%v]", ttl, tc.minTTL, tc.maxTTL)
			}
			if !tc.wantErr && rdb.Get(ctx, key).Val() != tc.wantToken {
				t.Fatal("缓存值不一致")
			}
		})
	}
}

func TestIT_GetRedisCacheHandler(t *testing.T) {
	addr, _ := itRedis(t)
	prev := r2redis.DBPool
	t.Cleanup(func() { r2redis.DBPool = prev })
	r2redis.NewDataBasePool()
	want, err := r2redis.DBPool.GetClient(context.Background(), addr, "cache", 15)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(r2redis.DBPool.CloseAll)
	if got := GetRedisCacheHandler(context.Background()); got != want {
		t.Fatalf("got=%p want=%p", got, want)
	}
}

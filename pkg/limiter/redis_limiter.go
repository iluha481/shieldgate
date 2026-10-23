package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var slidingWindowScript = redis.NewScript(`
	local key = KEYS[1]
	local now = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])
	local limit = tonumber(ARGV[3])

	local clearBefore = now - window
	redis.call('ZREMRANGEBYSCORE', key, '-inf', clearBefore)

	local currentRequests = redis.call('ZCARD', key)

	if currentRequests < limit then
		redis.call('ZADD', key, now, now)
		redis.call('PEXPIRE', key, window)
		return 1
	else
		return 0
	end
`)

type RedisLimiter struct {
	rdb *redis.Client
}

func NewRedisLimiter(rdb *redis.Client) *RedisLimiter {
	return &RedisLimiter{rdb: rdb}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, limit int) (bool, error) {
	if limit <= 0 {
		return true, nil
	}

	redisKey := fmt.Sprintf("rate_limit:%s", key)
	now := time.Now().UnixMilli()
	window := 1000

	res, err := slidingWindowScript.Run(ctx, l.rdb, []string{redisKey}, now, window, limit).Int()
	if err != nil {
		return false, fmt.Errorf("redis limiter failed: %w", err)
	}

	return res == 1, nil
}

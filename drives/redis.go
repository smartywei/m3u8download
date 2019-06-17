package drives

import (
	"github.com/go-redis/redis"
	"M3u8Tool/config"
	"time"
)

func GetConn() *redis.Client {

	redisHost := config.CFG.Section("redis").Key("Host").MustString("127.0.0.1")
	redisPort := config.CFG.Section("redis").Key("Port").MustString("6379")
	redisPassword := config.CFG.Section("redis").Key("Password").MustString("")
	redisDB := config.CFG.Section("redis").Key("DB").MustInt(0)

	return redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword, // no password set
		DB:       redisDB,       // use default DB
	})
}

func RedisSetKeyValue(key string, val interface{}, expiration time.Duration) error {
	coon := GetConn()
	defer coon.Close()
	return coon.Set(key, val, expiration).Err()
}

func RedisGetKeyValue(key string) (string, error) {
	coon := GetConn()
	defer coon.Close()
	return coon.Get(key).Result()
}

func RedisLPush(key string, values interface{}) error {
	coon := GetConn()
	defer coon.Close()
	return coon.LPush(key, values).Err()
}

func RedisRPop(key string) (string, error) {
	coon := GetConn()
	defer coon.Close()
	return coon.RPop(key).Result()
}
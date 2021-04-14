package cache

import (
	"log"
	"time"
	"v2ray-admin/backend/conf"
	"v2ray-admin/backend/redis"
)

type ICache interface {
	Get(key string) string
	Exist(key string) bool
	Put(key string, value string, expire int)
	Evict(key string)
}

var Cache ICache

func init() {
	log.Println("初始化缓存...")
	c := conf.App.Cache
	switch c.Manager {
	case "redis":
		Cache = RedisCacheManager()
	default:
		Cache = MemoryCacheManager()
	}
	log.Println("缓存初始化完成")
}

// redis cache
type RedisCache struct {
}

func (rc RedisCache) Get(key string) string {
	v, err := redis.Get(key)
	if err != nil {
		log.Panicln("cache[redis]:", err)
	}
	return v
}
func (rc RedisCache) Put(key string, value string, expire int) {
	err := redis.Set(key, value, expire)
	if err != nil {
		log.Panicln("cache[redis]:", err)
	}
}
func (rc RedisCache) Evict(key string) {
	err := redis.Del(key)
	if err != nil {
		log.Panicln("cache[redis]:", err)
	}
}
func (rc RedisCache) Exist(key string) bool {
	exist, err := redis.Exist(key)
	if err != nil {
		log.Panicln("cache[redis]:", err)
	}
	return exist
}
func RedisCacheManager() ICache {
	return &RedisCache{}
}

// memory
type MemoryCache struct {
	m  map[string]string
	af map[string]*time.Timer
}

func (mc MemoryCache) Get(key string) string {
	return mc.m[key]
}
func (mc MemoryCache) Put(key string, value string, expire int) {
	existTimer := mc.af[key]
	if existTimer != nil {
		mc.Evict(key)
		existTimer.Stop()
	}

	mc.m[key] = value
	// expire
	t := time.AfterFunc(time.Second*time.Duration(expire), func() {
		mc.Evict(key)
	})
	mc.af[key] = t
}
func (mc MemoryCache) Evict(key string) {
	delete(mc.m, key)
	delete(mc.af, key)
}
func (mc MemoryCache) Exist(key string) bool {
	_, ok := mc.m[key]
	return ok
}
func MemoryCacheManager() ICache {
	c := &MemoryCache{}
	c.m = make(map[string]string)
	c.af = make(map[string]*time.Timer)
	return c
}

package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"log"
	"v2ray-admin/backend/conf"
)

var rPool *redis.Pool

func init() {
	log.Println("初始化Redis...")

	c := conf.App.Redis
	if c.Enable {
		rPool = &redis.Pool{
			MaxIdle:   c.PoolMaxIdle,
			MaxActive: c.PollMaxActive,
			Dial: func() (redis.Conn, error) {
				conn, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))

				// Connection error handling
				if err != nil {
					log.Fatalf("redis: fail initializing the redis pool: %s \n", err.Error())
				}

				return conn, err
			},
		}

		// test connect
		_ = ping(rPool.Get())

		log.Println("Redis初始化完成")
	} else {
		log.Println("Redis未启用")
	}
}

func ping(c redis.Conn) error {
	s, err := redis.String(c.Do("PING"))
	if err != nil {
		return err
	}

	log.Println("redis: PING Response = " + s)

	return nil
}

func Set(key string, value string, expire int) error {
	c := rPool.Get()
	defer c.Close()

	_, err := c.Do("SET", key, value)
	if err != nil {
		return err
	}

	// expire
	_, err = c.Do("EXPIRE", key, expire)
	if err != nil {
		return err
	}

	return nil
}

func Get(key string) (value string, err error) {
	c := rPool.Get()
	defer c.Close()

	s, err := redis.String(c.Do("GET", key))
	if err == redis.ErrNil {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return s, nil
	}
}

func Del(key string) error {
	c := rPool.Get()
	defer c.Close()

	_, err := c.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}

func Exist(key string) (exist bool, err error) {
	c := rPool.Get()
	defer c.Close()

	val, err := c.Do("EXISTS", key)
	if err != nil {
		return false, err
	}

	return val.(int64) == 1, nil
}

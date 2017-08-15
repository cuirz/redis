package redis

import (
	"fmt"
	"time"
	"github.com/Unknwon/com"
	"gopkg.in/redis.v5"
	"strings"
)

type RedisCache struct {
	c               *redis.Client
	Prefix          string
	DefaultHsetName string
}

func (c *RedisCache) Expire(key string, duration time.Duration) error {
	key = c.Prefix + key
	return c.c.Expire(key, duration).Err()
}

func (c *RedisCache) HSetWithDuration(key string, array map[string]interface{}, duration time.Duration) error {
	err := c.HSet(key, array)
	if err != nil {
		return err
	}
	return c.Expire(key, duration)
}

func (c *RedisCache) HSet(key string, array map[string]interface{}) error {
	ks := strings.Split(key, ":")
	hset_name := c.DefaultHsetName
	if len(ks) > 1 {
		hset_name = ks[0]
	}
	key = c.Prefix + key

	var cmd *redis.BoolCmd
	_, err := c.c.TxPipelined(func(pipe *redis.Pipeline) error {
		cmd = pipe.HSet(hset_name, key, "0")
		if cmd.Err() != nil {
			return cmd.Err()
		}

		for k, v := range array {
			if err := pipe.HSet(key, k, v).Err(); err != nil {
				return err
			}
		}
		//return fmt.Errorf("错误了吧")
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return err
}

func (c *RedisCache) HGetAll(key string) interface{} {
	key = c.Prefix + key
	val, err := c.c.HGetAll(key).Result()
	if err != nil {
		return nil
	}
	return val
}

func (c *RedisCache) HGet(key string, field string) interface{} {
	key = c.Prefix + key
	val, err := c.c.HGet(key, field).Result()
	if err != nil {
		return nil
	}
	return val
}

func (c *RedisCache) GetField(prefix string, key string, field string) interface{} {
	key = c.Prefix + prefix + ":" + key
	val, err := c.c.HGet(key, field).Result()
	if err != nil {
		return nil
	}
	return val
}

// If expired is 0, it lives forever.
func (c *RedisCache) Set(key string, val interface{}, expire time.Duration) error {
	ks := strings.Split(key, ":")
	hset_name := c.DefaultHsetName
	if len(ks) > 1 {
		hset_name = ks[0]
	}
	key = c.Prefix + key
	if err := c.c.Set(key, com.ToStr(val), expire).Err(); err != nil {
		return err
	}

	return c.c.HSet(hset_name, key, "0").Err()
}

func (c *RedisCache) Get(key string) interface{} {
	val, err := c.c.Get(c.Prefix + key).Result()
	if err != nil {
		return nil
	}
	return val
}

func (c *RedisCache) Delete(key string) error {
	key = c.Prefix + key
	if err := c.c.Del(key).Err(); err != nil {
		return err
	}

	ks := strings.Split(key, ":")
	hset_name := c.DefaultHsetName
	if len(ks) > 1 {
		hset_name = ks[0]
	}
	return c.c.HDel(hset_name, key).Err()
}

func (c *RedisCache) Incr(key string) error {
	if !c.IsExist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}
	return c.c.Incr(c.Prefix + key).Err()
}

func (c *RedisCache) Decr(key string) error {
	if !c.IsExist(key) {
		return fmt.Errorf("key '%s' not exist", key)
	}
	return c.c.Decr(c.Prefix + key).Err()
}

// IsExist returns true if cached value exists.
func (c *RedisCache) IsExist(key string) bool {
	if c.c.Exists(c.Prefix + key).Val() {
		return true
	}

	ks := strings.Split(key, ":")
	hset_name := c.DefaultHsetName
	if len(ks) > 1 {
		hset_name = ks[0]
	}
	c.c.HDel(hset_name, c.Prefix+key)
	return false
}

// Flush deletes all cached data.
func (c *RedisCache) Flush(hsetname string) error {
	if strings.EqualFold(hsetname, "") {
		hsetname = c.DefaultHsetName
	}

	keys, err := c.c.HKeys(hsetname).Result()
	if err != nil {
		return err
	}
	if err = c.c.Del(keys...).Err(); err != nil {
		return err
	}
	return c.c.Del(hsetname).Err()
}

func (c *RedisCache) FlushDB() error {
	return c.c.FlushDb().Err()
}

// StartAndGC starts GC routine based on config string settings.
func (c *RedisCache) StartAndGC(option map[string]string) (err error) {
	c.DefaultHsetName = "RedisCache"

	opt := &redis.Options{
		Network: "tcp",
	}
	for k, v := range option {
		switch k {
		case "network":
			opt.Network = v
		case "host":
			opt.Addr = v
		case "password":
			opt.Password = v
		case "db":
			opt.DB = com.StrTo(v).MustInt()
		case "pool_size":
			opt.PoolSize = com.StrTo(v).MustInt()
		case "idle_timeout":
			opt.IdleTimeout, err = time.ParseDuration(v + "s")
			if err != nil {
				return fmt.Errorf("error parsing idle timeout: %v", err)
			}
		case "hset_name":
			c.DefaultHsetName = v
		case "prefix":
			c.Prefix = v
		default:
			return fmt.Errorf("redis: unsupported option '%s'", k)
		}
	}

	c.c = redis.NewClient(opt)
	if err = c.c.Ping().Err(); err != nil {
		return err
	}

	return nil
}

func Init(opt map[string]string) (*RedisCache, error) {
	r := &RedisCache{}
	err := r.StartAndGC(opt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

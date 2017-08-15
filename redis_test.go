package redis

import (
	"testing"
	"fmt"
)

//
//var data = map[string]string{
//	"host":         "192.168.1.204:6379",
//	"password":     "redis666",
//	"db":           "1",
//	"network":      "tcp",
//	"pool_size":    "100",
//	"idle_timeout": "180",
//	"prefix":       "charge:",
//	"hset_name":    "Agents",
//}

func benchmarkRedisClient(poolSize string) *RedisCache {
	data := map[string]string{
		"host":         "192.168.1.204:6379",
		"password":     "redis666",
		"db":           "1",
		"network":      "tcp",
		"pool_size":    poolSize,
		"idle_timeout": "180",
		"prefix":       "charge:",
		"hset_name":    "Agents",
	}

	redis, _ := Init(data)

	//if err := redis.FlushDB(); err != nil {
	//	panic(err)
	//}
	return redis
}

func TestInit(t *testing.T) {
	//data := map[string]string{
	//	"host":         "192.168.1.204:6379",
	//	"password":     "redis666",
	//	"db":           "1",
	//	"network":      "tcp",
	//	"pool_size":    "100",
	//	"idle_timeout": "180",
	//	"prefix":       "charge:",
	//	"hset_name":    "Agents",
	//}

	redis := benchmarkRedisClient("100")
	err := redis.HSet("act:10001", map[string]interface{}{
		//"parent": "papa",
		"name": "10001",
	})
	fmt.Println(err)

	v := redis.HGetAll("act:10001")
	fmt.Println(v)

	v = redis.HGet("act:10001","name")
	fmt.Println(v)

}

func BenchmarkInit(b *testing.B) {
	redis := benchmarkRedisClient("100")

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := redis.HSet("act:user", map[string]interface{}{
				"name": "cui",
			}); err != nil {
				b.Fatal(err)
			}
		}
	})

}

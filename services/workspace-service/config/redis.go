package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(uri string) *redis.Client {
	opt, err := redis.ParseURL(uri)
	if err != nil {
		fmt.Printf("Invalid Redis URI: %v\n", err)
		return nil
	}

	rdb := redis.NewClient(opt)

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("Could not connect to Redis: %v\n", err)
		return nil
	}

	fmt.Println("Connected to Redis successfully via URI")
	return rdb
}

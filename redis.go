package main

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

var url = os.Getenv("REDIS_URL")

var rdb = redis.NewClient(&redis.Options{
	Addr: url, // Adjust the address as needed
})

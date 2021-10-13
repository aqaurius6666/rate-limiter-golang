package main

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisURI string
type RedisPassword string
type RedisUsername string

type RedisService struct {
	ctx    context.Context
	client *redis.Client
	logger *logrus.Logger
}

func NewRedisService(ctx context.Context, logger *logrus.Logger, uri RedisURI, pass RedisPassword, username RedisUsername) (*RedisService, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     string(uri),
		DB:       0,
		Username: string(username),
		Password: string(pass),
	})
	return &RedisService{
		ctx:    ctx,
		client: rdb,
		logger: logger,
	}, nil
}

func (r *RedisService) Close() error {
	return r.client.Close()
}

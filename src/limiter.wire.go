//go:build wireinject
// +build wireinject

package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
)

type RedisOpts struct {
	Uri      RedisURI
	Pass     RedisPassword
	Username RedisUsername
	Remote   RemoteURL
}

func InitializeLimiterServiceServer(ctx context.Context, logger *logrus.Logger, opts RedisOpts) (*LimiterServiceServer, error) {
	wire.Build(
		wire.FieldsOf(&opts, "Uri", "Pass", "Username", "Remote"),
		NewRedisService,
		gin.Default,
		wire.Struct(new(LimiterServiceServer), "*"),
	)
	return &LimiterServiceServer{}, nil
}

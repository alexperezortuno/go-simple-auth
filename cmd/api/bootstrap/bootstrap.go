package bootstrap

import (
	"context"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/internal/platform/server"
)

var conf = config.GetConfigs()

func Run() error {
	ctx, srv := server.New(context.Background(), conf)
	return srv.Run(ctx)
}

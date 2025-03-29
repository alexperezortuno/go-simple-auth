package bootstrap

import (
	"context"
	"github.com/alexperezortuno/go-simple-auth/infra/postgres"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/internal/platform/server"
	"log"
	"os"
)

var conf = config.GetConfigs()

func Run() error {
	// Modo para crear usuario desde terminal
	if len(os.Args) >= 3 && os.Args[1] == "create-user" {
		log.Printf("creating user %s with password %s migrate is active %s", os.Args[2], os.Args[3], os.Args[4])
		if os.Args[4] == "true" {
			config.SetEnvBool("MIGRATE", os.Args[4])
		}
		postgres.InitDatabase(conf)
		postgres.CreateUser(os.Args[2], os.Args[3], conf)
		os.Exit(0)
	}
	ctx, srv := server.New(context.Background(), conf)
	return srv.Run(ctx)
}

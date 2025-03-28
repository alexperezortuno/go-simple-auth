package main

import (
	"github.com/alexperezortuno/go-simple-auth/cmd/api/bootstrap"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"log"
	"os"
)

func main() {
	// Modo para crear usuario desde terminal
	if len(os.Args) >= 3 && os.Args[1] == "create-user" {
		m := config.GetEnvBool(os.Args[4], false)
		log.Printf("creating user %s with password %s migrate is active %s", os.Args[2], os.Args[3], os.Args[4])
		if m {
			initDatabase(m)
		}
		createUser(os.Args[2], os.Args[3], m)
		return
	}

	if err := bootstrap.Run(); err != nil {
		log.Fatal(err)
	}
}

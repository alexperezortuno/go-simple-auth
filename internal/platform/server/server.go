package server

import (
	"context"
	"fmt"
	handler "github.com/alexperezortuno/go-simple-auth/delivery/http"
	"github.com/alexperezortuno/go-simple-auth/infra/postgres"
	"github.com/alexperezortuno/go-simple-auth/internal/config"
	"github.com/alexperezortuno/go-simple-auth/middleware"
	"github.com/alexperezortuno/go-simple-auth/usecase"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type Server struct {
	httpAddr        string
	engine          *gin.Engine
	shutdownTimeout time.Duration
}

func New(ctx context.Context, conf config.Config) (context.Context, Server) {
	srv := Server{
		engine:          gin.New(),
		httpAddr:        fmt.Sprintf("%s:%d", conf.Host, conf.Port),
		shutdownTimeout: time.Duration(conf.ShutdownTimeout) * time.Second,
	}

	log.Println(fmt.Sprintf("Check app in %s:%d/%s/%s", conf.Host, conf.Port, conf.SubDir, "health"))
	srv.registerRoutes(conf)
	return serverContext(ctx), srv
}

func (s *Server) registerRoutes(conf config.Config) {
	s.engine.Use(gzip.Gzip(gzip.DefaultCompression))
	s.engine.Use(gzip.Gzip(gzip.BestSpeed))
	// Añadir el middleware de registro de tiempo de respuesta
	s.engine.Use(middleware.RequestLogger())
	s.engine.Use(middleware.RateLimit())
	s.engine.Use(middleware.CustomRecovery())
	s.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	service := &usecase.UserService{}

	if conf.DBEngine == "postgres" {
		postgres.Initialize(conf)
		defer postgres.CloseConnection()
		postgres.Migrate(conf)

		// Repositorios
		repo := postgres.NewUserRepository(postgres.Connection)
		service = usecase.NewUserService(repo)
	}

	// Handlers
	handler.NewLoginHandler(s.engine, *service, conf)
	s.engine.GET(fmt.Sprintf("%s/health", conf.SubDir), func(c *gin.Context) { // Endpoint para verificar la salud del servidor
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Rutas protegidas
	protected := s.engine.Group(conf.SubDir)
	protected.Use(middleware.Auth())
	protected.POST("/renew", handler.RenewToken)       // Renovar token existente
	protected.POST("/validate", handler.ValidateToken) // Validar token
}

func (s *Server) Run(ctx context.Context) error {
	log.Println("Server running on", s.httpAddr)
	srv := &http.Server{
		Addr:    s.httpAddr,
		Handler: s.engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server shut down", err)
		}
	}()

	<-ctx.Done()
	ctxShutDown, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return srv.Shutdown(ctxShutDown)
}

func serverContext(ctx context.Context) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-c
		cancel()
	}()

	return ctx
}

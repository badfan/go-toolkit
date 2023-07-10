package webserver

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type Server struct {
	httpServer *http.Server
}

func (s *Server) Run(handler http.Handler, addr string) error {
	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func NewRouter() *gin.Engine {
	gin.SetMode(viper.GetString("gin_mode"))
	router := gin.New()
	router.Use(otelgin.Middleware(viper.GetString("service_name")))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}

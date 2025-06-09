package apiServer

import (
	"context"
	"splitExpense/config"
	"splitExpense/orchestrator"

	"github.com/gin-gonic/gin"
)

func Start() {
	opts := gin.OptionFunc(func(e *gin.Engine) {
		e.RedirectFixedPath = true
		e.RedirectTrailingSlash = true
	})
	r := gin.Default(opts)

	cfg := &config.Config{
		// Fill with actual config values or load from env
		DatabaseHost:     "localhost",
		DatabasePort:     "5432",
		DatabaseUser:     "postgres",
		DatabasePassword: "password",
		DatabaseName:     "splitexpense",
		DatabaseSSLMode:  "disable",
		Environment:      config.EnvironmentDevelopment,
	}
	ctx := context.Background()

	// Only create orchestrator, let it handle service dependencies internally
	app := orchestrator.NewExpenseApp(ctx, cfg)

	attachRoutes(r, app, cfg)

	r.Run()
}

type Method string

const (
	GET    Method = "GET"
	POST   Method = "POST"
	PUT    Method = "PUT"
	DELETE Method = "DELETE"
	HEAD   Method = "HEAD"
)

type RouteHandler interface {
	Method() Method
	Path() string
	Handle(g *gin.Context, c *config.Config)
}

func attachRoutes(g *gin.Engine, orchestrator orchestrator.ExpenseAppImpl, cfg *config.Config) {
	handlers := []RouteHandler{
		&LoginRouteHandler{orchestrator: orchestrator},
		// Add more handlers here
	}
	validateAndHandle := func(h RouteHandler) gin.HandlerFunc {
		return func(c *gin.Context) {
			h.Handle(c, cfg)
		}
	}

	for _, h := range handlers {
		switch h.Method() {
		case GET:
			g.GET(h.Path(), validateAndHandle(h))
		case POST:
			g.POST(h.Path(), validateAndHandle(h))
		case PUT:
			g.PUT(h.Path(), validateAndHandle(h))
		case DELETE:
			g.DELETE(h.Path(), validateAndHandle(h))
		case HEAD:
			g.HEAD(h.Path(), validateAndHandle(h))
		}
	}
}

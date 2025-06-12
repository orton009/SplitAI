package apiServer

import (
	"context"
	"splitExpense/config"
	"splitExpense/orchestrator"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Start() {
	opts := gin.OptionFunc(func(e *gin.Engine) {
		e.RedirectFixedPath = true
		e.RedirectTrailingSlash = true
	})
	r := gin.Default(opts)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	cfg := &config.Config{
		// Fill with actual config values or load from env
		DatabaseHost:     "localhost",
		DatabasePort:     "5432",
		DatabaseUser:     "postgres",
		DatabasePassword: "postgres",
		DatabaseName:     "postgres",
		DatabaseSSLMode:  "disable",
		Environment:      config.EnvironmentDevelopment,
	}
	ctx := context.Background()

	// Only create orchestrator, let it handle service dependencies internally
	app := orchestrator.NewExpenseApp(ctx, cfg)

	attachRoutes(r, app, cfg)

	r.Run(":8888")
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

type ApiHandler struct {
	handle       RouteHandler
	PreHandlers  []gin.HandlerFunc
	PostHandlers []gin.HandlerFunc
}

func routeMap(o orchestrator.ExpenseAppImpl) []ApiHandler {
	return []ApiHandler{
		{
			handle: &LoginRouteHandler{orchestrator: o},
		},
		{
			handle: &userSignUpHandler{orchestrator: o},
		},
		{
			handle:      &DeleteExpenseRouteHandler{orchestrator: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
		{
			handle:      &CreateGroupRouteHandler{orchestrator: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
		{
			handle:      &CreateOrUpdateExpenseRouteHandler{orchestrator: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
		{
			handle:      &SettleExpenseHandler{orchestrator: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
		{
			handle:      &JoinGroupRouteHandler{orchestrator: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
		{
			handle:      &AddFriendHandler{o: o},
			PreHandlers: []gin.HandlerFunc{Authenticate},
		},
	}
}

func attachRoutes(g *gin.Engine, orchestrator orchestrator.ExpenseAppImpl, cfg *config.Config) {

	handlers := routeMap(orchestrator)

	for _, h := range handlers {

		routeHandler := func(r RouteHandler) gin.HandlerFunc {
			return func(c *gin.Context) {
				r.Handle(c, cfg)
			}
		}
		handlers := []gin.HandlerFunc{}
		handlers = append(handlers, h.PreHandlers...)
		handlers = append(handlers, routeHandler(h.handle))
		handlers = append(handlers, h.PostHandlers...)

		switch h.handle.Method() {
		case GET:
			g.GET(h.handle.Path(), handlers...)
		case POST:
			g.POST(h.handle.Path(), handlers...)
		case PUT:
			g.PUT(h.handle.Path(), handlers...)
		case DELETE:
			g.DELETE(h.handle.Path(), handlers...)
		case HEAD:
			g.HEAD(h.handle.Path(), handlers...)
		}
	}
}

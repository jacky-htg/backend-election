package route

import (
	_ "backend-election/docs"
	"backend-election/internal/handler"
	"backend-election/internal/middleware"
	"backend-election/internal/pkg/database"
	"backend-election/internal/pkg/logger"
	"backend-election/internal/pkg/redis"
	"fmt"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	httpSwagger "github.com/swaggo/http-swagger"
)

func ApiRoute(log *logger.Logger, db *database.Database, cache *redis.Cache) *httprouter.Router {
	router := httprouter.New()
	router.ServeFiles("/docs/*filepath", http.Dir("./docs"))

	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("%s:%s/docs/swagger.json", os.Getenv("APP_HOST"), os.Getenv("APP_PORT"))),
	)
	router.Handler("GET", "/swagger/*filepath", swaggerHandler)

	var mid middleware.Middleware = middleware.Middleware{Log: log, DB: db.Conn, Cache: cache}
	publicMiddlewares := []func(httprouter.Handle) httprouter.Handle{
		mid.CORS,
		mid.PanicRecovery,
		mid.Semaphore,
		mid.RateLimit,
		mid.Idempotency,
	}
	privateMiddlewares := append(publicMiddlewares, mid.Authentication, mid.Authorization)

	pemilihHandler := handler.Pemilihs{Log: log, DB: db.Conn, Cache: cache}
	userHandler := handler.Users{Log: log, DB: db.Conn, Cache: cache}
	authHandler := handler.Auths{Log: log, DB: db.Conn}

	router.POST("/login", mid.WrapMiddleware(publicMiddlewares, authHandler.Login))
	router.GET("/users", mid.WrapMiddleware(privateMiddlewares, userHandler.List))
	router.GET("/users/:id", mid.WrapMiddleware(privateMiddlewares, userHandler.GetById))
	router.POST("/users", mid.WrapMiddleware(privateMiddlewares, userHandler.Create))
	router.PUT("/users/:id", mid.WrapMiddleware(privateMiddlewares, userHandler.Update))
	router.DELETE("/users/:id", mid.WrapMiddleware(privateMiddlewares, userHandler.Delete))

	// Pemilih
	router.POST("/pemilih", mid.WrapMiddleware(privateMiddlewares, pemilihHandler.GetByPemilihId))

	return router
}

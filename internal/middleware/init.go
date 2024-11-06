package middleware

import (
	"backend-election/internal/pkg/logger"
	"backend-election/internal/pkg/redis"
	"database/sql"

	"github.com/julienschmidt/httprouter"
)

type Middleware struct {
	Log   *logger.Logger
	DB    *sql.DB
	Cache *redis.Cache
}

func (m *Middleware) WrapMiddleware(mw []func(httprouter.Handle) httprouter.Handle, handler httprouter.Handle) httprouter.Handle {

	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}

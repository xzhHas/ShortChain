package server

import (
	"github.com/BitofferHub/shortUrlX/internal/conf"
	"github.com/BitofferHub/shortUrlX/internal/interfaces"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, h *interfaces.Handler) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			MiddlewareTraceID(),
			MiddlewareLog(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	// 参数是handler接口，但是gin.Engine实现了ServerHttp接口 所以可以使用
	srv.HandlePrefix("/", interfaces.NewRouter(h))
	return srv
}

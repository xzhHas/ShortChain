package server

import (
	"context"
	"fmt"
	"time"

	"github.com/BitofferHub/pkg/constant"
	log "github.com/BitofferHub/pkg/middlewares/log"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

func MiddlewareTraceID() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			fmt.Printf("ctx %v\n", ctx)
			// 元数据，获取上下文信息，令牌认真，ID追踪
			if md, ok := metadata.FromServerContext(ctx); ok {
				traceID := md.Get(fmt.Sprintf("x-md-global-%s", constant.TraceID))
				ctx = context.WithValue(ctx, constant.TraceID, traceID)
				log.InfoContextf(ctx, "traceID %v", traceID)
			}
			return handler(ctx, req)
		}
	}
}

func MiddlewareLog() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			log.InfoContextf(ctx,
				"component:%s,operation:%s,args:%s,code:%d,reason:%s",
				kind,
				operation,
				req,
				code,
				reason,
			)
			begin := time.Now()
			reply, err = handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			log.InfoContextf(ctx, "cost %v", time.Since(begin))
			log.InfoContextf(ctx, "reply %v", reply)

			return
		}
	}
}

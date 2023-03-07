package commons

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func MidlewareLogger(logger *log.Logger, filteredServers ...any) grpc.UnaryServerInterceptor {
	filterLookup := make(map[any]struct{})
	for _, srv := range filteredServers {
		filterLookup[srv] = struct{}{}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if _, ok := filterLookup[info.Server]; ok {
			return handler(ctx, req)
		}
		start := time.Now()
		data, err := handler(ctx, req)
		respStatus := status.New(codes.OK, "ok")
		if err != nil {
			respStatus, _ = status.FromError(err)
		}
		logger.Printf("%s %v %s", info.FullMethod, respStatus.Code(), time.Since(start))
		return data, err
	}
}

func MidlewareRecover(logger *log.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Println("recovered from panic", r)
				logger.Println(string(debug.Stack()))
				err = status.Errorf(codes.Internal, "an unexpected error occurred")
			}
		}()
		return handler(ctx, req)
	}
}

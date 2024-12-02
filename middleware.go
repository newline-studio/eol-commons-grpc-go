package commons

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleLogging(logger *slog.Logger, filteredServers ...any) func(srv any, method string, handler func() error) error {
	filterLookup := make(map[any]struct{})
	for _, srv := range filteredServers {
		filterLookup[srv] = struct{}{}
	}

	return func(srv any, method string, handler func() error) error {
		if _, ok := filterLookup[srv]; ok {
			return handler()
		}
		start := time.Now()
		err := handler()
		respStatus := status.New(codes.OK, "ok")
		if err != nil {
			respStatus, _ = status.FromError(err)
		}
		logger.Info(
			"handle rpc",
			"method", method,
			"code", respStatus.Code().String(),
			"duration", time.Since(start).String(),
		)
		return err
	}
}

func UnaryMiddlewareLogger(logger *slog.Logger, filteredServers ...any) grpc.UnaryServerInterceptor {
	loggingHandler := handleLogging(logger, filteredServers...)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var data any
		err := loggingHandler(info.Server, info.FullMethod, func() error {
			handlerRes, err := handler(ctx, req)
			data = handlerRes
			if err != nil {
				return err
			}
			return nil
		})
		return data, err
	}
}

func StreamMiddlewareLogger(logger *slog.Logger, filteredServers ...any) grpc.StreamServerInterceptor {
	loggingHandler := handleLogging(logger, filteredServers...)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return loggingHandler(srv, info.FullMethod, func() error {
			return handler(srv, ss)
		})
	}
}

func handleRecover(logger *slog.Logger) func(handler func() error) error {
	return func(handler func() error) error {
		defer func() {
			if r := recover(); r != nil {
				logger.Error(
					"recovered from panic",
					"error", r,
					"stack", string(debug.Stack()),
				)
			}
		}()
		return handler()
	}
}

func UnaryMiddlewareRecover(logger *slog.Logger) grpc.UnaryServerInterceptor {
	recoverHandler := handleRecover(logger)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var data any
		err := recoverHandler(func() error {
			handlerRes, err := handler(ctx, req)
			data = handlerRes
			if err != nil {
				return err
			}
			return nil
		})
		return data, err
	}
}

func StreamMiddlewareRecover(logger *slog.Logger) grpc.StreamServerInterceptor {
	recoverHandler := handleRecover(logger)
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return recoverHandler(func() error {
			return handler(srv, ss)
		})
	}
}

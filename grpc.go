package commons

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type grpcCall[I any, O any] func(ctx context.Context, in I, opts ...grpc.CallOption) (O, error)

func MakeGrpc[T any, U any](
	logger *log.Logger,
	ctx context.Context,
	call grpcCall[T, U],
	in T,
	opts ...grpc.CallOption,
) (U, error) {
	res, err := call(ctx, in, opts...)
	if err != nil {
		logGrpcError(logger, err, call)
	}
	return res, err
}

func MakeGrpcWithTimeout[T any, U any](
	logger *log.Logger,
	ctx context.Context,
	timeoutDuration time.Duration,
	call grpcCall[T, U],
	in T,
	opts ...grpc.CallOption,
) (U, error) {
	reqCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()
	return MakeGrpc(logger, reqCtx, call, in, opts...)
}

func logGrpcError(logger *log.Logger, err error, call any) {
	if errStatus, ok := status.FromError(err); ok {
		switch errStatus.Code() {
		case codes.DeadlineExceeded,
			codes.PermissionDenied,
			codes.Unimplemented,
			codes.Internal,
			codes.Unavailable:
			logger.Println(
				fmt.Sprintf(
					"gRPC request to %s encountered unexected error: (%s) %s",
					runtime.FuncForPC(reflect.ValueOf(call).Pointer()).Name(),
					errStatus.Code().String(),
					errStatus.Message(),
				),
			)
		}
	}
}

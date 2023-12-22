package interceptor

import (
	"context"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			logger := logrus.WithContext(ctx)
			if recoverErr := recover(); recoverErr != nil {
				logger.WithField("method", info.FullMethod).
					Errorf("[PANIC]: %v", recoverErr)
				logger.Error(string(debug.Stack()))
				err = status.Error(codes.Unavailable, "service is unavailable currently")
			}
		}()
		return handler(ctx, req)
	}
}

func RecoveryStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			logger := logrus.WithContext(ss.Context())
			if recoverErr := recover(); recoverErr != nil {
				logger.WithField("method", info.FullMethod).Errorf("[PANIC]: %v", recoverErr)
				logger.Error(string(debug.Stack()))
				err = status.Error(codes.Unavailable, "service is unavailable currently")
			}
		}()
		return handler(srv, ss)
	}
}

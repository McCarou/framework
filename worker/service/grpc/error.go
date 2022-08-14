package grpc

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ErrorHandlerGrpc(ctx context.Context, p interface{}) (err error) {
	logrus.Error("internal server error - %s", p)
	return status.Error(codes.Internal, "internal server error")
}

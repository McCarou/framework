package grpc

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/radianteam/framework/worker"
	"google.golang.org/grpc"
)

type ServiceGrpc struct {
	*worker.WorkerBase

	grpcServer *grpc.Server

	services []func(s *grpc.Server, wc *worker.WorkerContexts)
}

func NewServiceGrpc(config *worker.WorkerConfig) *ServiceGrpc {
	return &ServiceGrpc{WorkerBase: worker.NewWorkerBase(config)}
}

func (w *ServiceGrpc) AddService(f func(s *grpc.Server, wc *worker.WorkerContexts)) {
	w.services = append(w.services, f)
}

func (w *ServiceGrpc) Setup() {
	w.Logger.Infof("Setting up GRPC Service")

	opts := []grpc_logrus.Option{}

	grpc_logrus.ReplaceGrpcLogger(w.Logger)

	w.grpcServer = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.UnaryServerInterceptor(w.Logger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logrus.StreamServerInterceptor(w.Logger, opts...),
		),
	)

	for _, element := range w.services {
		element(w.grpcServer, w.Contexts)
	}
}

func (w *ServiceGrpc) Run() {
	w.Logger.Infof("Running GRPC Service")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", w.Config.ListenHost, w.Config.Port))
	if err != nil {
		w.Logger.Fatalf("failed to listen: %v", err)
	}

	w.grpcServer.Serve(lis)
}

func (w *ServiceGrpc) Stop() {
	w.Logger.Infof("stop signal received! Graceful shutting down")

	w.grpcServer.GracefulStop()
}

package serviceworker

import (
	"fmt"
	"net"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/radianteam/framework/worker"
	"google.golang.org/grpc"
)

type RegFuncGrpcServiceWorker func(s *grpc.Server, wc *worker.WorkerAdapters)

type GrpcConfig struct {
	Listen string `json:"listen,omitempty" config:"listen,required"`
	Port   int16  `json:"port,omitempty" config:"port,required"`
}

type GrpcServiceWorker struct {
	*worker.BaseWorker

	grpcServer *grpc.Server

	config *GrpcConfig

	regFuncs     []RegFuncGrpcServiceWorker
	errorHandler grpc_recovery.RecoveryHandlerFuncContext
}

func NewGrpcServiceWorker(name string, config *GrpcConfig) *GrpcServiceWorker {
	return &GrpcServiceWorker{
		BaseWorker: worker.NewBaseWorker(name),
		config:     config,
	}
}

func (w *GrpcServiceWorker) AddRegFunc(f RegFuncGrpcServiceWorker) {
	w.regFuncs = append(w.regFuncs, f)
}

func (w *GrpcServiceWorker) SetErrorHandler(f grpc_recovery.RecoveryHandlerFuncContext) {
	w.errorHandler = f
}

func (w *GrpcServiceWorker) Setup() {
	w.Logger.Infof("Setting up GRPC Service")

	opts := []grpc_logrus.Option{}

	grpc_logrus.ReplaceGrpcLogger(w.Logger)

	unaryInters := []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.UnaryServerInterceptor(w.Logger, opts...),
	}

	if w.errorHandler == nil {
		unaryInters = append(unaryInters, grpc_recovery.UnaryServerInterceptor([]grpc_recovery.Option{grpc_recovery.WithRecoveryHandlerContext(w.errorHandler)}...))
	}

	streamInters := []grpc.StreamServerInterceptor{
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_logrus.StreamServerInterceptor(w.Logger, opts...),
	}

	if w.errorHandler == nil {
		streamInters = append(streamInters, grpc_recovery.StreamServerInterceptor([]grpc_recovery.Option{grpc_recovery.WithRecoveryHandlerContext(w.errorHandler)}...))
	}

	w.grpcServer = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(unaryInters...),
		grpc_middleware.WithStreamServerChain(streamInters...),
	)

	for _, regFunc := range w.regFuncs {
		regFunc(w.grpcServer, w.Adapters)
	}
}

func (w *GrpcServiceWorker) Run() {
	w.Logger.Infof("Running GRPC Service")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", w.config.Listen, w.config.Port))
	if err != nil {
		w.Logger.Fatalf("failed to listen: %v", err)
	}

	w.grpcServer.Serve(lis)
}

func (w *GrpcServiceWorker) Stop() {
	w.Logger.Infof("Stop signal received! Graceful shutting down")

	w.grpcServer.GracefulStop()
}

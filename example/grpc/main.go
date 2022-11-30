package main

import (
	"github.com/radianteam/framework"
	"github.com/radianteam/framework/worker"
	grpc_radian "github.com/radianteam/framework/worker/service/grpc"

	"google.golang.org/grpc"
)

type FibonacciServer struct {
	*UnimplementedFibonacciPrinterServer

	adapters *worker.WorkerAdapters
}

func (*FibonacciServer) GetFibonacci(number *Number, server FibonacciPrinter_GetFibonacciServer) error {
	a := int32(0)
	b := int32(1)

	for i := int32(0); i < number.Number; i++ {
		if err := server.Send(&Number{Number: b}); err != nil {
			return err
		}
		b += a
		a = b - a
	}

	return nil
}

// create servicer for worker
func NewServicer() func(s *grpc.Server, wc *worker.WorkerAdapters) {
	return func(s *grpc.Server, wc *worker.WorkerAdapters) {
		RegisterFibonacciPrinterServer(s, &FibonacciServer{adapters: wc})
	}
}

func main() {
	// create a new microservice instance
	radian := framework.NewRadianMicroservice("example")

	// create a new gRPC worker
	grpcConfig := &grpc_radian.GrpcConfig{
		Listen: "0.0.0.0",
		Port:   3009,
	}
	// create a worker
	grpcWorker := grpc_radian.NewGrpcServiceWorker("service_grpc", grpcConfig)
	grpcWorker.AddRegFunc(NewServicer())

	// append the worker to the framework
	radian.AddWorker(grpcWorker)

	// run the worker
	radian.RunAll()
}

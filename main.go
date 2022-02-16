package main

import (
	sidecar "bitbucket.org/realtimeai/kubeslice-router-sidecar/pkg/proto"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"bitbucket.org/realtimeai/kubeslice-router-sidecar/logger"
	slicectl "bitbucket.org/realtimeai/kubeslice-router-sidecar/server"
)

// startGrpcServer shall start the GRPC server to communicate to Slice Controller
func startGrpcServer(grpcPort string) error {
	address := fmt.Sprintf(":%s", grpcPort)
	logger.GlobalLogger.Infof("Starting GRPC Server for SLICEROUTER_POD Pod at %v", address)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		logger.GlobalLogger.Errorf("Unable to connect to Server: %v", err.Error())
		return err
	}

	srv := grpc.NewServer()
	sidecar.RegisterSliceRouterSidecarServiceServer(srv, &slicectl.SliceRouterSidecar{})
	err = srv.Serve(lis)
	if err != nil {
		logger.GlobalLogger.Errorf("Start GRPC Server Failed with %v", err.Error())
		return err
	}
	logger.GlobalLogger.Infof("GRPC Server exited gracefully")

	return nil
}

// shutdownHandler triggers application shutdown.
func shutdownHandler(wg *sync.WaitGroup) {
	// signChan channel is used to transmit signal notifications.
	signChan := make(chan os.Signal, 1)
	// Catch and relay certain signal(s) to signChan channel.
	signal.Notify(signChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Blocking until a signal is sent over signChan channel. Progress to
	// next line after signal
	sig := <-signChan
	logger.GlobalLogger.Infof("Teardown started with ", sig, "signal")

	wg.Done()
	os.Exit(1)
}

func main() {
	var grpcPort, logLevel, metricCollectorPort string

	grpcPort = os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "5000"
	}

	metricCollectorPort = os.Getenv("METRIC_COLLECTOR_PORT")
	if metricCollectorPort == "" {
		metricCollectorPort = "18080"
	}

	logLevel = os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}

	// Create a Logger Module
	logger.GlobalLogger = logger.NewLogger(logLevel)

	err := slicectl.BootstrapSliceRouterPod()
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to bootstrap Kubeslice-router-sidecar pod")
	}

	// Start the GRPC Server to communicate with slice controller.
	go func() {
		err := startGrpcServer(grpcPort)
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to bootstrap startGrpcServer")
		}
	}()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go shutdownHandler(wg)

	wg.Wait()
	logger.GlobalLogger.Infof("kubeslice-router-sidecar exited")
}

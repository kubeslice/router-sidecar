/*  Copyright (c) 2022 Avesha, Inc. All rights reserved.
 *
 *  SPDX-License-Identifier: Apache-2.0
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/kubeslice/router-sidecar/pkg/logger"
	"github.com/kubeslice/router-sidecar/pkg/server"
	sidecar "github.com/kubeslice/router-sidecar/pkg/sidecar/sidecarpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	sidecar.RegisterSliceRouterSidecarServiceServer(srv, &server.SliceRouterSidecar{})

	// TODO: only for debug purpose please revert
	reflection.Register(srv)

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

	err := server.BootstrapSliceRouterPod()
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

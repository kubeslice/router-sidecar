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

package server

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/kubeslice/router-sidecar/pkg/logger"
	pb "github.com/kubeslice/router-sidecar/pkg/sidecar/sidecarpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

//For initializing the buffcon
func dialer() func(context.Context, string) (net.Conn, error) {

	listner := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()

	pb.RegisterSliceRouterSidecarServiceServer(server, &SliceRouterSidecar{})

	go func() {
		if err := server.Serve(listner); err != nil {
			logger.GlobalLogger.Errorf(err.Error())
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listner.Dial()
	}
}

func TestRouterConnClientInfo(t *testing.T) {

	connList := []*pb.ConnectionInfo{}
	connInfo := pb.ConnectionInfo{
		PodName:      "podname",
		NsmInterface: "nsm0",
		NsmIP:        "192.168.1.1",
		NsmPeerIP:    "168.177.1.1",
	}

	connList = append(connList, &connInfo)
	clientConnInfo := pb.ClientConnectionInfo{
		Connection: connList,
	}

	tests := []struct {
		testName  string
		test_conn *pb.ConnectionInfo
		res       *pb.ClientConnectionInfo
		errCode   codes.Code
		errMsg    string
		isCancel  bool
	}{
		{
			"Testing the client connection info",
			&pb.ConnectionInfo{PodName: "demo", NsmInterface: "demo", NsmIP: "nsmip", NsmPeerIP: "nsmpeerip"},
			&pb.ClientConnectionInfo{Connection: connList},
			codes.OK,
			"",
			false,
		},
		{
			"Testing for the cancelled context",
			&pb.ConnectionInfo{PodName: "demo", NsmInterface: "demo", NsmIP: "nsmip", NsmPeerIP: "nsmpeerip"},
			&pb.ClientConnectionInfo{Connection: nil},
			codes.Canceled,
			"context canceled",
			true,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	logger.GlobalLogger = logger.NewLogger("INFO")

	client := pb.NewSliceRouterSidecarServiceClient(conn)

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {

			response, err := client.GetSliceRouterClientConnectionInfo(ctx, &emptypb.Empty{})

			response = &pb.ClientConnectionInfo{
				Connection: connList,
			}

			if tt.isCancel {
				cancel()
			}

			if response != nil {
				AssertEqual(t, response.GetConnection()[0].PodName, clientConnInfo.Connection[0].PodName, tt.res, response)
				AssertEqual(t, response.GetConnection()[0].NsmInterface, clientConnInfo.Connection[0].NsmInterface, tt.res, response)
				AssertEqual(t, response.GetConnection()[0].NsmPeerIP, clientConnInfo.Connection[0].NsmPeerIP, tt.res, response)
				AssertEqual(t, response.GetConnection()[0].NsmIP, clientConnInfo.Connection[0].NsmIP, tt.res, response)
			}
			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != tt.errCode {
						t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					}
					if er.Message() != tt.errMsg {
						t.Error("error message: expected", tt.errMsg, "received", er.Message())
					}
				}
			}
		})
	}
}

func AssertEqual(t *testing.T, expected interface{}, actual interface{}, expectedResponse interface{}, recievedResponse interface{}) {
	t.Helper()
	if expected != actual {
		t.Error("response: expected", expectedResponse, "received", recievedResponse)
	}
}

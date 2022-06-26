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
	"fmt"
	"log"
	"testing"

	pb "github.com/kubeslice/router-sidecar/pkg/sidecar/sidecarpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	st "google.golang.org/grpc/status"
)

const (
	LocalSliceGwVpnIP     = "156.176.1.2"
	LocalSliceGwNsmSubnet = "182.168.1.1/24"
)

func TestUpdateConnCtx(t *testing.T) {

	tests := []struct {
		testName string
		req      *pb.SliceGwConContext
		res      *pb.SidecarResponse
		errCode  codes.Code
		errMsg   string
		isCancel bool
	}{
		{
			"testing update connection context",
			&pb.SliceGwConContext{SliceId: "SliceId", LocalSliceGwId: "LocalSliceGwId", LocalSliceGwVpnIP: LocalSliceGwVpnIP, LocalSliceGwNsmSubnet: LocalSliceGwNsmSubnet, RemoteSliceGwNsmSubnet: "192.168.1.1/24", LocalSliceGwHostType: pb.SliceGwHostType_SLICE_GW_CLIENT, LocalNsmGwPeerIP: "192.156.1.1"},
			&pb.SidecarResponse{StatusMsg: "Slice Gw Connection Context Updated Successfully"},
			codes.InvalidArgument,
			"",
			false,
		},
		{
			"testing for Invalid Remote Slice Gateway Subnet",
			&pb.SliceGwConContext{SliceId: "SliceId", LocalSliceGwId: "LocalSliceGwId", LocalSliceGwVpnIP: LocalSliceGwVpnIP, LocalSliceGwNsmSubnet: LocalSliceGwNsmSubnet, RemoteSliceGwNsmSubnet: "", LocalSliceGwHostType: pb.SliceGwHostType_SLICE_GW_CLIENT, LocalNsmGwPeerIP: "192.156.1.1"},
			&pb.SidecarResponse{StatusMsg: ""},
			codes.InvalidArgument,
			"Invalid Remote Slice Gateway Subnet",
			false,
		},
		{
			"testing for Invalid Remote Slice Gateway Subnet",
			&pb.SliceGwConContext{SliceId: "SliceId", LocalSliceGwId: "LocalSliceGwId", LocalSliceGwVpnIP: LocalSliceGwVpnIP, LocalSliceGwNsmSubnet: LocalSliceGwNsmSubnet, RemoteSliceGwNsmSubnet: "192.168.1.1/24", LocalSliceGwHostType: pb.SliceGwHostType_SLICE_GW_CLIENT, LocalNsmGwPeerIP: ""},
			&pb.SidecarResponse{StatusMsg: ""},
			codes.InvalidArgument,
			"Invalid Local NSM Gateway Peer IP",
			false,
		},
	}
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := grpc.DialContext(ctx, "", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewSliceRouterSidecarServiceClient(conn)
	var errVal error = nil
	fmt.Println(errVal)
	request := pb.SliceGwConContext{}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {

			request = *tt.req

			response, err := client.UpdateSliceGwConnectionContext(ctx, &request)

			if tt.isCancel {
				cancel()
			}

			if response != nil {
				if response.StatusMsg != tt.res.StatusMsg {
					t.Error("response: expected", tt.res, "received", response)
				}
			}
			if err != nil {
				if er, ok := st.FromError(err); ok {
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

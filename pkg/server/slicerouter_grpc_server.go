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

	"github.com/kubeslice/router-sidecar/pkg/logger"
	sidecar "github.com/kubeslice/router-sidecar/pkg/sidecar/sidecarpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SliceRouterSidecar represents the GRPC Server for Slice Router Sidecar.
type SliceRouterSidecar struct {
	sidecar.UnimplementedSliceRouterSidecarServiceServer
}

// Slice router gets the slice GW connection information from the slice controller. This is needed to install
// remote cluster subnet routes into the slice router so that inter-cluster traffic can be forwarded to the right
// slice GW.
func (s *SliceRouterSidecar) UpdateSliceGwConnectionContext(ctx context.Context, conContext *sidecar.SliceGwConContext) (*sidecar.SidecarResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}
	if conContext == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Connection Context is Empty")
	}
	if conContext.GetRemoteSliceGwNsmSubnet() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Remote Slice Gateway Subnet")
	}

	// Note: Do not check for the validity of the conContext.GetLocalNsmGwPeerIPList() here. It is being
	// done in the sliceRouterInjectRoute func.

	err := sliceRouterInjectRoute(conContext.GetRemoteSliceGwNsmSubnet(), conContext.GetLocalNsmGwPeerIPList())
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to add route in slice router: %v", err)
	}

	return &sidecar.SidecarResponse{StatusMsg: "Slice Gw Connection Context Updated Successfully"}, nil
}

// GetClientConnectionInfo requests the slice router sidecar to send connection information of clients
// that are connected to the slice router.
// Slice router is the connectivity entry point to the slice. Any pod that wants to be part of the slice
// connects to the slice router to send and recieve traffic over the slice.
func (s *SliceRouterSidecar) GetSliceRouterClientConnectionInfo(ctx context.Context, in *emptypb.Empty) (*sidecar.ClientConnectionInfo, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}

	connInfo, err := sliceRouterGetClientConnections()
	if err != nil {
		return nil, err
	}

	clientConnInfo := sidecar.ClientConnectionInfo{
		Connection: connInfo,
	}

	logger.GlobalLogger.Debugf("sending conn list: %v", clientConnInfo)

	return &clientConnInfo, nil
}

func (s *SliceRouterSidecar) GetRouteInKernel(ctx context.Context, v *sidecar.VerifyRouteAddRequest) (*sidecar.VerifyRouteAddResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}
	if v.GetDstIP() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Remote Slice Gateway Subnet")
	}
	if v.GetNsmIP() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid NSM IP")
	}
	isPresent, err := vl3GetRouteInKernel(v.DstIP, v.NsmIP)
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to verify route in slice router: %v", err)
		return &sidecar.VerifyRouteAddResponse{IsRoutePresent: false}, err
	}
	return &sidecar.VerifyRouteAddResponse{IsRoutePresent: isPresent}, nil
}

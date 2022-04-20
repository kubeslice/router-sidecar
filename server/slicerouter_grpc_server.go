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

	"github.com/aveshasystems/kubeslice-router-sidecar/logger"
	router "github.com/aveshasystems/kubeslice-router-sidecar/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// SliceRouterSidecar represents the GRPC Server for Slice Router Sidecar.
type SliceRouterSidecar struct {
	router.UnimplementedSliceRouterSidecarServiceServer
}

// Slice router gets the slice GW connection information from the slice controller. This is needed to install
// remote cluster subnet routes into the slice router so that inter-cluster traffic can be forwarded to the right
// slice GW.
func (s *SliceRouterSidecar) UpdateSliceGwConnectionContext(ctx context.Context, conContext *router.SliceGwConContext) (*router.SidecarResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}
	if conContext == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Connection Context is Empty")
	}
	if conContext.GetRemoteSliceGwNsmSubnet() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Remote Slice Gateway Subnet")
	}
	if conContext.GetLocalNsmGwPeerIP() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid Local NSM Gateway Peer IP")
	}
	logger.GlobalLogger.Infof("conContext : %v", conContext)

	err := sliceRouterInjectRoute(conContext.GetRemoteSliceGwNsmSubnet(), conContext.GetLocalNsmGwPeerIP())
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to add route in slice router: %v", err)
	} else {
		logger.GlobalLogger.Infof("Added route in slice router: %v via %v",
			conContext.GetRemoteSliceGwNsmSubnet(), conContext.GetLocalNsmGwPeerIP())
	}

	return &router.SidecarResponse{StatusMsg: "Slice Gw Connection Context Updated Successfully"}, nil
}

// GetClientConnectionInfo requests the slice router sidecar to send connection information of clients
// that are connected to the slice router.
// Slice router is the connectivity entry point to the slice. Any pod that wants to be part of the slice
// connects to the slice router to send and recieve traffic over the slice.
func (s *SliceRouterSidecar) GetSliceRouterClientConnectionInfo(ctx context.Context, in *emptypb.Empty) (*router.ClientConnectionInfo, error) {
	if ctx.Err() == context.Canceled {
		return nil, status.Errorf(codes.Canceled, "Client cancelled, abandoning.")
	}

	connInfo, err := sliceRouterGetClientConnections()
	if err != nil {
		return nil, err
	}
	logger.GlobalLogger.Infof("Rxed conn list: %v", connInfo)
	clientConnInfo := router.ClientConnectionInfo{
		Connection: connInfo,
	}

	logger.GlobalLogger.Infof("sending conn list: %v", clientConnInfo)

	return &clientConnInfo, nil
}

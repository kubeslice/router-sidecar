package server

import (
	"bitbucket.org/realtimeai/kubeslice-router-sidecar/logger"
	router "bitbucket.org/realtimeai/kubeslice-router-sidecar/pkg/proto"
	"context"
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

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
	"errors"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/kubeslice/router-sidecar/pkg/logger"
	sidecar "github.com/kubeslice/router-sidecar/pkg/sidecar/sidecarpb"

	"github.com/lorenzosaino/go-sysctl"
	"github.com/vishvananda/netlink"
	"go.ligato.io/vpp-agent/v3/proto/ligato/configurator"
	"go.ligato.io/vpp-agent/v3/proto/ligato/vpp"
	vpp_l3 "go.ligato.io/vpp-agent/v3/proto/ligato/vpp/l3"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
)

const (
	vppAgentEndpoint                  = "localhost:9113"
	SliceRouterDataplaneVpp    string = "vpp"
	SliceRouterDataplaneKernel string = "kernel"
	/* Routing table reconcilation interval in seconds */
	routingTableReconcileInterval float64 = 60.0
)

// remoteSubnetRouteMap holds all the routes that were injected by the vL3 sidecar into the
// vL3 routing table.
var remoteSubnetRouteMap map[string][]string

// Records the last time the routing table in the slice router was reconciled.
var lastRoutingTableReconcileTime time.Time

func sendConfigToVppAgent(vppconfig *vpp.ConfigData, cfgDelete bool) error {

	dataChange := &configurator.Config{
		VppConfig: vppconfig,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn, err := grpc.Dial(vppAgentEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.GlobalLogger.Errorf("can't dial grpc server: %v", err)
		return err
	}
	defer conn.Close()

	client := configurator.NewConfiguratorServiceClient(conn)

	logger.GlobalLogger.Infof("Sending DataChange to vppagent: %v", dataChange)

	if cfgDelete {
		_, err = client.Delete(ctx, &configurator.DeleteRequest{
			Delete: dataChange,
		})
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to delete vpp config: %v", err)
		}
	} else {
		_, err = client.Update(ctx, &configurator.UpdateRequest{
			Update: dataChange,
		})
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to update vpp config: %v", err)
		}
	}

	return err
}

func vl3InjectRouteInVpp(dstIP string, nextHopIP string) error {
	vppconfig := getVppConfig(dstIP, nextHopIP)
	return sendConfigToVppAgent(vppconfig, false)
}

func getVppConfig(dstIP string, nextHopIP string) *vpp.ConfigData {
	vppconfig := &vpp.ConfigData{}
	route := &vpp.Route{
		Type:        vpp_l3.Route_INTER_VRF,
		DstNetwork:  dstIP,
		NextHopAddr: nextHopIP,
	}
	vppconfig.Routes = append(vppconfig.Routes, route)
	return vppconfig
}

func vl3DeleteRouteInVpp(dstIP string, nextHopIP string) error {
	vppconfig := getVppConfig(dstIP, nextHopIP)
	return sendConfigToVppAgent(vppconfig, true)
}

func vl3InjectRouteInKernel(dstIP string, nextHopIPSlice []*netlink.NexthopInfo) error {
	_, dstIPNet, err := net.ParseCIDR(dstIP)
	if err != nil {
		return err
	}

	route := netlink.Route{Dst: dstIPNet, MultiPath: nextHopIPSlice}

	if err := netlink.RouteAddEcmp(&route); err != nil {
		logger.GlobalLogger.Errorf("Route add failed in kernel. Dst: %v, NextHop: %v, Err: %v", dstIPNet, nextHopIPSlice, err)
		return err
	}
	logger.GlobalLogger.Infof("Route added successfully in the kernel. Dst: %v, NextHop: %v", dstIPNet, nextHopIPSlice)

	return nil
}
func vl3UpdateEcmpRoute(dstIP string, NsmIPToRemove string) error {
	_, dstIPNet, err := net.ParseCIDR(dstIP)
	if err != nil {
		return err
	}
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	logger.GlobalLogger.Infof("routes from routeList %v : %v", routes)
	ecmpRoutes := make([]*netlink.NexthopInfo, 0)
	for _, route := range routes {
		if route.Dst == dstIPNet {
			gwObj := &netlink.NexthopInfo{Gw: net.ParseIP(route.Gw.String())}
			ecmpRoutes = append(ecmpRoutes, gwObj)
		}
	}
	if len(ecmpRoutes) == 0 {
		return errors.New("ecmp routes not yet present")
	}
	updatedMultiPath, index := updateMultipath(ecmpRoutes, NsmIPToRemove)
	err = netlink.RouteDel(&netlink.Route{Gw: ecmpRoutes[index].Gw})
	if err != nil {
		logger.GlobalLogger.Errorf("Unable to delete ecmp routes, Err: %v", err)
		return err
	}
	return vl3InjectRouteInKernel(dstIP, updatedMultiPath)
}
func updateMultipath(nextHopIPs []*netlink.NexthopInfo, gwToRemove string) ([]*netlink.NexthopInfo, int) {
	logger.GlobalLogger.Infof("next hop ips %v\t nsm ip : %v", nextHopIPs, gwToRemove)
	index := -1
	for i, _ := range nextHopIPs {
		if nextHopIPs[i].Gw.String() == gwToRemove {
			index = i
			break
		}
	}
	return append(nextHopIPs[:index], nextHopIPs[index+1:]...), index + 1
}

func vl3GetNsmInterfacesInVpp() ([]*sidecar.ConnectionInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	conn, err := grpc.Dial(vppAgentEndpoint, grpc.WithInsecure())
	if err != nil {
		logger.GlobalLogger.Errorf("can't dial grpc server: %v", err)
		return nil, err
	}
	defer conn.Close()

	client := configurator.NewConfiguratorServiceClient(conn)

	vppConfig, err := client.Get(ctx, &configurator.GetRequest{})
	if err != nil {
		logger.GlobalLogger.Errorf("Failed to get vpp config: %v", err)
		return nil, err
	}

	intfConfig := vppConfig.GetConfig().GetVppConfig().GetInterfaces()
	logger.GlobalLogger.Infof("Vpp intf config: %v", intfConfig)
	if len(intfConfig) == 0 {
		return nil, nil
	}

	connList := []*sidecar.ConnectionInfo{}

	for _, intf := range intfConfig {
		if len(intf.IpAddresses) == 0 {
			continue
		}
		nsmPeerIP := strings.TrimSuffix(intf.IpAddresses[0], "/30")
		nsmIpOctetList := strings.Split(nsmPeerIP, ".")
		nsmIpLastOctet, _ := strconv.Atoi(nsmIpOctetList[3])
		nsmIpOctetList[3] = strconv.Itoa(nsmIpLastOctet - 1)
		nsmIP := strings.Join(nsmIpOctetList, ".")
		conn := sidecar.ConnectionInfo{
			PodName:      intf.Name,
			NsmInterface: "nsm0",
			NsmIP:        nsmIP,
			NsmPeerIP:    nsmPeerIP,
		}
		connList = append(connList, &conn)
	}
	logger.GlobalLogger.Infof("Conn list: %v", connList)

	return connList, nil
}

// vl3GetNsmInterfacesInKernel()
// Returns a list of nsm interfaces created to connect clients to the
// slice router.
func vl3GetNsmInterfacesInKernel() ([]*sidecar.ConnectionInfo, error) {
	links, err := netlink.LinkList()
	if err != nil {
		logger.GlobalLogger.Errorf("Could not get link list, Err: %v", err)
	}

	connList := []*sidecar.ConnectionInfo{}

	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, "nsm") {
			addrList, err := netlink.AddrList(link, unix.AF_INET)
			if err != nil {
				logger.GlobalLogger.Errorf("Failed to get address list for intf: %v, err: %v",
					link.Attrs().Name, err)
				continue
			}
			if len(addrList) != 1 {
				logger.GlobalLogger.Infof("More than one address on nsm intf: %v", addrList)
				continue
			}
			// nsmIP is the IP address assigned to the client. The prefix pool is a /30 address.
			// That gives us 4 IP addresses. NSM NSEs assign the .1 to the client and the .2 to the
			// server. The .0 adress is the network address and the .3 is the broadcast address.
			// We derive the nsm IP assigned to the client and the nsm IP assigned to the server
			// from the broadcast address.
			// nsmIP is the client IP and nsmPeerIP is the IP on the slice router.
			nsmIP := net.IP{
				addrList[0].Broadcast[0],
				addrList[0].Broadcast[1],
				addrList[0].Broadcast[2],
				addrList[0].Broadcast[3] - 2,
			}
			nsmPeerIP := net.IP{
				addrList[0].Broadcast[0],
				addrList[0].Broadcast[1],
				addrList[0].Broadcast[2],
				addrList[0].Broadcast[3] - 1,
			}

			conn := sidecar.ConnectionInfo{
				PodName:      link.Attrs().Alias,
				NsmInterface: "nsm0",
				NsmIP:        nsmIP.String(),
				NsmPeerIP:    nsmPeerIP.String(),
			}
			connList = append(connList, &conn)
		}
	}

	logger.GlobalLogger.Infof("Conn list: %v", connList)

	return connList, nil
}

func vl3ReconcileRoutesInKernel() error {
	// Build a map of existing routes in the vl3
	installedRoutes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	routeMap := make(map[string][]netlink.Route)
	for _, route := range installedRoutes {
		// Default route will have a Dst of nil so it is
		// important to have a null check here. Else we will
		// crash trying to deref a null pointer.
		if route.Dst == nil {
			continue
		}
		routeMap[route.Dst.String()] = append(routeMap[route.Dst.String()], route)
	}
	logger.GlobalLogger.Infof("Route map: %v", routeMap)
	logger.GlobalLogger.Infof("Slice Route map: %v", remoteSubnetRouteMap)

	for remoteSubnet, nextHopIpList := range remoteSubnetRouteMap {
		for i := 0; i < len(nextHopIpList); i++ {
			_, ok := routeMap[remoteSubnet]
			// If the route is absent or the nexthop is incorrect, reinstall the route.
			if !ok || containsRoute(routeMap[remoteSubnet], nextHopIpList[i]) {

				nextHopIpSlice := []*netlink.NexthopInfo{}
				logger.GlobalLogger.Infof("nsmips: %v, index: %v", routeMap[remoteSubnet], i)
				gwObj := &netlink.NexthopInfo{Gw: net.ParseIP(routeMap[remoteSubnet][i].Gw.String())}
				nextHopIpSlice = append(nextHopIpSlice, gwObj)

				logger.GlobalLogger.Infof("Installed route does not reflect slice state. Reconciling dst: %v, gw: %v", remoteSubnet, nextHopIpList[i])
				err := vl3InjectRouteInKernel(remoteSubnet, nextHopIpSlice)
				if err != nil {
					logger.GlobalLogger.Errorf("Failed to install route: dst: %v, gw: %v", remoteSubnet, nextHopIpSlice)
				}
			}
		}
	}
	return nil
}

func sliceRouterReconcileRoutingTable() error {
	if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
		return nil
	} else {
		return vl3ReconcileRoutesInKernel()
	}
}

// Function to inject remote cluster subnet routes into the local slice router.
// The next hop IP would be the IP address of the slice-gw that connects to the remote cluster.
func sliceRouterInjectRoute(remoteSubnet string, nextHopIPList []string) error {
	logger.GlobalLogger.Infof("Received NSM IPS from operator: %v", nextHopIPList)
	if time.Since(lastRoutingTableReconcileTime).Seconds() > routingTableReconcileInterval {
		err := sliceRouterReconcileRoutingTable()
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to reconcile routing table: %v", err)
		}

		lastRoutingTableReconcileTime = time.Now()

		logger.GlobalLogger.Infof("RT reconciled at: %v", lastRoutingTableReconcileTime)
	}

	_, routePresent := remoteSubnetRouteMap[remoteSubnet]
	nextHopIpSlice := []*netlink.NexthopInfo{}

	for i := 0; i < len(nextHopIPList); i++ {

		gwObj := &netlink.NexthopInfo{Gw: net.ParseIP(nextHopIPList[i])}
		nextHopIpSlice = append(nextHopIpSlice, gwObj)

		if routePresent && checkRouteAdd(remoteSubnetRouteMap[remoteSubnet], nextHopIPList[i]) {
			logger.GlobalLogger.Infof("Ignoring route add request. Route already installed. RemoteSubnet: %v, NextHop: %v",
				remoteSubnet, nextHopIPList[i])
			return nil
		}

		if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
			// If a route was previously installed for the remote subnet then we should
			// delete it before adding a route with a new nexthop IP.
			// VPP treats a route modify as a route add operation, creating multiple
			// entries for a destination prefix and treating them as equal cost multipath
			// routes.
			// In our case, we should have only one route with the nexthop as the nsm IP on
			// the slice gw pod connecting the remote subnet.
			if len(remoteSubnetRouteMap[remoteSubnet]) != 0 {
				err := vl3DeleteRouteInVpp(remoteSubnet, remoteSubnetRouteMap[remoteSubnet][i])
				if err != nil {
					logger.GlobalLogger.Errorf("Failed to delete route with old gw IP. RemoteSubent: %v, NextHop: %v",
						remoteSubnet, remoteSubnetRouteMap[remoteSubnet][i])
					return err
				}
			}
			err := vl3InjectRouteInVpp(remoteSubnet, nextHopIPList[i])
			if err != nil {
				return err
			}
		}

		if getSliceRouterDataplaneMode() == SliceRouterDataplaneKernel && i == len(nextHopIPList)-1 {
			err := vl3InjectRouteInKernel(remoteSubnet, nextHopIpSlice)
			if err != nil {
				return err
			}
		}
		remoteSubnetRouteMap[remoteSubnet] = append(remoteSubnetRouteMap[remoteSubnet], nextHopIPList[i])
	}
	return nil
}
func checkRouteAdd(nextHopIpList []string, s string) bool {
	for _, nextHop := range nextHopIpList {
		if nextHop == s {
			return true
		}
	}
	return false
}

func containsRoute(nextHopIpList []netlink.Route, s string) bool {
	for _, nextHop := range nextHopIpList {
		if nextHop.Gw.String() == s {
			return true
		}
	}
	return false
}

func sliceRouterGetClientConnections() ([]*sidecar.ConnectionInfo, error) {
	if getSliceRouterDataplaneMode() == SliceRouterDataplaneKernel {
		return vl3GetNsmInterfacesInKernel()
	} else if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
		return vl3GetNsmInterfacesInVpp()
	}

	return nil, nil
}

func getSliceRouterDataplaneMode() string {
	return os.Getenv("DATAPLANE")
}

func BootstrapSliceRouterPod() error {
	if getSliceRouterDataplaneMode() == SliceRouterDataplaneKernel {
		err := sysctl.Set("net.ipv4.ip_forward", "1")
		if err != nil {
			logger.GlobalLogger.Fatalf("Failed to enable IP forwarding in the kernel", err)
			return err
		}
		err = sysctl.Set("net.ipv4.fib_multipath_hash_policy", "1")
		if err != nil {
			logger.GlobalLogger.Fatalf("failed to set hash policy to L4 for mutipath routes", err)
			return err
		}
	}
	remoteSubnetRouteMap = make(map[string][]string)
	lastRoutingTableReconcileTime = time.Now()
	return nil
}

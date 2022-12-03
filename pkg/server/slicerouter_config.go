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
	"sync"
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
var remoteSubnetRouteMap sync.Map

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
	if err := netlink.RouteReplace(&route); err != nil {
		logger.GlobalLogger.Errorf("Route add failed in kernel. Dst: %v, NextHop: %v, Err: %v", dstIPNet, nextHopIPSlice, err)
		return err
	}
	logger.GlobalLogger.Infof("Route added successfully in the kernel. Dst: %v, NextHop: %v", dstIPNet, nextHopIPSlice)

	return nil
}
func vl3UpdateEcmpRoute(dstIP string, NsmIPToRemove string) error {
	logger.GlobalLogger.Info("request to remove nsmIP from routes", "nsmip", NsmIPToRemove)
	_, dstIPNet, err := net.ParseCIDR(dstIP)
	if err != nil {
		return err
	}
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	ecmpRoutes := make([]*netlink.NexthopInfo, 0)
	for _, route := range routes {
		if route.Dst.String() == dstIPNet.String() {
			ecmpRoutes = route.MultiPath
		}
	}
	if len(ecmpRoutes) == 0 {
		return errors.New("ecmp routes not yet present")
	}
	logger.GlobalLogger.Info("ecmpRoutes", "ecmpRoutes", ecmpRoutes)

	updatedMultiPath, _ := updateMultipath(ecmpRoutes, NsmIPToRemove)
	logger.GlobalLogger.Info("updatedMultiPath", "updatedMultiPath", updatedMultiPath)
	logger.GlobalLogger.Info("ecmpRoutes after update", "ecmpRoutes", ecmpRoutes)
	err = netlink.RouteReplace(&netlink.Route{Dst: dstIPNet, MultiPath: updatedMultiPath})
	if err != nil {
		logger.GlobalLogger.Errorf("Unable to replace ecmp routes, Err: %v", err)
		return err
	}
	remoteSubnetRouteMap.Store(dstIP, contructArrayFromNextHop(updatedMultiPath))
	logger.GlobalLogger.Info("remoteSubnetRouteMap", "remoteSubnetRouteMap", remoteSubnetRouteMap)
	return nil
}

func updateMultipath(nextHopIPs []*netlink.NexthopInfo, gwToRemove string) ([]*netlink.NexthopInfo, int) {
	index := -1
	for i, _ := range nextHopIPs {
		if nextHopIPs[i].Gw.String() == gwToRemove {
			index = i
			break
		}
	}
	// if gwToRemove not found
	if index == -1 {
		return nextHopIPs, index
	}
	return append(nextHopIPs[:index], nextHopIPs[index+1:]...), index
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
		return nil, err
	}

	installedRoutes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		logger.GlobalLogger.Errorf("Could not get route list, Err: %v", err)
		return nil, err
	}

	intfMap := make(map[int]string)

	for _, route := range installedRoutes {
		if route.Dst == nil {
			continue
		}
		intfMap[route.LinkIndex] = route.Dst.String()
	}

	logger.GlobalLogger.Infof("intf map: %v", intfMap)

	connList := []*sidecar.ConnectionInfo{}

	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Name, "vl3-") {
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

			// nsmIP is the IP address on the app pod, whereas nsmPeerIP is the IP address on the
			// corresponding link on the vl3 slice router
			nsmIP := strings.Split(intfMap[link.Attrs().Index], "/")[0]
			nsmPeerIP := addrList[0].IP.String()

			conn := sidecar.ConnectionInfo{
				PodName:      link.Attrs().Alias,
				NsmInterface: "nsm0",
				NsmIP:        nsmIP,
				NsmPeerIP:    nsmPeerIP,
			}
			connList = append(connList, &conn)
		}
	}

	logger.GlobalLogger.Infof("Conn list: %v", connList)

	return connList, nil
}

func vl3GetRouteInKernel(dstIP string, nsmIP string) (bool, error) {
	logger.GlobalLogger.Info("get route in kernel", "dstIP", dstIP, "nsmip", nsmIP)
	_, dstIPNet, err := net.ParseCIDR(dstIP)
	if err != nil {
		return false, err
	}
	gwIP := net.ParseIP(nsmIP)

	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return false, err
	}
	ecmpRoutes := make([]*netlink.NexthopInfo, 0)
	for _, route := range routes {
		if route.Dst.String() == dstIPNet.String() {
			// check if the route is present
			ecmpRoutes = route.MultiPath
		}
	}
	if len(ecmpRoutes) == 0 {
		return false, errors.New("ecmp routes not yet present")
	}

	logger.GlobalLogger.Info("ranging over ecmp routes", ecmpRoutes)
	for _, r := range ecmpRoutes {
		if r.Gw.String() == nsmIP {
			return true, nil
		}
	}
	logger.GlobalLogger.Errorf("NextHop is Invalid route not added YET. Dst: %v, NextHop: %v, Err: %v", dstIPNet, gwIP, err)
	return false, nil
}

func vl3ReconcileRoutesInKernel() error {
	// Build a map of existing routes in the vl3
	installedRoutes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	routeMap := make(map[string][]netlink.Route, 0)
	for _, route := range installedRoutes {
		// Default route will have a Dst of nil so it is
		// important to have a null check here. Else we will
		// crash trying to deref a null pointer.
		if route.Dst == nil {
			continue
		}
		routeMap[route.Dst.String()] = append(routeMap[route.Dst.String()], route)
	}
	logger.GlobalLogger.Infof("installed Route: %v", installedRoutes)
	logger.GlobalLogger.Infof("Route map: %v", routeMap)
	logger.GlobalLogger.Infof("Slice Route map: %v", remoteSubnetRouteMap)

	nextHopInfoSlice := []*netlink.NexthopInfo{}
	remoteSubnetRouteMap.Range(func(key, value any) bool {
		nextHopList := value.([]string)
		remoteSubnet := key.(string)
		for _, ip := range nextHopList {
			_, ok := routeMap[remoteSubnet]
			if !ok || !containsRoute(routeMap[remoteSubnet], ip) {
				nextHopInfoSlice = getNextHopInfoSlice(nextHopList)
			}
		}
		if len(nextHopInfoSlice) > 0 {
			logger.GlobalLogger.Infof("Installed route does not reflect slice state. Reconciling dst: %v, gw: %v", remoteSubnet, nextHopInfoSlice)
			err := vl3InjectRouteInKernel(remoteSubnet, nextHopInfoSlice)
			if err != nil {
				logger.GlobalLogger.Errorf("Failed to install route: dst: %v, gw: %v", remoteSubnet, nextHopInfoSlice)
				return false
			}
			remoteSubnetRouteMap.Store(remoteSubnet, contructArrayFromNextHop(nextHopInfoSlice))
		}
		logger.GlobalLogger.Errorf("Skipping installing routes since they are already present!")
		return true
	})
	return nil
}

func getNextHopInfoSlice(nextHopIPList []string) []*netlink.NexthopInfo {
	nextHopIpSlice := []*netlink.NexthopInfo{}
	for _, ip := range nextHopIPList {
		if net.ParseIP(ip) == nil {
			continue
		}
		gwObj := &netlink.NexthopInfo{Gw: net.ParseIP(ip)}
		nextHopIpSlice = append(nextHopIpSlice, gwObj)
	}
	return nextHopIpSlice
}

// contructArrayFromNextHop takes  []*netlink.NexthopInfo and flattens nextHop IPs to []string
func contructArrayFromNextHop(nextHopIP []*netlink.NexthopInfo) []string {
	var nextHopIPList []string
	for _, ip := range nextHopIP {
		nextHopIPList = append(nextHopIPList, ip.Gw.String())
	}
	return nextHopIPList
}

func sliceRouterReconcileRoutingTable() error {
	if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
		return nil
	} else {
		return vl3ReconcileRoutesInKernel()
	}
}

func buildNextHopInfo(nextHopIPList []string) []*netlink.NexthopInfo {
	nextHopIpSlice := []*netlink.NexthopInfo{}
	for _, nextHop := range nextHopIPList {
		gwObj := &netlink.NexthopInfo{Gw: net.ParseIP(nextHop)}
		nextHopIpSlice = append(nextHopIpSlice, gwObj)
	}
	return nextHopIpSlice
}

// Function to inject remote cluster subnet routes into the local slice router.
// The next hop IP would be the IP address of the slice-gw that connects to the remote cluster.
func sliceRouterInjectRoute(remoteSubnet string, nextHopIPList []string) error {
	logger.GlobalLogger.Infof("Received NSM IPS from operator: %v", nextHopIPList)
	if time.Since(lastRoutingTableReconcileTime).Seconds() > routingTableReconcileInterval {
		err := sliceRouterReconcileRoutingTable()
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to reconcile routing table: %v", err)
			return err
		}

		lastRoutingTableReconcileTime = time.Now()

		logger.GlobalLogger.Infof("RT reconciled at: %v", lastRoutingTableReconcileTime)
	}
	logger.GlobalLogger.Infof("sliceRouterInjectRoute", "remoteSubnetRouteMap", remoteSubnetRouteMap)

	_, routePresent := remoteSubnetRouteMap.Load(remoteSubnet)
	nextHopInfoSlice := getNextHopInfoSlice(nextHopIPList)

	for i := 0; i < len(nextHopIPList); i++ {

		nextHopList, _ := remoteSubnetRouteMap.Load(remoteSubnet)
		if routePresent && checkRouteAdd(nextHopList.([]string), nextHopIPList[i]) {
			logger.GlobalLogger.Infof("Ignoring route add request. Route already installed. RemoteSubnet: %v, NextHop: %v",
				remoteSubnet, nextHopIPList[i])
			continue
		}
		if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
			// If a route was previously installed for the remote subnet then we should
			// delete it before adding a route with a new nexthop IP.
			// VPP treats a route modify as a route add operation, creating multiple
			// entries for a destination prefix and treating them as equal cost multipath
			// routes.
			// In our case, we should have only one route with the nexthop as the nsm IP on
			// the slice gw pod connecting the remote subnet.
			nextHopList, _ := remoteSubnetRouteMap.Load(remoteSubnet)
			nextHopListConverted := nextHopList.([]string)
			if len(nextHopListConverted) != 0 {
				err := vl3DeleteRouteInVpp(remoteSubnet, nextHopListConverted[i])
				if err != nil {
					logger.GlobalLogger.Errorf("Failed to delete route with old gw IP. RemoteSubent: %v, NextHop: %v",
						remoteSubnet, nextHopListConverted[i])
				}
			}
			err := vl3InjectRouteInVpp(remoteSubnet, nextHopIPList[i])
			if err != nil {
				logger.GlobalLogger.Errorf("Failed to inject route in vpp: %v", err)
			}
		} else {
			err := vl3InjectRouteInKernel(remoteSubnet, nextHopInfoSlice)
			if err != nil {
				logger.GlobalLogger.Errorf("Failed to inject route in kernel: %v", err)
				// do not add entry in gloabl map in case of error and continue for next route enteries
				continue
			}
		}
		nextHopList, _ = remoteSubnetRouteMap.Load(remoteSubnet)
		remoteSubnetRouteMap.Store(remoteSubnet, append(nextHopList.([]string), nextHopIPList[i]))
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
	lastRoutingTableReconcileTime = time.Now()
	return nil
}

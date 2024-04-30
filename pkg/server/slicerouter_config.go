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
	"fmt"
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

	"sync"

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

	logger.GlobalLogger.Debugf("intf map: %v", intfMap)

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
				logger.GlobalLogger.Infof("No address or more than one address on nsm intf: %v", addrList)
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

	logger.GlobalLogger.Debugf("Conn list: %v", connList)

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
		links, err := netlink.LinkList()
		if err != nil {
			logger.GlobalLogger.Errorf("Could not get link list, Err: %v", err)
			return false, err
		}

		for _, link := range links {
			if strings.HasPrefix(link.Attrs().Name, "vl3") {
				// Get the routes
				logger.GlobalLogger.Info("link name", "link", link.Attrs().Name, "link index", link.Attrs().Index)
				routes, err := netlink.RouteList(link, netlink.FAMILY_V4)
				if err != nil {
					return false, err
				}
				logger.GlobalLogger.Info("routes list inside", "routes", routes)
				// range throw the routes
				for _, route := range routes {
					if route.Gw.String() == nsmIP {
						// route with Dst=nsmIP is added in routing table
						return true, nil
					}
				}
			}
		}
		return false, errors.New(fmt.Errorf("route with Dst=%s not yet present", nsmIP).Error())
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

func printSliceRouteMap() {
	logger.GlobalLogger.Debugf("Slice Route map:")
	remoteSubnetRouteMap.Range(func(key, value any) bool {
		nextHopList := value.([]string)
		remoteSubnet := key.(string)
		logger.GlobalLogger.Debugf("remoteSubnet: %v, nexthop: %v", remoteSubnet, nextHopList)
		return true
	})
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

	logger.GlobalLogger.Debugf("Installed routes map: %v", routeMap)
	printSliceRouteMap()

	nextHopInfoSlice := []*netlink.NexthopInfo{}
	remoteSubnetRouteMap.Range(func(key, value any) bool {
		nextHopList := value.([]string)
		remoteSubnet := key.(string)
		for _, ip := range nextHopList {
			_, ok := routeMap[remoteSubnet]
			if !ok || !containsRoute(routeMap[remoteSubnet], ip) {
				nextHopInfoSlice, err = getNetlinkNextHopInfo(nextHopList)
				if err != nil {
					return false
				}
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
		} else {
			logger.GlobalLogger.Debugf("Skipping installing routes since they are already present!")
		}
		return true
	})
	return nil
}

func getNetlinkNextHopInfo(nextHopIPList []string) ([]*netlink.NexthopInfo, error) {
	nextHopIpSlice := []*netlink.NexthopInfo{}
	for _, nextHopIP := range nextHopIPList {
		installedRoutes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
		if err != nil {
			return nil, err
		}
		var linkIdx int = -1
		for _, route := range installedRoutes {
			if route.Dst == nil {
				continue
			}
			// Default route will have a Dst of nil so it is
			// important to have a null check here. Else we will
			// crash trying to deref a null pointer.
			if route.Dst.String() == nextHopIP+"/32" {
				linkIdx = route.LinkIndex
				gwObj := &netlink.NexthopInfo{LinkIndex: linkIdx, Gw: net.ParseIP(nextHopIP), Flags: int(netlink.FLAG_ONLINK)}
				nextHopIpSlice = append(nextHopIpSlice, gwObj)
				break
			}
		}
		if linkIdx == -1 {
			return nil, errors.New(fmt.Sprintf("link idx of nexthop not found for %v", nextHopIP))
		}
	}
	return nextHopIpSlice, nil
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

func sliceRouterDeleteRouteToDst(dstIP string) error {
	_, dstIPNet, err := net.ParseCIDR(dstIP)
	if err != nil {
		return err
	}

	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if route.Dst.String() == dstIPNet.String() {
			err := netlink.RouteDel(&route)
			return err
		}
	}

	return errors.New("Route to delete not found")
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

		logger.GlobalLogger.Debugf("RT reconciled at: %v", lastRoutingTableReconcileTime)
	}

	printSliceRouteMap()

	if len(nextHopIPList) == 0 {
		// Treat this as a signal to delete the route to the remoteSubnet
		err := sliceRouterDeleteRouteToDst(remoteSubnet)
		if err != nil {
			return err
		}
		remoteSubnetRouteMap.Delete(remoteSubnet)
		return nil
	}

	installRoute := false

	cachedNextHopValue, routePresent := remoteSubnetRouteMap.Load(remoteSubnet)
	if !routePresent {
		installRoute = true
	} else {
		cachedNextHopList := cachedNextHopValue.([]string)
		// Route is present in the cache. Check if the stored nexthop matches with the nexthop received
		// in the input param.
		// We reinstall the route if the two lists do not match.
		if len(cachedNextHopList) != len(nextHopIPList) {
			installRoute = true
		} else {
			for _, nextHopInCache := range cachedNextHopList {
				if !contains(nextHopIPList, nextHopInCache) {
					installRoute = true
					break
				}
			}
		}
	}

	if !installRoute {
		return nil
	}

	// Convert nexthop IPs in string to netlink nexthop info struct
	netlinkNextHopList, err := getNetlinkNextHopInfo(nextHopIPList)
	if err != nil {
		return err
	}

	if getSliceRouterDataplaneMode() == SliceRouterDataplaneVpp {
		for i := 0; i < len(nextHopIPList); i++ {
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
		}
	} else {
		err := vl3InjectRouteInKernel(remoteSubnet, netlinkNextHopList)
		if err != nil {
			logger.GlobalLogger.Errorf("Failed to inject route in kernel: %v", err)
			return err
		}
	}

	// at the end of for loop , the global map should contain the exact routes that are installed
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return err
	}
	ecmpRoutes := make([]*netlink.NexthopInfo, 0)
	for _, route := range routes {
		if route.Dst.String() == remoteSubnet {
			ecmpRoutes = route.MultiPath
		}
	}
	remoteSubnetRouteMap.Store(remoteSubnet, contructArrayFromNextHop(ecmpRoutes))
	return nil
}

func contains(items []string, s string) bool {
	for _, item := range items {
		if item == s {
			return true
		}
	}
	return false
}

func containsRoute(routeList []netlink.Route, s string) bool {
	for _, route := range routeList {
		if len(route.MultiPath) > 0 {
			for _, path := range route.MultiPath {
				if path.Gw.String() == s {
					return true
				}
			}
		} else {
			if route.Gw.String() == s {
				return true
			}
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
		// Turn on the forwarding in the kernel. It is an absolute must since the router
		// needs to forward traffic to app and gw pods.
		err := sysctl.Set("net.ipv4.ip_forward", "1")
		if err != nil {
			logger.GlobalLogger.Fatalf("Failed to enable IP forwarding in the kernel", err)
			return err
		}
		// Set the ecmp hash policy to consider L3 and L4 (IP + Port) if possible. This will help with
		// improving the load balancing between the multi paths.
		// This configuration might not be available on some operating systems. First check if the config
		// option is available before attempting to update it.
		val, err := sysctl.Get("net.ipv4.fib_multipath_hash_policy")
		if err == nil {
			// Config option is available. Set the config if it does not have the needed value.
			if val != "1" {
				err = sysctl.Set("net.ipv4.fib_multipath_hash_policy", "1")
				if err != nil {
					logger.GlobalLogger.Fatalf("failed to set hash policy to L4 for mutipath routes", err)
					return err
				}
			} else {
				logger.GlobalLogger.Debugf("Hash policy already set")
			}
		} else {
			// If the config option is not available, log an error message and let the platform use the default method
			// to load balance.
			logger.GlobalLogger.Errorf("Hash policy cannot be set on this platform..", err)
		}
	}
	lastRoutingTableReconcileTime = time.Now()
	return nil
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"

	"github.com/erikbos/apiauth/pkg/db"
	// "github.com/erikbos/apiauth/pkg/types"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v2"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
	xds "github.com/envoyproxy/go-control-plane/pkg/server"
	"github.com/envoyproxy/go-control-plane/pkg/util"
	"github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"net"
	"os"
	"sync/atomic"
)

var (
	localhost = "127.0.0.1"
	//	port        = ":9901"
	//	gatewayPort = ":9902"
	version int32
)

type logger struct{}

func (logger logger) Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}
func (logger logger) Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

// Hasher returns node ID as an ID
type Hasher struct {
}

// ID function
func (h Hasher) ID(node *core.Node) string {
	if node == nil {
		return "unknown"
	}
	return node.Id
}

func runManagementServer(server xds.Server, GRPCPort string) {
	lis, err := net.Listen("tcp", GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	log.Info("running management server")

	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, server)
	v2.RegisterEndpointDiscoveryServiceServer(grpcServer, server)
	v2.RegisterClusterDiscoveryServiceServer(grpcServer, server)
	v2.RegisterRouteDiscoveryServiceServer(grpcServer, server)
	v2.RegisterListenerDiscoveryServiceServer(grpcServer, server)

	//gateway
	//go proxy.Call()

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to start GRPC server: %v", err)
	}
}

func runManagementGateway(srv xds.Server, HTTPPort string) {

	log.Info("running gateway")
	err := http.ListenAndServe(HTTPPort, &xds.HTTPGateway{Server: srv})
	if err != nil {
		log.Fatalf("failed to HTTP serve: %v", err)
	}
}

func main() {
	configFilename := flag.String("configfilename", "envoycp-config.yaml", "Configuration filename")
	flag.Parse()
	config := loadConfiguration(*configFilename)
	// FIXME we should check if we have all required parameters (use viper package?)

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000000",
		FullTimestamp:   true,
		DisableColors:   true,
	})
	log.SetLevel(log.DebugLevel)

	// var err error
	db, err := db.Connect(config.DatabaseHostname, config.DatabasePort,
		config.DatabaseUsername, config.DatabasePassword, config.DatabaseKeyspace)
	if err != nil {
		log.Fatalf("Database connect failed: %v", err)
		os.Exit(1)
	}

	//ctx := context.Background()
	EnvoyConfig := cache.NewSnapshotCache(true, Hasher{}, logger{})
	//config := cache.NewSnapshotCache(false, hash{}, logger{})
	signal := make(chan struct{})
	srv := xds.NewServer(EnvoyConfig, &callbacks{
		signal:   signal,
		fetches:  0,
		requests: 0,
	})

	go runManagementServer(srv, config.GRPCXDSListen)
	go runManagementGateway(srv, config.HTTPXDSListen)

	for {
		var clusterName = "service_google"
		// var remoteHost = "www.google.com"
		// var sni = "www.google.com"
		log.Infof(">>>>>>>>>>>>>>>>>>> creating cluster " + clusterName)

		//c := []cache.Resource{resource.MakeCluster(resource.Ads, clusterName)}
		/*h := &core.Address{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  remoteHost,
				Protocol: core.TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(443),
				},
			},
		}}*/

		EnvoyClusters, err := getClusterConfig(db)

		var listenerName = "listener_0"
		var targetHost = "www.google.com"
		var targetPrefix = "/"
		var virtualHostName = "local_service"
		var routeConfigName = "local_route"
		log.Infof(">>>>>>>>>>>>>>>>>>> creating listener " + listenerName)
		//virtual host (inside http filter)
		v := route.VirtualHost{
			Name:    virtualHostName,
			Domains: []string{"*"},
			Routes: []route.Route{{
				Match: route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: targetPrefix,
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						HostRewriteSpecifier: &route.RouteAction_HostRewrite{
							HostRewrite: targetHost,
						},
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			}},
		}

		//http filter (inside listener)
		manager := &hcm.HttpConnectionManager{
			StatPrefix: "ingress_http",
			RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
				RouteConfig: &v2.RouteConfiguration{
					Name:         routeConfigName,
					VirtualHosts: []route.VirtualHost{v},
				},
			},
			HttpFilters: []*hcm.HttpFilter{{
				Name: util.Router,
			}},
		}
		pbst, err := types.MarshalAny(manager)
		if err != nil {
			fmt.Println("yellow")
			panic(err)
		}
		fmt.Println(pbst.TypeUrl)

		//create listener
		var l = []cache.Resource{
			&v2.Listener{
				Name: listenerName,
				Address: core.Address{
					Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol: core.TCP,
							Address:  localhost,
							PortSpecifier: &core.SocketAddress_PortValue{
								PortValue: 10000,
							},
						},
					},
				},
				FilterChains: []listener.FilterChain{{
					Filters: []listener.Filter{{
						Name: util.HTTPConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: pbst,
						},
					}},
				}},
			}}

		log.Printf("%+v", l)
		atomic.AddInt32(&version, 1)
		//nodeId := config.GetStatusKeys()[1]

		log.Infof(">>>>>>>>>>>>>>>>>>> creating snapshot Version " + fmt.Sprint(version))
		snap := cache.NewSnapshot(fmt.Sprint(version), nil, EnvoyClusters, nil, l)

		_ = EnvoyConfig.SetSnapshot("jenny", snap)

		reader := bufio.NewReader(os.Stdin)
		_, _ = reader.ReadString('\n')
	}

}

module github.com/erikbos/apiauth

require (
	github.com/bmatcuk/doublestar v1.2.2
	github.com/coocood/freecache v1.1.0
	github.com/cosiner/argv v0.0.1 // indirect
	github.com/envoyproxy/go-control-plane v0.8.0
	github.com/gin-gonic/gin v1.5.0
	github.com/go-delve/delve v1.4.0 // indirect
	github.com/gocql/gocql v0.0.0-20191126110522-1982a06ad6b9
	github.com/gogo/googleapis v1.2.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.8 // indirect
	github.com/oschwald/maxminddb-golang v1.6.0
	github.com/peterh/liner v1.2.0 // indirect
	github.com/prometheus/client_golang v1.3.0
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.6 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.starlark.net v0.0.0-20200203144150-6677ee5c7211 // indirect
	golang.org/x/arch v0.0.0-20191126211547-368ea8f32fff // indirect
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c // indirect
	google.golang.org/grpc v1.21.0
	gopkg.in/yaml.v2 v2.2.8
)

replace github.com/erikbos/apiauth/pkg/db => ../apiauth/pkg/db

go 1.13

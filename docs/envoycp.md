# Intro

Envoycp is the control plane for envoyproxy. It configures all listeners, listeners, routes and their specifics in envoyproxy.

## How does it work

Envoycp continously monitors the database for updates changes to listeners, routes and clusters. In case there is a change a new envoyproxy configuration will be compiled and pushed to all envoyproxy.

## Envoycp endpoints

Envoycp exposes two endpoints:

| name     | scope   | protocol | purpose                                |
| -------- | ------- | -------- | -------------------------------------- |
| webadmin | private | http     | admin console, prometheus metrics, etc |
| xds      | private | grpc     | xds configuration with envoyproxy      |

Both `webadmin.listen` and `xds.listen` should be set to configure listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself
- _public_, meant to be publicly on public internet

## Deployment

### Configuration

Envoycp requires a starup configuration which needs to be provided as YAML file, see below for the supported fields. For an example configuration file see [envoycp.yaml](../deployment/docker/envoycp.yaml).

### Logfiles

Envoycp writes multiple logfiles, one for each function of envoycp. All are written as structured JSON, filename rotation schedule can be set via configuration file. The two logfiles are:

1. `logging.filename` as log for application messages
2. `webadmin.logging.filename` as access log for all REST API calls

### Envoycp configuration file

The supported fields are:

| yaml field                  | purpose                                              | example            |
| --------------------------- | ---------------------------------------------------- | ------------------ |
| logging.level               | application log level                                | info / debug       |
| logging.filename            | filename to write application log to                 | /dev/stdout        |
| logging.maxsize             | Maximum size in megabytes before rotate              | 100                |
| logging.maxage              | Max days to retain old log files                     | 7                  |
| logging.maxbackups          | Maximum number of old log files to retain            | 14                 |
| webadmin.listen             | Webadmin address and port                            | 0.0.0.0:2113       |
| webadmin.ipacl              | Webadmin ip acl, without this no access              | 172.16.0.0/19      |
| webadmin.tls.certfile       | TLS certificate file                                 |                    |
| webadmin.tls.keyfile        | TLS certificate key file                             |                    |
| webadmin.logging.level      | logging level of webadmin                            | info / debug       |
| webadmin.logging.filename   | filename to write web access log to                  | dbadmin-access.log |
| webadmin.logging.maxsize    | Maximum size in megabytes before rotate              | 100                |
| webadmin.logging.maxage     | Max days to retain old log files                     | 7                  |
| webadmin.logging.maxbackups | Maximum number of old log files to retain            | 14                 |
| database.hostname           | Cassandra hostname to connect to                     | cassandra          |
| database.port               | Cassandra port to connect on                         | 9042 / 10350       |
| database.tls                | Enable TLS for database session                      | true / false       |
| database.username           | Database username                                    | cassandra          |
| database.password           | Database password                                    | cassandra          |
| database.keyspace           | Database keyspace for Gatekeeper tables              | gatekeeper         |
| database.timeout            | Timeout for session                                  | 0.5s               |
| database.connectattempts    | Number of attempts to establish connection           | 5                  |
| database.queryretries       | Number of times to retry query                       | 2                  |
| xds.listen                  | Listen address and port for XDS requests from Envoy  | 0.0.0.0:9901       |
| xds.configcompileinterval   | Minimum interval between XDS configuration snapshots | 1s                 |
| xds.cluster                 | Name of cluster that runs XDS                        |                    |
| xds.timeout                 | Maximum duration of XDS requests                     | 2s                 |

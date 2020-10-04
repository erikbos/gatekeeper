# Cluster

A cluster defines an upstream backend. Each provides one or more APIs to be consumed by application developers. Attributes can be set to configure connectivity related settings like TLS, timeouts, health checks, connection handling, retry behaviour, etc.

## Supported operations

| Method | Path                                         | What                               |
| ------ | -------------------------------------------- | ---------------------------------- |
| GET    | /v1/clusters                                 | retrieve all clusters              |
| POST   | /v1/clusters                                 | creates a new cluster              |
| GET    | /v1/clusters/_clustername_                   | retrieve a cluster                 |
| POST   | /v1/clusters/_clustername_                   | updates an existing cluster        |
| DELETE | /v1/clusters/_clustername_                   | delete cluster                     |
| GET    | /v1/clusters/_clustername_/attributes        | retrieve all attributes of cluster |
| POST   | /v1/clusters/_clustername_/attributes        | updates all attributes of cluster  |
| GET    | /v1/clusters/_clustername_/attributes/_name_ | retrieve one cluster attribute     |
| POST   | /v1/clusters/_clustername_/attributes/_name_ | update one cluster attribute       |
| DELETE | /v1/clusters/_clustername_/attributes/_name_ | delete one cluster attribute       |

_For POST content-type: application/json is required._

## Example cluster entity

Cluster `ticketshop` running on  host`ticketbackend.svc` port `80`:

```json
{
    "name": "ticketshop",
    "displayName": "Ticket API",
    "hostName": "ticketbackend.svc",
    "port": 80,
    "attributes": [
        {
            "name": "Host",
            "value": "ticketbackend.svc"
        },
        {
            "name": "Port",
            "value": "80"
        }
    ]
}
```

## Fields specification

| fieldname   | optional  | purpose                                                                   |
| ----------- | --------- | ------------------------------------------------------------------------- |
| name        | mandatory | name (cannot be updated afterwards!)                                      |
| displayName | optional  | friendly name                                                             |
| hostName    | mandatory | hostname of backend                                                       |
| port        | mandatory | port of backend                                                           |
| attributes  | optional  | configure connectivity parameters between Envoyproxy and upstream backend |

## Attribute specification

| attribute name                | purpose                                                                                 | example values               |
| ----------------------------- | --------------------------------------------------------------------------------------- | ---------------------------- |
| Host                          | Host to connect                                                                         | backend.example.com          |
| Port                          | Port number to connect on                                                               | 80                           |
| ConnectTimeout                | The timeout for new network connections to cluster                                      | 1s                           |
| IdleTimeout                   | The idle timeout for requests on a connection                                           | 60s                          |
| DNSLookupFamily               | IP network address family to use when resolving cluster hostname                        | IPV4_ONLY,IPV6_ONLY,Auto     |
| DNSRefreshRate                | Refreshrate for resolving cluster hostname                                              | 5s                           |
| DNSResolvers                  | Resolver ip address(es) to resolve cluster hostname (multiple can be comma separated)   | 1.1.1.1,8.8.8.8              |
| TLS                           | Whether to enable TLS or not, HTTP/2 always uses TLS                                    | true, false                  |
| SNIHostName                   | Hostname to send during TLS handshake (if not set hostname will be used)                | backend.example.com          |
| TLSMinimumVersion             | Minimum version of TLS to use                                                           | TLS1.0,TLS1.1, TLS1.2 TLS1.3 |
| TLSMaximumVersion             | Maximum version of TLS to use                                                           | TLS1.0,TLS1.1, TLS1.2 TLS1.3 |
| TLSCipherSuites               | Allowed TLS cipher suite                                                                |                              |
| HTTPProtocol                  | Protocol to use when contacting upstream                                                | HTTP/1.1, HTTP/2, HTTP/3     |
| LbPolicy                      | Endpoint load balancing algorithm    | ROUND_ROBIN, LEAST_REQUEST, RING_HASH, RANDOM, MAGLEV                           |
| HealthCheckProtocol           | Network protocol to use for health check                                                | HTTP                         |
| HealthCheckHostHeader         | Host header to use for health check                                                     | www.example.com              |
| HealthCheckPath               | HTTP Path of health check probe                                                         | /liveness                    |
| HealthCheckInterval           | Health check interval                                                                   | 5s                           |
| HealthCheckTimeout            | Health check timeout                                                                    | 5s                           |
| HealthCheckUnhealthyThreshold | Threshold of events before declaring cluster unhealth                                   | 3                            |
| HealthCheckHealthyThreshold   | Threshold of events before declaring clustern health                                    | 1                            |
| HealthCheckLogFile            | Logfile name for healthcheck probes                                                     | /tmp/healthcheck             |
| MaxConnections                | The maximum number of connections to make to the upstream cluster                       | 1000                         |
| MaxPendingRequests            | The maximum number of pending requests to make to the upstream cluster                  | 1024                         |
| MaxRequests                   | The maximum number of parallel requests to make to the upstream cluster                 | 1024                         |
| MaxRetries                    | The maximum number of parallel retries to make to the upstream cluster                  | 3                            |

All attributes listed above are mapped onto configuration properties of [Envoy Cluster API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v3/api/v3/cluster.proto#cluster) for detailed explanation of purpose and allowed value of each attribute.

The cluster options exposed this way are a subset of Envoy's capabilities, in general any cluster configuration option Envoy supports can be exposed  this way. Feel free to open an issue if you need more of Envoy's functionality exposed.

## Envoycp control plane

Envoycp monitors the database for changed clusters at `xds.configcompileinterval` interval. In case of changes envoycp will compile a new Envoy configuration and notify all envoyproxy instances.

## Example cluster configurations

Cluster `ticketshop` running on `ticketbackend.svc` port `80`:

```json
{
    "name": "ticketshop",
    "displayName": "Ticket API",
    "hostName": "ticketbackend.svc",
    "port": 80,
    "attributes": [
        {
            "name": "Host",
            "value": "ticketbackend.svc"
        },
        {
            "name": "Port",
            "value": "80"
        }
    ]
}
```

Cluster `people` with elaborate TLS, health check and DNS resolving settings:

```json
{
    "name": "people",
    "displayName": "People API",
    "hostName": "127.0.0.1",
    "port": 8000,
    "attributes": [
        {
            "name": "Host",
            "value": "ticketbackend.svc"
        },
        {
            "name": "Port",
            "value": "443"
        }
        {
            "name": "TLS",
            "value": "true"
        },
        {
            "name": "TLSMinimumVersion",
            "value": "TLS1.2"
        },
        {
            "name": "TLSCipherSuites",
            "value": "[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305],ECDHE-ECDSA-AES256-GCM-SHA384"
        },
        {
            "name": "HTTPProtocol",
            "value": "HTTP/2"
        },
        {
            "name": "SNIHostName",
            "value": "www.example.com"
        },
        {
            "name": "HealthCheckProtocol",
            "value": "HTTP"
        },
        {
            "name": "MaxConnections",
            "value": "700"
        },
        {
            "name": "HealthCheckPath",
            "value": "/people/1"
        },
        {
            "name": "HealthCheckInterval",
            "value": "2s"
        },
        {
            "name": "HealthCheckTimeout",
            "value": "1s"
        },
        {
            "name": "HealthCheckLogFile",
            "value": "/tmp/healthcheck.log"
        },
        {
            "name": "DNSRefreshRate",
            "value": "5s"
        },
        {
            "name": "DNSResolvers",
            "value": "8.8.8.8,1.1.1.1"
        }
    ]
}

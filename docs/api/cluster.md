# Cluster

A cluster defines an upstream backend. Each provides one or more APIs to be consumed by application developers. Attributes can be set to configure connectivity related settings like TLS, timeouts, health checks, connection handling, retry behaviour, etc.

## Supported methods and paths

| Method | Path                                         | What                               |
| ------ | -------------------------------------------- | ---------------------------------- |
| GET    | /v1/clusters                                 | retrieve all clusters              |
| POST   | /v1/clusters                                 | creates a new cluster              |
| GET    | /v1/clusters/_clustername_                   | retrieve a cluster                 |
| POST   | /v1/clusters/_clustername_                   | updates an existing cluster        |
| DELETE | /v1/clusters/_clustername_                   | deletes a cluster                  |
| GET    | /v1/clusters/_clustername_/attributes        | retrieve all attributes of cluster |
| POST   | /v1/clusters/_clustername_/attributes        | updates all attributes of cluster  |
| GET    | /v1/clusters/_clustername_/attributes/_name_ | retrieve attribute of cluster      |
| POST   | /v1/clusters/_clustername_/attributes/_name_ | updates attribute of cluster       |
| DELETE | /v1/clusters/_clustername_/attributes/_name_ | delete attribute of cluster        |

* For POST content-type: application/json is required.

## Example cluster definition

```json
{
    "name": "ticketshop",
    "displayName": "Ticket API",
    "hostName": "ticketbackend.svc",
    "port": 80,
    "attributes": [
        {
        "name": "ConnectTimeout",
        "value": "7s"
    },
    {
        "name": "IdleTimeout",
        "value": "4s"
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
| ConnectTimeout                | The timeout for new network connections to cluster                                      | 1s                           |
| IdleTimeout                   | The idle timeout for connections, for the period in which there are no active requests. | 60s                          |
| DNSLookupFamily               | IP network address family for contact cluster                                           | IPv6                         |
| DNSRefreshRate                | Refreshrate for resolving cluster hostname                                              | 5s                           |
| DNSResolvers                  | Resolver ip address to resolve cluster hostname (multiple can be comma separated)       | 1.1.1.1,8.8.8.8              |
| HealthCheckProtocol           | Network protocol to use for health check                                                | HTTP                         |
| HealthCheckPath               | HTTP Path of health check probe                                                         | /liveness                    |
| HealthCheckInterval           | Health check interval                                                                   | 5s                           |
| HealthCheckTimeout            | Health check timeout                                                                    | 5s                           |
| HealthCheckUnhealthyThreshold | Threshold of events before declaring cluster unhealth                                   | 3                            |
| HealthCheckHealthyThreshold   | Threshold of events before declaring clustern health                                    | 1                            |
| HealthCheckLogFile            | Logfile name for healthcheck probes                                                     | /tmp/healthcheck             |
| HTTPProtocol                  | Protocol to use when contacting upstream                                                | HTTP/1.1, HTTP/2, HTTP/3     |
| MaxConnections                | The maximum number of connections that Envoy will make to the upstream cluster          | 1000                         |
| MaxPendingRequests            | The maximum number of pending requests that Envoy will allow to the upstream cluster    | 1024                         |
| MaxRequests                   | The maximum number of parallel requests that Envoy will make to the upstream cluster    | 1024                         |
| MaxRetries                    | The maximum number of parallel retries that Envoy will allow to the upstream cluster    | 3                            |
| TLSEnabled                    | Whether to enable TLS or not, HTTP/2 always uses TLS                                    | true, false                  |
| SNIHostName                   | Hostname to send during TLS handshake (if not set hostname will be used)           |                              |
| TLSMinimumVersion             | Minimum version of TLS to use                                                           | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSMaximumVersion             | Maximum version of TLS to use                                                           | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSCipherSuites               | Allowed TLS cipher suite                                                                |                              |

All attributes listed above are mapped on configuration properties of [Envoy Cluster API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v3/api/v3/cluster.proto#cluster) for detailed explanation of purpose and allowed value of each attribute.

The cluster options exposed this way are a subset of Envoy's capabilities, in general any cluster configuration option Envoy supports can be exposed  this way. Feel free to open an issue if you need more of Envoy's functionality exposed.

## Background

Envoycp checks the database for new or changed clusters every second. Unrecognized attributes will be ignored and a warning will be logged. In case of any changes envoycp will compile a new proxy configuration and push it to all envoyproxy instances.


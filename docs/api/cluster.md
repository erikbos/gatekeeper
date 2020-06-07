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

| attribute name                | purpose | possible values              |
| ----------------------------- | ------- | ---------------------------- |
| ConnectTimeout                |         |                              |
| IdleTimeout                   |         |                              |
| DNSLookupFamily               |         |                              |
| DNSRefreshRate                |         |                              |
| DNSResolvers                  |         |                              |
| HealthCheckProtocol           |         | HTTP                         |
| HealthCheckPath               |         |                              |
| HealthCheckInterval           |         |                              |
| HealthCheckTimeout            |         |                              |
| HealthCheckUnhealthyThreshold |         |                              |
| HealthCheckHealthyThreshold   |         |                              |
| HealthCheckLogFile            |         |                              |
| HTTPProtocol                  |         | HTTP/1.1, HTTP/2, HTTP/3     |
| MaxConnections                |         |                              |
| MaxConnections                |         |                              |
| MaxPendingRequests            |         |                              |
| MaxRequests                   |         |                              |
| MaxRetries                    |         |                              |
| TLSEnabled                    |         | true, false                  |
| SNIHostName                   |         |                              |
| TLSMinimumVersion             |         | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSMaximumVersion             |         | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSCipherSuites               |         |                              |

All attributes listed above are mapped on configuration properties of [Envoy Cluster API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/cluster.proto#cluster) for detailed explanation of purpose and allowed value of each attribute.

(The route options exposed this way are a subnet of Envoy's capabilities, in general any cluster configuration option Envoy supports can be easily exposed)

## Background

Envoycp check the database for new or changed clusters every second. In case of any changes envoy will compile a new proxy configuration and send it to all envoyproxy instances.

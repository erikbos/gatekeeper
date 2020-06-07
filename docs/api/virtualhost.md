# Virtualhost

A virtualhost defines listing port, virtual settings and incoming connectivity parameters.

_Listening port, http host(s) and TLS settings are all configured in a virtual host._

## Supported methods and paths

| Method | Path                                             | What                                   |
| ------ | ------------------------------------------------ | -------------------------------------- |
| GET    | /v1/virtualhosts                                 | retrieve all virtualhosts              |
| POST   | /v1/virtualhosts                                 | creates a new virtualhost              |
| GET    | /v1/virtualhosts/_virtualhost_                   | retrieve a virtualhost                 |
| POST   | /v1/virtualhosts/_virtualhost_                   | updates an existing virtualhost        |
| DELETE | /v1/virtualhosts/_virtualhost_                   | deletes a virtualhost                  |
| GET    | /v1/virtualhosts/_virtualhost_/attributes        | retrieve all attributes of virtualhost |
| POST   | /v1/virtualhosts/_virtualhost_/attributes        | update all attribute of virtualhost    |
| GET    | /v1/virtualhosts/_virtualhost_/attributes/_name_ | retrieve attribute of virtualhost      |
| POST   | /v1/virtualhosts/_virtualhost_/attributes/_name_ | update attribute of virtualhost        |
| DELETE | /v1/virtualhosts/_virtualhost_/attributes/_name_ | delete attribute of virtualhost        |

_For POST content-type: application/json is required._

## Example virtualhost definition

```json
{
    "name": "example_80",
    "displayName": "Example Inc.",
    "virtualHosts": [
         "www.petstore.com"
    ],
    "port": 80,
    "organizationName": "petstore",
    "routeGroup": "routes_80"
}
```

## Fields specification

| fieldname        | optional  | purpose                                           |
| ---------------- | --------- | ------------------------------------------------- |
| name             | mandatory | name (cannot be updated afterwards)               |
| displayName      | optional  | friendly name                                     |
| virtualHosts     | mandatory | array of virtal hostnames                         |
| port             | mandatory | port Envoy will listen on                         |
| organizationName | mandatory | organization name                                 |
| routeGroup       | mandatory | indicate which http routing table will be applied |

## Attribute specification

| attribute name    | purpose | possible values              |
| ----------------- | ------- | ---------------------------- |
| HTTPProtocol      |         | HTTP/1.1, HTTP/2, HTTP/3     |
| TLSEnabled        |         | true, false                  |
| TLSCertificate    |         |                              |
| TLSCertificateKey |         |                              |
| TLSMinimumVersion |         | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSMaximumVersion |         | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSCipherSuites   |         |                              |

All attributes listed above are mapped on configuration properties of [Envoy listener API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/listener.proto#listener) for detailed explanation of purpose and allowed value of each attribute.

(The route options exposed this way are a subnet of Envoy's capabilities, in general any listener and http configuration option Envoy supports can be easily exposed)

## Background

Envoycp check the database for new or changed virtualhosts every second. In case of any changes envoy will compile a new proxy configuration and send it to all envoyproxy instances.

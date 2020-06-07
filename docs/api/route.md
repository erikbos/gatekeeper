# Route

A route defines how a specific path needs to handled and forwarded on. All operations which are applied at HTTP-level to every request are configured here.

## Supported methods and paths

| Method | Path                                     | What                             |
| ------ | ---------------------------------------- | -------------------------------- |
| GET    | /v1/routes                               | retrieve all routes              |
| POST   | /v1/routes                               | creates a new route              |
| GET    | /v1/routes/_routename_                   | retrieve a route                 |
| POST   | /v1/routes/_routename_                   | updates an existing route        |
| DELETE | /v1/routes/_routename_                   | deletes a routes                 |
| GET    | /v1/routes/_routename_/attributes        | retrieve all attributes of route |
| POST   | /v1/routes/_routename_/attributes        | update all attribute of route    |
| GET    | /v1/routes/_routename_/attributes/_name_ | retrieve attribute of route      |
| POST   | /v1/routes/_routename_/attributes/_name_ | update attribute of route        |
| DELETE | /v1/routes/_routename_/attributes/_name_ | delete attribute of route        |

* For POST content-type: application/json is required.

## Example route definition

```json
{
    "name": "default_ticketshop",
    "displayName": "Default route ticketshop",
    "routeGroup": "routes_443",
    "path": "/ticketshop",
    "pathType": "prefix",
    "cluster": "ticketshop"
}
```

## Fields specification

| fieldname   | optional  | purpose                                                     |
| ----------- | --------- | ----------------------------------------------------------- |
| name        | mandatory | name (cannot be updated afterwards)                         |
| displayName | optional  | friendly name                                               |
| path        | mandatory | path to match on                                            |
| pathType    | mandatory | _path_ for an exact path match                              |
|             |           | _prefix_ to match a route starting with a particular string |
|             |           | _regexp_ use a (RE2) regular expression to match            |
| routeGroup  | mandatory | routing table name                                          |

## Attribute specification

Every route can have optional attributes which control what Envoy will do to match the incoming request, respond directly without contacting a backend, or to add additional headers before the request is forwarded upstream.

| attribute name           | purpose | possible values |
| ------------------------ | ------- | --------------- |
| DirectResponseStatusCode |         |                 |
| DirectResponseBody       |         |                 |
| PrefixRewrite            |         |                 |
| CORSAllowCredentials     |         |                 |
| CORSAllowMethods         |         |                 |
| CORSAllowHeaders         |         |                 |
| CORSExposeHeaders        |         |                 |
| CORSMaxAge               |         |                 |
| HostHeader               |         |                 |
| BasicAuth                |         |                 |
| RetryOn                  |         |                 |
| PerTryTimeout            |         |                 |
| NumRetries               |         |                 |
| RetryOnStatusCodes       |         |                 |

All attributes listed above are mapped on configuration properties of [Envoy route API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route_components.proto#envoy-api-msg-route-route) for detailed explanation of purpose and allowed value of each attribute.

(The route options exposed this way are a subnet of Envoy's capabilities, in general any route configuration Envoy supports can be easily added)

## Background

Envoycp check the database for new or changed routes every second. In case of any changes envoy will compile a new proxy configuration and send it to all envoyproxy instances.

## More examples

Direct response without contact backend:

```json
{
    "name": "default80",
    "displayName": "Default HTTP route",
    "routeGroup": "routes_80",
    "path": "/",
    "pathType": "prefix",
    "cluster": "none",
    "attributes": [
    {
        "name": "DirectResponseStatusCode",
        "value": "200"
    },
    {
        "name": "DirectResponseBody",
        "value": "We do not support plain HTTP anymore, please use HTTPS"
    }
    ]
}
```

Enable handling of Cross-Origin Resource Sharing (CORS):

```json

{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "cluster": "people",
    "attributes": [
    {
        "name": "CORSAllowMethods",
        "value": "GET,POST,DELETE,OPTIONS"
    },
    {
        "name": "CORSAllowHeaders",
        "value": "User-Agent-X"
    },
    {
        "name": "CORSExposeHeaders",
        "value": "Shoesize"
    },
    {
        "name": "CORSMaxAge",
        "value": "3600"
    },
    {
        "name": "CORSAllowCredentials",
        "value": "true"
    }
    ]
}
```

Set route specific request retry behaviour to reduce error rates:

```json

{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "cluster": "people",
    "attributes": [
    {
        "name": "NumRetries",
        "value": "3"
    },
    {
        "name": "RetryOn",
        "value": "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes"
    },
    {
        "name": "PerTryTimeout",
        "value": "250ms"
    },
    {
        "name": "RetryOnStatusCodes",
        "value": "503,504"
    }
    ]
}
```

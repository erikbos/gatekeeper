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

| attribute name           | purpose                                                               | possible values |
| ------------------------ | --------------------------------------------------------------------- | --------------- |
| DisableAuthentication    | Disable authentication on route                                       | true            |
| DirectResponseStatusCode | Return an arbitrary HTTP response directly, without proxying.         | 200             |
| DirectResponseBody       | Body to return when DirectResponseStatusCode is set                   | Hello World     |
| PrefixRewrite            | Rewrites path when contacting upstream                                |                 |
| CORSAllowCredentials     | Specifies whether the resource allows credentials                     | false           |
| CORSAllowMethods         | Specifies the content for the [Access-Control-Allow-Methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods) header    |                 |
| CORSAllowHeaders         | Specifies the content for the [Access-Control-Allow-Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers) header     |                 |
| CORSExposeHeaders        | Specifies the content for the [Access-Control-Expose-Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers) header    |                 |
| CORSMaxAge               | Specifies the content for the [Access-Control-Max-Age](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age) header           |                 |
| HostHeader               | HTTP host header to set when contact upstream cluster                 |                 |
| BasicAuth                | HTTP Basic authentication header to set when contact upstream cluster | user:secret     |
| RetryOn                  | Specifies the conditions under which retry takes place.               | [See envoy documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-retry-on)|
| PerTryTimeout            | Specify upstream timeout per retry attempt                            | 150ms           |
| NumRetries               | Specify the allowed number of retries                                 | 1               |
| RetryOnStatusCodes       |                                                                       | 503,504         |

All attributes listed above are mapped on configuration properties of [Envoy route API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route_components.proto#envoy-api-msg-route-route) for detailed explanation of purpose and allowed value of each attribute.

The route options exposed this way are a subset of Envoy's capabilities, in general any route configuration option Envoy supports can be exposed  this way. Feel free to open an issue if you need more of Envoy's functionality exposed.

## Background

Envoycp check the database for new or changed routes every second. In case of any changes envoycp will compile a new proxy configuration and push it to all envoyproxy instances.

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

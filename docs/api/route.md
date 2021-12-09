# Route

A route defines how requests for an exact path or path prefix needs to handled: forwarding to an upstream cluster, mirroring, [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing), etc. Most of these need to be configured using specific attributes.

## Supported operations

| Method | Path                                     | What                             |
| ------ | ---------------------------------------- | -------------------------------- |
| GET    | /v1/routes                               | Retrieve all routes              |
| POST   | /v1/routes                               | Creates a new route              |
| GET    | /v1/routes/_routename_                   | Retrieve a route                 |
| POST   | /v1/routes/_routename_                   | Updates an existing route        |
| DELETE | /v1/routes/_routename_                   | Deletes a routes                 |
| GET    | /v1/routes/_routename_/attributes        | Retrieve all route attributes    |
| POST   | /v1/routes/_routename_/attributes        | Update all route attributes      |
| GET    | /v1/routes/_routename_/attributes/_name_ | Retrieve one route attribute     |
| POST   | /v1/routes/_routename_/attributes/_name_ | Update one route attribute       |
| DELETE | /v1/routes/_routename_/attributes/_name_ | Delete one route attribute       |

_For POST content-type: application/json is required._

## Example route entity

Forward all traffic for path `/ticket` of route group `route_443` to upstream cluster `ticketshop`:

```json
{
    "name": "ticketshop",
    "displayName": "ticketshop v1 API",
    "routeGroup": "routes_443",
    "path": "/ticketshop",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Cluster",
            "value": "ticketshop"
        }
    ]
}
```

## Fields specification

| fieldname   | optional  | purpose                                                         |
| ----------- | --------- | --------------------------------------------------------------- |
| name        | mandatory | name (cannot be updated afterwards)                             |
| displayName | optional  | friendly name                                                   |
| path        | mandatory | path to match on                                                |
| pathType    | mandatory | Use _path_ for an exact path match                              |
|             |           | Use _prefix_ to match a path starting with a particular prefix  |
|             |           | Use _regexp_ to match using a [RE2](https://en.wikipedia.org/wiki/RE2_(software)) regular expression |
| routeGroup  | mandatory | routing table name                                              |
| attributes  | optional  | Specific configuration to apply                                 |

## Attribute specification

Every route can have optional attributes which control what Envoy will do to match the incoming request, respond directly without contacting a backend, or to add additional headers before the request is forwarded to an upstream cluster.

| attribute name           | purpose                                                       | possible values         |
| ------------------------ | ------------------------------------------------------------- | ----------------------- |
| Cluster                  | Name of upstream cluster to forward requests to               |                         |
| WeightedClusters         | Weighted list of clusters to load balance requests across     | backend:95,newbackend:5 |
| ExtAuthz                 | Enable/disable request authentication via extauthz            | false, true             |
| RateLimiting             | Enable/disable request ratelimiting via ratelimiter           | false, true             |
| DirectResponseStatusCode | Return an arbitrary HTTP response directly, without proxying. | 200                     |
| DirectResponseBody       | Responsebody to return when direct response is done           | Hello World             |
| RedirectStatusCode       | Return an HTTP redirect                                       | 301,302,303,307 or 308  |
| RedirectScheme           | Set HTTP scheme when generating a redirect                    | http or https           |
| RedirectHostName         | Set hostname when generating a redirect                       | www.example.com         |
| RedirectPort             | Set port when generating a redirect                           | 443                     |
| RedirectPath             | Set path when generating a redirect                           | /test/                  |
| RedirectStripQuery       | Enable removal of query parameters when redirecting           | true                    |
| PrefixRewrite            | Rewrites path when contacting upstream                        |                         |
| CORSAllowCredentials     | Specifies whether the resource allows credentials             | false                   |
| CORSAllowMethods         | Specifies the content for the [Access-Control-Allow-Methods](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods) header    |                 |
| CORSAllowHeaders         | Specifies the content for the [Access-Control-Allow-Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers) header     |                 |
| CORSExposeHeaders        | Specifies the content for the [Access-Control-Expose-Headers](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Headers) header    |                 |
| CORSMaxAge               | Specifies the content for the [Access-Control-Max-Age](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age) header           |                 |
| HostHeader               | HTTP host header to set when forwarding to upstream cluster           |                 |
| RequestHeaderToAdd1      | Additional header to set when forwarding to upstream cluster          |                 |
| RequestHeaderToAdd2      | Additional header to set when forwarding to upstream cluster          |                 |
| RequestHeaderToAdd3      | Additional header to set when forwarding to upstream cluster          |                 |
| RequestHeaderToAdd4      | Additional header to set when forwarding to upstream cluster          |                 |
| RequestHeaderToAdd5      | Additional  header to set when forwarding to upstream cluster         |                 |
| RequestHeadersToRemove   | Headers to remove before forwarding to upstream cluster               | accept,x-age    |
| BasicAuth                | Basic authentication header to set when contact upstream cluster      | user:secret     |
| RetryOn                  | Specifies the conditions under which retry takes place.               | [See envoy documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/router_filter#config-http-filters-router-x-envoy-retry-on)|
| PerTryTimeout            | Specify upstream timeout per retry attempt                            | 150ms           |
| NumRetries               | Specify the allowed number of retries                                 | 1               |
| RetryOnStatusCodes       | Upstream status codes which are to be retried                         | 503,504         |

All attributes listed above are mapped onto configuration properties of [Envoy route API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route.proto) for detailed explanation of purpose and allowed value of each attribute.

The route options exposed this way are a subset of Envoy's capabilities, in general any route configuration option Envoy supports can be exposed  this way. Feel free to open an issue if you need more of Envoy's functionality exposed.

## Controlplane

Controlplane monitors the database for changed routes at `xds.configcompileinterval` interval. In case of changes controlplane will compile a new Envoy configuration and notify all envoyproxy instances.

## Example route configurations

Forward all traffic for path `/ticket` of routeGroup `route_443` to upstream cluster `ticketshop`:

```json
{
    "name": "ticketshop",
    "displayName": "ticketshop v1 API",
    "routeGroup": "routes_443",
    "path": "/ticketshop",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Cluster",
            "value": "ticketshop"
        }
    ]
}
```

Direct response by Envoy, without contacting an upstream cluster, with status code `200` and `responsebody` for path `/`:

```json
{
    "name": "default80",
    "displayName": "Default HTTP route",
    "routeGroup": "routes_80",
    "path": "/",
    "pathType": "path",
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

Redirect path prefix `/login` using status code `301` to `https://www.example.com/new_login`:

```json
{
    "name": "old_login_redirect",
    "displayName": "Redirect old login",
    "routeGroup": "routes_80",
    "path": "/login",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "RedirectStatusCode",
            "value": "301"
        },
        {
            "name": "RedirectScheme",
            "value": "https"
        },
        {
            "name": "RedirectHostName",
            "value": "www.example.com"
        },
        {
            "name": "RedirectPath",
            "value": "/new_login/"
        }
    ],
}
```

Forward `/people` to cluster `people` and enable handling of [CORS](https://en.wikipedia.org/wiki/Cross-origin_resource_sharing) by Envoy:

```json
{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Cluster",
            "value": "people"
        },
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

Forward `/people` to cluster `people` and configure up to `3` request retries in case upstream cluster returns `503,504`:

```json
{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Cluster",
            "value": "people"
        },
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

Forward `/people` to cluster `people` remove header `Content-Type` and set header `appid` to the value
of Metadata key `app.id` (emitted by `authserver`)

```json
{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "HostHeader",
            "value": "www.example.com"
        },
        {
            "name": "RequestHeadersToAdd1",
            "value": "appid=%DYNAMIC_METADATA([\"envoy.filters.http.ext_authz\", \"app.id\"])%"
        },
        {
            "name": "RequestHeadersToAdd2",
            "value": "service=public"
        },
        {
        "name": "RequestHeadersToRemove",
        "value": "Delete-This-Header,Content-Type"
        }
    ]
}
```

Set multiple upstream clusters for path `/people`, use weight distribution `25` / `75`. Multiple upstream clusters can be separated by comma. Each need to be assigned a load balancing weight using *:value* suffix.

```json
{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "WeightedClusters",
            "value": "people:25,people:75"
        }
    ]
}
```

Request forwarding of path `/people` to upstream cluster `people`, while mirroring `12%` of those requests to second upstream cluster `people_v2`:

```json
{
    "name": "people",
    "displayName": "Default people",
    "routeGroup": "routes_443",
    "path": "/people",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Cluster",
            "value": "people"
        },
        {
            "name": "RequestMirrorCluster",
            "value": "people_v2"
        },
        {
            "name": "RequestMirrorPercentage",
            "value": "12"
        }
    ]
}
```

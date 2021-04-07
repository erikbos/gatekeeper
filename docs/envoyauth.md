# Intro

Envoyauth's purpose is to authenticate for envoyproxy. Incoming requests to envoyproxy will first be send to envoyauth for authentication and authorization.

## Authentication

Authentication works by XXXX

## Envoyauth endpoints

Envoyauth exposes multiple endpoints:

| name      | scope   | protocol | purpose                                     |
| --------- | ------- | -------- | ------------------------------------------- |
| webadmin  | private | http     | admin console, prometheus metrics, etc      |
| envoyauth | private | grpc     | authentication requests by envoyproxy       |
| oauth     | public  | http     | requests for [OAuth2 access tokens](#OAuth2)|

For each there is a corresponding `<endpoint>.listen` config field option to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself.
- _public_, meant to be exposed on public internet.

## Deployment

### Configuration

Envoyauth requires a startup configuration which needs to be provided as YAML file, see below for the supported fields. For an example configuration file see [envoyauth.yaml](../deployment/docker/envoyauth.yaml)

### Enabling authentication of requests

To add authentication into a listener's request path the following [listener attributes](api/listener.md#Attribute) needs to be set:

- `Filters` must include `envoy.filters.http.ext_authz` as one of the filters. E.g.: "envoy.filters.http.ext_authz,envoy.filters.http.cors"

The following [listener attributes](api/listener.md#Attribute) configure the ("extauthz") authentication cluster details:

- `ExtAuthzCluster`, *required*, name of cluster which runs envoyauth
- `ExtAuthzTimeout` optional, maximum allowed duration of authentication requests to envoyauth
- `ExtAuthzFailureModeAllow` optional, should requests be forwarded or not in case envoyauth is not reachable or does not respond in time
-  `ExtAuthzRequestBodySize` optional, number of bytes of the request body envoyproxy should forward to envoyauth. This is a prerequisite for envoyauth being able to inspect the body of a request.

### Enable authentication for a route

By default authentication for a route is disabled. To have Envoyproxy forward request to the authentication cluster set [route attribute](api/route.md#Attribute) `ExtAuthz` to `true`. The name of the authentication cluster used is determined by listener attribute `AuthenticationCluster`.

### OAuth2

Envoyauth supports issueing and authentication using [OAuth 2 Client Credentials](https://aaronparecki.com/oauth-2-simplified/#client-credentials) mode.

Next to the above mentioned attributes to inse, two entities need to be configured:

1. A route that forwards the paths `oauth.tokenissuepath` and `oauth.tokeninfopath` to an OAuth cluster. Authentication should be disabled as OAuth2 endpoints are meant to be public.
2. A cluster that accesses envoyauth on port `oauth.listen`, to make sure OAuth requests go to this public endpoint of envoyauth.

Example route entity:

```json
{
    "name": "oauth2",
    "displayName": "OAuth2 authentication service 👍🏼",
    "RouteGroup": "routes_443",
    "path": "/oauth2",
    "pathType": "prefix",
    "attributes": [
        {
            "name": "Authentication",
            "value": "false"
        },
        {
            "name": "Cluster",
            "value": "oauth2"
        },
    ],
}
```

Example cluster entity:

```json
{
    "name": "oauth2",
    "displayName": "OAuth2 token API",
    "attributes": [
        {
            "name": "Host",
            "value": "envoyauth"
        },
        {
            "name": "Port",
            "value": "4001"
        }
        {
            "name": "TLS",
            "value": "false"
        }
    ],
}
```

OAuth2 background information:

- [OAuth 2 Client Credentials](https://aaronparecki.com/oauth-2-simplified/#client-credentials)
- [An introduction to OAuth](https://www.digitalocean.com/community/tutorials/an-introduction-to-oauth-2)
- [OAuth 2.0 RFC](https://tools.ietf.org/html/rfc6749)
- [OAuth 2.0 Bearer Token Usage RFC](https://tools.ietf.org/html/rfc6750)

### Caching

Envoyauth has a built in-memory cache for retrieved entities from Cassandra. This will prevent doing Cassandra queries for entities that has already been retrieved earlier to speed up authentication requests.

### Logfiles

Envoyauth writes multiple logfiles, one for each function of envoyauth. All are written as structured JSON, filename rotation schedule can be set via configuration file. The three logfiles are:

1. `logging.filename` as log for application messages
2. `webadmin.logging.filename` as access log for REST API calls
3. `oauth2.logging.filename` as access log OAuth2 token calls

### Envoyauth configuration file

The supported fields are:

| yaml field                  | purpose                                          | example            |
| --------------------------- | ------------------------------------------------ | ------------------ |
| logging.level               | Application log level                            | info / debug       |
| logging.filename            | Filename to write application log to             | /dev/stdout        |
| logging.maxsize             | Maximum size in megabytes before rotate          | 100                |
| logging.maxage              | Max days to retain old log files                 | 7                  |
| logging.maxbackups          | Maximum number of old log files to retain        | 14                 |
| envoyauth.listen            | Address and port for authentication requests     | 0.0.0.0:4000       |
| webadmin.listen             | Webadmin address and port                        | 0.0.0.0:2113       |
| webadmin.ipacl              | Webadmin ip acl, without this no access          | 172.16.0.0/19      |
| webadmin.tls.certfile       | TLS certificate file                             |                    |
| webadmin.tls.keyfile        | TLS certificate key file                         |                    |
| webadmin.logging.level      | Logging level of webadmin                        | info / debug       |
| webadmin.logging.filename   | Filename to write web access log to              | dbadmin-access.log |
| webadmin.logging.maxsize    | Maximum size in megabytes before rotate          | 100                |
| webadmin.logging.maxage     | Max days to retain old log files                 | 7                  |
| webadmin.logging.maxbackups | Maximum number of old log files to retain        | 14                 |
| oauth.listen                | Listen address and port for OAuth token requests | 0.0.0.0:4001       |
| oauth.tls.certfile          | TLS certificate file                             |                    |
| oauth.tls.keyfile           | TLS certificate key file                         |                    |
| oauth.logging.level         | Logging level of oauth endpoint                  | info / debug       |
| oauth.logging.filename      | Filename to write oauth token access log to      | dbadmin-access.log |
| oauth.logging.maxsize       | Maximum size in megabytes before rotate          | 100                |
| oauth.logging.maxage        | Max days to retain old log files                 | 7                  |
| oauth.logging.maxbackups    | Maximum number of old log files to retain        | 14                 |
| oauth.tokenissuepath        | Path for OAuth2 token issue requests             | /oauth2/token      |
| oauth.tokeninfopath         | Path for OAuth2 token info requests              | /oauth2/info       |
| database.hostname           | Cassandra hostname to connect to                 | cassandra          |
| database.port               | Cassandra port to connect on                     | 9042 / 10350       |
| database.tls                | Enable TLS for database session                  | true / false       |
| database.username           | Database username                                | cassandra          |
| database.password           | Database password                                | cassandra          |
| database.keyspace           | Database keyspace for Gatekeeper tables          | gatekeeper         |
| database.timeout            | Timeout for session                              | 0.5s               |
| database.connectattempts    | Number of attempts to establish connection       | 5                  |
| database.queryretries       | Number of times to retry query                   | 2                  |
| cache.size                  | In-memory cache size in bytes                    | 1048576            |
| cache.ttl                   | Time-to-live for cached objects in seconds       | 15                 |
| cache.negativettl           | Time-to-live for non-existing objects in seconds | 15                 |
| maxmind.database            | Geoip database file                              |                    |

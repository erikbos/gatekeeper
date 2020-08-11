# Intro

Envoyauth's purpose is to authenticate for envoyproxy. Incoming requests to envoyproxy will first be send to envoyauth for authentication and authorization.

## Authentication

Authentication works by XXXX

## Envoyauth endpoints

Envoyauth exposes multiple endpoints:

| name      | scope   | protocol | purpose                                |
| --------- | ------- | -------- | -------------------------------------- |
| webadmin  | private | http     | admin console, prometheus metrics, etc |
| envoyauth | private | grpc     | authentication requests by envoyproxy  |
| oauth     | public  | http     | OAuth2 request for access tokens       |

For each there is a corresponding _endpoint.listen_ config field option to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself.
- _public_, meant to be exposed on public internet.

## Deployment

### Configuration

Envoyauth requires a starup configuration which needs to be provided as YAML file, see below for the supported fields. See [config/envoyauth-config-default.yaml[(For an example envoyauth configuration file).

The configuration of envoyproxy to have it forward requests to envoyauth for authentication is done by envoycp.

### Envoycp

Envoyproxy's authentication configuration is pushed by envoycp, envoycp's configuration section _xds.extauthz_ determine whether it's enabled, what the clustername, timeouts, etc.

The envoycp section _xds.extauthz_ contains parameters like:

- _enabled_, envoyproxy-wide switch to enable or disable forwarding requests to envoyauth
- _cluster_, name of cluster which runs envoyauth
- _timeout_, maximum allowed duration of authentication requests to envoyauth
- _failuremodeallow_, in case envoyauth fails should requests be forwarded or not
- _requestbodysize_, number of bytes of the request body envoyproxy should forward to envoyauth. This is a prerequisite for envoyauth being able to inspect the body of a request.

These configuration paramters get pushed out to every new envoyproxy instance.

#### Disable authentication for a route

Envoyproxy can be configured to skip authentication for a specific route. This is done by adding the attribute "DisableAthentication" with value "true" for a specific route.

### Caching

Envoyauth has a built in-memory cache for retrieved data from Cassandra. This will prevent doing Cassandra queries for contents which has already been retrieved earlier and speed up authentication requests.

### Envoyauth configuration file

The supported fields are:

| yaml field               | purpose                                                             | example                |
| ------------------------ | ------------------------------------------------------------------- | ---------------------- |
| webadmin.listen          | webadmin address and port                                           | 0.0.0.0:2113           |
| webadmin.ipacl           | webadmin ip acl, without this no access                             | 172.16.0.0/19          |
| webadmin.logfilename     | Filename of Filename of webadmin access log                         | /var/log/envoyauth.log |
| envoyauth.listen         | listen address and port for authentication requests from Envoyproxy | 0.0.0.0:4000           |
| oauth.listen             | listen address and port for OAuth token requests                    | 0.0.0.0:4001           |
| oauth.tokenissuepath     | Path for OAuth token issue requests                                 | /oauth2/token          |
| oauth.tokeninfopath      | Path for OAuth token info requests                                  | /oauth2/info           |
| database.hostname        | Cassandra hostname to connect to                                    | cassandra              |
| database.port            | Cassandra port to connect on                                        | 9042 / 10350           |
| database.tls             | Enable TLS for database session                                     | true / false           |
| database.username        | Database username                                                   | cassandra              |
| database.password        | Database password                                                   | cassandra              |
| database.keyspace        | Database keyspace for Gatekeeper tables                             | gatekeeper             |
| database.timeout         | Timeout for session                                                 | 0.5s                   |
| database.connectattempts | Number of attempts to establish connection                          | 5                      |
| database.queryretries    | Number of times to retry query                                      | 2                      |
| cache.size               | in-memory cache size in bytes                                       | 1048576                |
| cache.ttl                | time-to-live for cached objects in seconds                          | 15                     |
| cache.negativettl        | time-to-live for non-existing objects in seconds                    | 15                     |
| maxmind.filename         | geoip database filename                                             |                        |

# Intro

Envoycp is the control plane for envoyproxy. It configures all listeners, virtualhosts, routes and their specifics in envoyproxy.

## How does it work

Envoycp continously monitors the database for updates changes to virtualhosts, routes and clusters. In case there is a change a new envoyproxy configuration will be compiled and pushed to all envoyproxy.

## Envoycp endpoints

Envoycp exposes two endpoints:

| name     | scope   | protocol | purpose                                |
| -------- | ------- | -------- | -------------------------------------- |
| webadmin | private | http     | admin console, prometheus metrics, etc |
| xds      | private | grpc     | xds configuration with envoyproxy      |

For each there is a corresponding _*.listen_ config field option to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself
- _public_, meant to be publicly on public internet

## Deployment

### Configuration

Envoycp requires a starup configuration which needs to be provided as YAML file, see below for the supported fields. See
[config/envoycp-config-default.yaml[(For an example envoycp configuration file).

### Envoycp configuration file

The supported fields are:

| yaml field                     | purpose                                                                             | example                       |
| ------------------------------ | ----------------------------------------------------------------------------------- | ----------------------------- |
| webadmin.listen                | Webadmin address and port                                                           | 0.0.0.0:2113                  |
| webadmin.ipacl                 | Webadmin ip acl, without this no access                                             | 172.16.0.0/19                 |
| webadmin.logfile               | Filename of webadmin access log                                                     | /var/log/envoyproxy.log       |
| database.hostname              | Cassandra hostname to connect to                                                    | cassandra                     |
| database.port                  | Cassandra port to connect on                                                        | 9042 / 10350                  |
| database.tls                   | Enable TLS for database session                                                     | true / false                  |
| database.username              | Database username                                                                   | cassandra                     |
| database.password              | Database password                                                                   | cassandra                     |
| database.keyspace              | Database keyspace for Gatekeeper tables                                             | gatekeeper                    |
| database.timeout               | Timeout for session                                                                 | 0.5s                          |
| xds.grpclisten                 | listen address and port for XDS requests from Envoyproxy                            | 0.0.0.0:9901                  |
| xds.configpushinterval         | Frequency of checking changes to configuration                                      | 1s                            |
| xds.extauthz.enabled           | Envoyproxy-wide switch to enable or disable request authentication                  | true / false                  |
| xds.extauthz.cluster           | Name of cluster which runs envoyauth                                                |                               |
| xds.extauthz.timeout           | Maximum allowed duration of authentication requests to envoyauth                    |                               |
| xds.extauthz.failuremodeallow  | In case envoyauth does not answer in time/fails should requests be forwarded or not | true / false                  |
| xds.extauthz.requestbodysize   | Number of bytes of request body envoyproxy should forward to envoyauth              | 300                           |
| xds.envoy.logging.grpc.cluster | if set, configure envoyproxy to stream accesslog to this cluster                    | accesslogcluster              |
| xds.envoy.logging.grpc.logname | envoyproxy's logname when streaming accesslogs                                      | proxy                         |
| xds.envoy.logging.file.path    | if set, config envoyproxy to write accesslogs to this local file                    | /var/log/envoyproxyaccess.log |
| xds.envoy.logging.file.fields | json array of [fields to log](https://www.envoyproxy.io/docs/envoy/latest/configuration/observability/access_log/usage) | [Example field logging config](/example/deployment/docker/envoycp.yaml) |

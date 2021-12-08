# Intro

Envoyals is the access logger receiver for envoyproxy.

## how to enable

Set the attribute [AccessLogCluster](api/listener.md#Attribute) on a listener to enable envoyproxy to stream accesslogs.

## Envoyals endpoints

Envoyals exposes two endpoints:

| name      | scope   | protocol | purpose                                |
| --------- | ------- | -------- | -------------------------------------- |
| webadmin  | private | http     | admin console, prometheus metrics, etc |
| accesslog | private | grpc     | access log receive from envoyproxy     |

Both `webadmin.listen` and `accesslog.listen` should be set to configure listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself

## Deployment

### Configuration

Envoyals requires a starup configuration which needs to be provided as YAML file, see below for the supported fields. For an example configuration file see [envoyals.yaml](../deployment/docker/envoyals.yaml).

### Logfiles

Envoyals writes multiple logfiles, one for each function of envoyals. All are written as structured JSON, filename rotation schedule can be set via configuration file. The three logfiles are:

1. `logging.filename` as log for application messages
2. `webadmin.logging.filename` as access log for all REST API calls
3. `accesslog.logging.filename` as access log entries received from envoy

### Envoyals configuration file

The supported fields are:

| yaml field                   | purpose                                   | example            |
| ---------------------------- | ----------------------------------------- | ------------------ |
| logging.level                | application log level                     | info / debug       |
| logging.filename             | filename to write application log to      | /dev/stdout        |
| logging.maxsize              | Maximum size in megabytes before rotate   | 100                |
| logging.maxage               | Max days to retain old log files          | 7                  |
| logging.maxbackups           | Maximum number of old log files to retain | 14                 |
| webadmin.listen              | Webadmin address and port                 | 0.0.0.0:6002       |
| webadmin.ipacl               | Webadmin ip acl, without this no access   | 172.16.0.0/19      |
| webadmin.tls.certfile        | TLS certificate file                      |                    |
| webadmin.tls.keyfile         | TLS certificate key file                  |                    |
| webadmin.logging.level       | logging level of webadmin                 | info / debug       |
| webadmin.logging.filename    | filename to write web access log to       | managementserver-access.log |
| webadmin.logging.maxsize     | Maximum size in megabytes before rotate   | 100                |
| webadmin.logging.maxage      | Max days to retain old log files          | 7                  |
| webadmin.logging.maxbackups  | Maximum number of old log files to retain | 14                 |
| accesslog.listen             | accesslog address and port                | 0.0.0.0:6001       |
| accesslog.maxstreamduration  | max envoy stream duration                 | 10m                |
| accesslog.logging.level      | accesslog log level, not used             | info / debug       |
| accesslog.logging.filename   | filename to write access log to           | /dev/stdout        |
| accesslog.logging.maxsize    | Maximum size in megabytes before rotate   | 100                |
| accesslog.logging.maxage     | Max days to retain old log files          | 7                  |
| accesslog.logging.maxbackups | Maximum number of old log files to retain | 14                 |

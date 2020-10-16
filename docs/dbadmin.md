# Intro

DBadmin provides a REST API to create, read, update, delete entities of Gatekeeper's configuration database.

For each entity a detailed API description is available:

1. [listeners](docs/api/listener.md)
2. [routes](docs/api/route.md)
3. [clusters](docs/api/cluster.md)
4. [organizations](docs/api/organization.md)
5. [developers](docs/api/developer.md)
6. [developer apps](docs/api/developerapp.md)
7. [key](docs/api/key.md)
8. [apiproduct](docs/api/apiproduct.md)

## dbadmin endpoints

Dbadmin exposes one endpoint:

| name     | scope   | protocol | purpose                                                     |
| -------- | ------- | -------- | ----------------------------------------------------------- |
| webadmin | private | http     | create, read, update, delete for all configuration entities |
|          |         |          | admin console, prometheus metrics, etc                      |

The `webadmin.listen` config field should be used to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself.

## Deployment

### Configuration

Dbadmin supports the following start up command line arguments:

| argument | purpose                                                 | example                                            |
| -------- | ------------------------------------------------------- | -------------------------------------------------- |
| config   | startup configuration file, see below supported fields. | [dbadmin.yaml](../deployment/docker/dbadmin.yaml)  |
| createschema | Create database schema and tables at startup, if these do not exist | |
| replicacount | Use to indicate replica count for keyspace when using `createschema` | 3 |
| enableapiauthentication | Enable REST API authentication, [users](api/user.md) and [roles](api/route.md) will be used for access and authorization `/v1/` path |

### Logfiles

Dbadmin writes multiple logfiles, one for each function of dbadmin. All are written as structured JSON, filename rotation schedule can be set via configuration file. The three logfiles are:

1. `logging.filename` as log for application messages
2. `webadmin.logging.filename` as access log for all REST API calls
3. `changelog.logging.filename` as entity changelog, all CRUD-operations, it logs full entity details so it might contain sensitive information!

### Dbadmin configuration file

The supported fields are:

| yaml field                   | purpose                                    | example               |
| ---------------------------- | ------------------------------------------ | --------------------- |
| logging.level                | Application log level                      | info / debug          |
| logging.filename             | Filename to write application log to       | /dev/stdout           |
| logging.maxsize              | Maximum size in megabytes before rotate    | 100                   |
| logging.maxage               | Max days to retain old log files           | 7                     |
| logging.maxbackups           | Maximum number of old log files to retain  | 14                    |
| webadmin.listen              | Webadmin address and port                  | 0.0.0.0:7777          |
| webadmin.ipacl               | Webadmin ip acl, without this no access    | 172.16.0.0/19         |
| webadmin.tls.certfile        | TLS certificate file                       |                       |
| webadmin.tls.keyfile         | TLS certificate key file                   |                       |
| webadmin.logging.level       | Logging level of webadmin                  | info / debug          |
| webadmin.logging.filename    | Filename to write web access log to        | dbadmin-access.log    |
| webadmin.logging.maxsize     | Maximum size in megabytes before rotate    | 100                   |
| webadmin.logging.maxage      | Max days to retain old log files           | 7                     |
| webadmin.logging.maxbackups  | Maximum number of old log files to retain  | 14                    |
| changelog.logging.level      | Logging level of changelog                 | info                  |
| changelog.logging.filename   | Filename to write changed entities to      | dbadmin-changelog.log |
| changelog.logging.maxsize    | Maximum size in megabytes before rotate    | 100                   |
| changelog.logging.maxage     | Max days to retain old log files           | 7                     |
| changelog.logging.maxbackups | Maximum number of old log files to retain  | 14                    |
| database.hostname            | Cassandra hostname to connect to           | cassandra             |
| database.port                | Cassandra port to connect on               | 9042 / 10350          |
| database.tls                 | Enable TLS for database session            | true / false          |
| database.username            | Database username                          | cassandra             |
| database.password            | Database password                          | cassandra             |
| database.keyspace            | Database keyspace for Gatekeeper tables    | gatekeeper            |
| database.timeout             | Timeout for session                        | 0.5s                  |
| database.connectattempts     | Number of attempts to establish connection | 5                     |
| database.queryretries        | Number of times to retry query             | 2                     |

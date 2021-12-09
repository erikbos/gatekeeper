# Intro

Managementserver provides a REST API to create, read, update, delete entities of Gatekeeper's configuration database.

For each entity a detailed API description is available:

1. [listeners](docs/api/listener.md)
2. [routes](docs/api/route.md)
3. [clusters](docs/api/cluster.md)
4. [developers](docs/api/developer.md)
5. [developer apps](docs/api/developerapp.md)
6. [key](docs/api/key.md)
7. [apiproduct](docs/api/apiproduct.md)
8. [user](docs/api/user.md)
9. [role](docs/api/role.md)

## Managementserver endpoints

Managementserver exposes one endpoint:

| name     | scope   | protocol | purpose                                                     |
| -------- | ------- | -------- | ----------------------------------------------------------- |
| webadmin | private | http     | create, read, update, delete for all configuration entities |
|          |         |          | admin console, prometheus metrics, etc                      |

The `webadmin.listen` config field should be used to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself.

## Deployment

### Configuration

Managementserver supports the following start up command line arguments:

| argument                 | purpose                                                  | example                                           |
| ------------------------ | -------------------------------------------------------- | ------------------------------------------------- |
| config                   | startup configuration file, see below supported fields.  | [managementserver.yaml](../deployment/docker/managementserver.yaml) |
| disableapiauthentication | Disable REST API authentication on /v1 path              |                                                   |
| showcreateschema         | Show CQL statements to create database                   |                                                   |
| createschema             | Create database schema and tables, if these do not exist |                                                   |
| replicacount             | Replica count for keyspace when using `createschema`     | 3                                                 |

### Application and API request logfiles

Managementserver writes multiple logfiles, one for each function of managementserver. All are written as structured JSON, filename rotation schedule can be set via configuration file.

The following configuration properties control application and API request logging:

1. `logging.*` to configure properties of application logging.
2. `webadmin.logging.*` to configure properties of logging REST API requests.

### Auditlogs

Managementserver keep track of all changes to entities, all create, update and delete operations are logged to a logfile and written to database. An [Audit](audit.md) API is available to retrieve an audit trail.

Both old and new value of a changed entity is logged to have a full detailed changed history.

The following configuration properties control audit logging:

1. `auditlog.logging.*` to configure all audit logfile properties such as filename, log rotation.
2. `auditlog.database.*` to configure which database to use to write audit log entries to.

### Managementserver configuration file

The supported fields are:

| yaml field                     | purpose                                    | example                        |
| ------------------------------ | ------------------------------------------ | ------------------------------ |
| logging.level                  | Application log level                      | info / debug                   |
| logging.filename               | Filename to write application log to       | /dev/stdout                    |
| logging.maxsize                | Maximum size in megabytes before rotate    | 100                            |
| logging.maxage                 | Max days to retain old log files           | 7                              |
| logging.maxbackups             | Maximum number of old log files to retain  | 14                             |
| webadmin.listen                | Webadmin address and port                  | 0.0.0.0:7777                   |
| webadmin.ipacl                 | Webadmin ip acl, without this no access    | 172.16.0.0/19                  |
| webadmin.tls.certfile          | TLS certificate file                       |                                |
| webadmin.tls.keyfile           | TLS certificate key file                   |                                |
| webadmin.logging.level         | Logging level of webadmin                  | info / debug                   |
| webadmin.logging.filename      | Filename to write web access log to        | managementserver-access.log    |
| webadmin.logging.maxsize       | Maximum size in megabytes before rotate    | 100                            |
| webadmin.logging.maxage        | Max days to retain old log files           | 7                              |
| webadmin.logging.maxbackups    | Maximum number of old log files to retain  | 14                             |
| database.hostname              | Cassandra hostname to connect to           | cassandra                      |
| database.port                  | Cassandra port to connect on               | 9042 / 10350                   |
| database.tls                   | Enable TLS for database session            | true / false                   |
| database.username              | Database username                          | cassandra                      |
| database.password              | Database password                          | cassandra                      |
| database.keyspace              | Database keyspace for Gatekeeper tables    | gatekeeper                     |
| database.timeout               | Timeout for session                        | 0.5s                           |
| database.connectattempts       | Number of attempts to establish connection | 5                              |
| database.queryretries          | Number of times to retry query             | 2                              |
| audit.logging.level            | Logging level of changelog                 | info                           |
| audit.logging.filename         | Filename to write changed entities to      | managementserver-changelog.log |
| audit.logging.maxsize          | Maximum size in megabytes before rotate    | 100                            |
| audit.logging.maxage           | Max days to retain old log files           | 7                              |
| audit.logging.maxbackups       | Maximum number of old log files to retain  | 14                             |
| audit.database.hostname        | Cassandra hostname to connect to           | cassandra                      |
| audit.database.port            | Cassandra port to connect on               | 9042 / 10350                   |
| audit.database.tls             | Enable TLS for database session            | true / false                   |
| audit.database.username        | Database username                          | cassandra                      |
| audit.database.password        | Database password                          | cassandra                      |
| audit.database.keyspace        | Database keyspace for audit table          | gatekeeper                     |
| audit.database.timeout         | Timeout for session                        | 0.5s                           |
| audit.database.connectattempts | Number of attempts to establish connection | 5                              |
| audit.database.queryretries    | Number of times to retry query             | 2                              |

# Intro

DBadmin provides the REST API to create, read, update, delete all entities of Gatekeeper's configuration database.

For each entity there is a detailed API description available:

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
| webadmin | private | http     | admin console, prometheus metrics, etc                      |

The _webadmin.listen_ config field should be used to set listening address and port.

Scope:

- _private_, must not be exposed outside the deployment itself.

## Deployment

### Configuration

Dbadmin requires a starup configuration which needs to be provided as YAML file, see below for the supported fields, see [config/dbadmin-config-default.yaml[(For an example dbadmin configuration file).

### Dbadmin configuration file

The supported fields are:

| yaml field               | purpose                                    | example              |
| ------------------------ | ------------------------------------------ | -------------------- |
| loglevel                 | logging level of application               | info / debug         |
| webadmin.listen          | Webadmin address and port                  | 0.0.0.0:7777         |
| webadmin.ipacl           | Webadmin ip acl, without this no access    | 172.16.0.0/19        |
| webadmin.certfile        | TLS certificate file                       |                      |
| webadmin.keyfile         | TLS certificate key file                   |                      |
| webadmin.logfile         | Access log file                            | /var/log/dbadmin.log |
| database.hostname        | Cassandra hostname to connect to           | cassandra            |
| database.port            | Cassandra port to connect on               | 9042 / 10350         |
| database.tls             | Enable TLS for database session            | true / false         |
| database.username        | Database username                          | cassandra            |
| database.password        | Database password                          | cassandra            |
| database.keyspace        | Database keyspace for Gatekeeper tables    | gatekeeper           |
| database.timeout         | Timeout for session                        | 0.5s                 |
| database.connectattempts | Number of attempts to establish connection | 5                    |
| database.queryretries    | Number of times to retry query             | 2                    |

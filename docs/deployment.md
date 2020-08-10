# Deployment guide

## Build containers

Clone repository:
```sh
git clone http://github.com/erikbos/gatekeeper
cd gatekeeper
```

Build Gatekeeper images:

```sh
make docker-images
```
to build Containers for [dbadmin](docs/dbadmin.md), [envoyauth](docs/envoyauth.md), [envoycp](docs/envoycp.md) and [testbackend](docs/testbackend.md).

## Deploy Gatekeeper

### Docker compose

The following starts all containers for Gatekeeper using compose: one-node Cassandra instance, envoyproxy, envoyauth, envoycp and dbadmin
```sh
docker-compose -f examples/deployment/docker/cassandra.yaml up
```

Please note:

* At startup all relevant database tables will be created (using  [Cassandra_create_tables.cql](scripts/Cassandra_create_tables.cql))
* Cassandra will be configured a single node database cluster by changing its replication factor down to 1 (Using [Cassandra_create_tables.cql](scripts/Cassandra_switch_to_one_node.cql))
* To persist the database across restarts the directory /tmp/cassandra_data is used.
* Cassandra runs as a single node database cluster in a container we need to change its replication factor down to 1
* All containers start at the same time (compose does not support waits) as Cassandra takes 30 seconds te start all other containers will warn about database unavailablity initially.

### Kubernetes

todo

### AWS ECS

todo

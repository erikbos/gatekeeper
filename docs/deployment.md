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
docker-compose -f examples/deployment/docker/gatekeeper.yaml up
```

Please note:

* At startup database schema will be created (Using dbadmin cmdline argument *--create-tables*)
* Database will be configured as a single node by changing its replication count to 1 (Using dbadmin cmdline argument *--create-tables*
* For production do not use replication count of 1(!)
* To persist the database across restarts the directory /tmp/cassandra_data is used.
* All containers start at the same time (compose does not support waits) as Cassandra takes 30 seconds te start all other containers might warn about not yet being able to connect to database. 

### Kubernetes

todo

### AWS ECS

todo

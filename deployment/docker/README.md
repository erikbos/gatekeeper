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

to build Containers for [managementserver](docs/managementserver.md), [authserver](docs/authserver.md), [controlplane](docs/controlplane.md) and [testbackend](docs/testbackend.md).

## Deploy Gatekeeper

### Docker compose

The following starts all containers for Gatekeeper using compose: management server, one-node Cassandra instance, envoyproxy, authserver, controlplane

```sh
docker-compose -f deployment/docker/gatekeeper.yaml up
```

Please note:

* At startup database schema will be created (Similar to involing managementserver cmdline argument *--create-tables*)
* Database will be configured as a single node by changing its replication count to 1 (Using managementserver cmdline argument *--create-tables*
* For production do not use replication count of 1(!)
* To persist the database across restarts the directory /tmp/cassandra_data is used.
* All containers start at the same time (compose does not support waits) as Cassandra takes 30 seconds te start all other containers might warn about not yet being able to connect to database.

### Azure

Use the following steps to deploy on Azure:

Create a Cassandra database

```sh
location="westus"
rg="gatekeeper"
dbaccount="gatekeeper-$RANDOM"

# Create resource group
az group create --name $rg --location $location

# Create Cassandra database (Cososmdb with Cassandra API)
az cosmosdb create --resource-group $rg --name $dbaccount --capabilities EnableCassandra --locations regionName=$location
#az cosmosdb create --resource-group $rg --name $dbaccount --capabilities EnableCassandra EnableServerless --locations regionName=$location
```

Get the username and password from the connection string _Primary
Cassandra Connection String_.

```sh
# Get connect details
az cosmosdb keys list --type connection-strings --resource-group $rg --name $dbaccount
```

To deploy Gatekeeper using [container images published on Github](https://github.com/erikbos?tab=packages)

```sh
ns="test2"

# create namespace
kubectl create namespace $ns

# switch kubectl context to new namespace
 kubectl config set-context $(kubectl config current-context) --namespace=$ns

# install gatekeeper in namespace
helm install gatekeeper ./helm --wait --namespace $ns -f helm/values.yaml \
        --set global.e2e=true \
        --set global.images.tag=$(VERSION) \
        --set cassandra.persistance='{"enabled":"true"}'
```

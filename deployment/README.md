# Deployment guide

Gatekeeper consits out of multiple containers [managementserver](docs/managementserver.md), [authserver](docs/authserver.md), [controlplane](docs/controlplane.md) [accesslogserver](docs/accesslogserver.md) and [testbackend](docs/testbackend.md).

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

## Deploy Gatekeeper

### local deployment using Docker compose

The following starts Gatekeeper's containers using compose: management server, one-node Cassandra instance, envoyproxy, authserver, accesslogserver and controlplane.

```sh
docker-compose -f deployment/docker/docker-compose.yaml up
```

Please note:

* At startup database schema will be created (Similar to involing managementserver cmdline argument *--create-tables*)
* Database will be configured as a single node by changing its replication count to 1 (Using managementserver cmdline argument *--create-tables*
* For production do not use replication count of 1(!)
* To persist the database across restarts the directory /tmp/cassandra_data is used.
* All containers start at the same time (compose does not support waits) as Cassandra takes 30 seconds te start all other containers might warn about not yet being able to connect to database.

### Kubernetes and Azure Cosmos DB

Use the following steps to deploy on Azure:

Create a CosmosDB database with Cassandra API:

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

Get database username and password from the connection string *Primary Cassandra Connection String:*

```sh
# Get connect details
az cosmosdb keys list --type connection-strings --resource-group $rg --name $dbaccount
```

Deploy Gatekeeper using [container images published on Github](https://github.com/erikbos?tab=packages)

```sh
ns="test2"

# create namespace
kubectl create namespace $ns

# switch kubectl context to new namespace
 kubectl config set-context $(kubectl config current-context) --namespace=$ns

# TODO
# insert steps to pass database credentials as helm parameters

# install gatekeeper in namespace
helm install gatekeeper ./helm --wait --namespace $ns -f helm/values.yaml \
        --set global.e2e=true \
        --set global.images.tag=$(VERSION) \
        --set cassandra.persistance='{"enabled":"true"}'
```

# Gatekeeper documentation

## Architecture

high-level [architecture](architecture.md) overview

## Deployment

Gatekeeper consists out of three components (next to Envoyproxy):

1. [Dbadmin](dbadmin.md) provides management API to configure all entities in Gatekeeper
2. [Envoycp](envoycp.md) is control plane for Envoyproxy
3. [Envoyauth](envoyauth.md) is (optional) authentication server for Envoyproxy

## Management API

Gatekeeper has an elaborate [Create Read Update Delete](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) REST-based management API to manage all entities in the system. All entities are defined using JSON.

For requesting forwarding the following entities need to be configured:

1. [Listeners](api/listener.md) to configure tcp listening port, virtual hosts and TLS parameters
2. [Routes](api/route.md), to set http paths and corresponding upstream clusters
3. [Clusters](api/cluster.md) to define how to connect to a backend cluster.

For authentication of requests:

1. [organizations](api/organization.md)
2. [developers](api/developer.md)
3. [developer apps](api/developerapp.md)
4. [key](api/key.md)
5. [apiproduct](api/apiproduct.md)

Example API calls can be found in [examples](api/examples)

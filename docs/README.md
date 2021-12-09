# Gatekeeper documentation

## Architecture

high-level [architecture](architecture.md) overview

## Deployment

Gatekeeper consists out of three components (next to Envoyproxy):

1. [Managementserver](managementserver.md) provides management API to configure all entities in Gatekeeper
2. [Controlplane](controlplane.md) is control plane for Envoyproxy
3. [Authserver](authserver.md) is authentication server for Envoyproxy
4. [Accesslogserver](accesslogserver.md) is access logging server for Envoyproxy

## Management API

Managementserver provides the [Create Read Update Delete](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) REST-based management API to manage all entities in the system. All entities are defined using JSON.

Directory [API](api/README.md) contains detailed documentation for _listeners_, _routes_, _clusters_, _developers_, _developer apps_, _keys_, _apiproducts_, _users_ and _roles_.

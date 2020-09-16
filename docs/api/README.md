# Management API

Gatekeeper has an elaborate [Create Read Update Delete](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) REST-based management API to manage all entities in the system. All entities are defined using JSON.

For requesting forwarding the following entities need to be configured:

1. [Listeners](listener.md) to configure tcp listening port, virtual hosts and TLS parameters
2. [Routes](route.md), to set http paths and corresponding upstream clusters
3. [Clusters](cluster.md) to define how to connect to a backend cluster.

For authentication of requests:

1. [organizations](organization.md)
2. [developers](developer.md)
3. [developer apps](developerapp.md)
4. [key](key.md)
5. [apiproduct](apiproduct.md)

Example API calls can be found in [examples](examples)

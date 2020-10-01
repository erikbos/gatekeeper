# Management API

Gatekeeper has an elaborate [Create Read Update Delete](https://en.wikipedia.org/wiki/Create,_read,_update_and_delete) REST-based management API to manage all entities in the system. All entities are defined using JSON, authentication of this API is controled using [Users](user.md) and [Roles](role.md) entities.

For requesting forwarding the following entities need to be configured:

1. [Listeners](listener.md) to configure tcp listening port, virtual hosts and TLS parameters
2. [Routes](route.md), to set http paths and corresponding upstream clusters
3. [Clusters](cluster.md) to define how to connect to a backend cluster

For authentication of requests:

1. [Organizations](organization.md)
2. [Developers](developer.md)
3. [Developer apps](developerapp.md)
4. [Key](key.md)
5. [Aproduct](apiproduct.md)

Example API calls can be found in [examples](examples)

# Gatekeeper

Gatekeeper is an Gatekeeper proxy with API entitlement functionality. It uses [Envoy](https://www.envoyproxy.io/) as a proxy.

## Table of Contents

* [Introduction](#introduction)
* [Understanding](#understanding)
* [Getting Started](#getting-started)
* [Repository Structure](#repository-structure)
* [Contributing and Support](#contributing-and-support)
* [License](#license)
* [Disclaimer](#disclaimer)

## Introduction

Gatekeeper provides API entitlement management:

* Administration of API [developers](docs/api/developer.MD) and their [applications](docs/api/developerapp.MD).

* Fine grained access control to backends by defining [API products](docs/api/apiproduct.MD).

* Authenticate and authorize developer applications using [API Keys and/or OAuth2](docs/api/key.MD).

For backend teams:

* Gatekeeper's authentication has detailed metrics on authentication and authorization

* Envoyproxy offer detailed request metrics on error rates, response latencies, request size.

* Security: Gatekeeper supports TLS for [downstream](docs/api/virtualhost.MD) and [upstream](docs/api/cluster.MD) traffic.

* Dynamic Routing: Gatekeeper can [route](docs/api/route.MD) traffic across multiple backends.

* High Availbility: by allowing [retry behaviour](docs/api/route.MD) to be configured per path to reduce error rates.

* Health Checks: Gatekeeper can actively [monitor](docs/api/cluster.MD) backends.

* Ease of deployment: it consists out 3 containers and any Cassandra database as backend.


## Understanding Gatekeeper

* [Use Cases](doc/use-cases.md)

* [Architecture](doc/architecture.md)

## Repository structure

* [build](build): Scripts for packaging Gatekeeper components in a Docker images

* [config](config): Example configuration files for Gatekeeper components

* [docs](docs): Extended documentation (use cases, architecture, api specs, etc.)

* [examples](examples): Example management API calls to configure Gatekeeper

* [scripts](scripts): Scripts to deploy Gatekeeper

* [src](src): Source code for all Gatekeeper components

## Contributing and support

Please note Gatekeeper is still under heavy development, but feel free to open a Github issue!

## License

[Apache v2](LICENSE)

Definition of Gatekeeper's API is inspired by work [created and shared by Google](https://docs.apigee.com/reference/apis/apigee/rest/v1/) and used according to terms described in the [Creative Commons 4.0 Attribution License](https://creativecommons.org/licenses/by/4.0/)

## Disclaimer

Gatekeeper is current under heavy development.

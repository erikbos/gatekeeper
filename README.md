# Gatekeeper

[![Build pipeline](https://github.com/erikbos/gatekeeper/workflows/Build%20pipeline/badge.svg)](https://github.com/erikbos/gatekeeper/actions?query=workflow%3A%22Build+pipeline%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/erikbos/gatekeeper)](https://goreportcard.com/report/github.com/erikbos/gatekeeper)

Gatekeeper is an Gatekeeper proxy with rich API entitlement functionality. It uses [Envoy](https://www.envoyproxy.io/) as a proxy.

## Table of Contents

* [Introduction](#introduction)
* [Getting Started](#getting-started)
* [Repository Structure](#repository-structure)
* [Contributing and Support](#contributing-and-support)
* [License](#license)
* [Disclaimer](#disclaimer)

## Introduction

Gatekeeper offers a simple REST [api](docs/api/README.md) to unlock Envoyproxy's advancing routing capabilities:

* Gatekeeper supports TLS for [downstream](docs/api/listener.md) and [upstream](docs/api/cluster.md) traffic.

* Dynamic Routing: Gatekeeper can [route](docs/api/route.md) traffic across multiple backends.

* High Availability: by allowing [retry behaviour](docs/api/route.md) to be configured per path to reduce error rates.

* Health Checks: Gatekeeper can actively [monitor](docs/api/cluster.md) backends.

* Gatekeeper's authentication has detailed metrics on authentication and authorization.

* Envoyproxy offer detailed request metrics on error rates, response latencies, request size.

* Ease of deployment: it consists out 3 containers and a Cassandra database as data store.

* Designed for multi cloud region deployment by default.

Gatekeeper provides API entitlement management:

* Administration of API [developers](docs/api/developer.md) and their [applications](docs/api/developerapp.md).

* Fine grained access control to backends by defining [API products](docs/api/apiproduct.md).

* Authenticate and authorize developer applications using [API Keys and/or OAuth2](docs/api/key.md).

## Repository structure

* [docs](docs): All documentation:

  * [Architecture](docs/deployment/architecture.md) High-level overview.

  * [deployment](docs/): Deployment documentation of each component.

  * [api](docs/api/): management API specification.

  * [examples](docs/api/examples/): Example management API calls.

* [build](build): Scripts for packaging Gatekeeper components in Docker images.

* [deployment](deployment/docker/): example Docker compose configuration.

* [cmd](cmd): Source code of individual Gatekeeper components.

* [pkg](pkg): Source code of shared Gatekeeper components.

## Contributing and support

Please note Gatekeeper is still under heavy development, but feel free to open a Github issue!

## License

[Apache v2](LICENSE), some of Gatekeeper's API is inspired by work [created and shared by Google](https://docs.apigee.com/reference/apis/apigee/rest/) and used according to terms described in the [Creative Commons 4.0 Attribution License](https://creativecommons.org/licenses/by/4.0/)

## Disclaimer

Gatekeeper is current under heavy development.

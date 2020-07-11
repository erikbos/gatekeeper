# Gatekeeper

[![Build pipeline](https://github.com/erikbos/gatekeeper/workflows/Build%20pipeline/badge.svg)](https://github.com/erikbos/gatekeeper/actions?query=workflow%3A%22Build+pipeline%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/erikbos/gatekeeper)](https://goreportcard.com/report/github.com/erikbos/gatekeeper)

Gatekeeper is an Gatekeeper proxy with rich API entitlement functionality. It uses [Envoy](https://www.envoyproxy.io/) as a proxy.

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

* Administration of API [developers](docs/api/developer.md) and their [applications](docs/api/developerapp.md).

* Fine grained access control to backends by defining [API products](docs/api/apiproduct.md).

* Authenticate and authorize developer applications using [API Keys and/or OAuth2](docs/api/key.md).

For backend teams:

* Gatekeeper's authentication has detailed metrics on authentication and authorization

* Envoyproxy offer detailed request metrics on error rates, response latencies, request size.

* Security: Gatekeeper supports TLS for [downstream](docs/api/virtualhost.md) and [upstream](docs/api/cluster.md) traffic.

* Dynamic Routing: Gatekeeper can [route](docs/api/route.md) traffic across multiple backends.

* High Availbility: by allowing [retry behaviour](docs/api/route.md) to be configured per path to reduce error rates.

* Health Checks: Gatekeeper can actively [monitor](docs/api/cluster.md) backends.

* Ease of deployment: it consists out 3 containers and a Cassandra database as backend.

## Understanding Gatekeeper

* [Use Cases](doc/use-cases.md)

* [Architecture](doc/architecture.md)

## Repository structure

* [build](build): Scripts for packaging Gatekeeper components in a Docker images

* [configs](configs): Example configuration files for Gatekeeper components

* [docs](docs): Extended documentation (use cases, architecture, api specs, etc.)

* [examples](examples): Example management API calls to configure Gatekeeper

* [scripts](scripts): Scripts to deploy Gatekeeper

* [cmd](cmd): Source code of individual Gatekeeper components

* [pkg](pkg): Source code of shared Gatekeeper components

## Contributing and support

Please note Gatekeeper is still under heavy development, but feel free to open a Github issue!

## License

[Apache v2](LICENSE), some of Gatekeeper's API is inspired by work [created and shared by Google](https://docs.apigee.com/reference/apis/apigee/rest/) and used according to terms described in the [Creative Commons 4.0 Attribution License](https://creativecommons.org/licenses/by/4.0/)

## Disclaimer

Gatekeeper is current under heavy development.

# Role

A role defines an role in the system. `dbadmin` uses both `user` and `role` to determine whether an API call to any endpoint under `/v1/**` endpoint should be allowed or not.

## Supported operations

| Method | Path                 | What                     |
| ------ | -------------------- | ------------------------ |
| GET    | /v1/roles            | retrieve all roles       |
| POST   | /v1/roles            | creates a new role       |
| GET    | /v1/roles/_rolename_ | retrieve a role          |
| POST   | /v1/roles/_rolename_ | updates an existing role |
| DELETE | /v1/roles/_rolename_ | delete role              |

_For POST content-type: application/json is required._

## Example role entity

Role `admin` that allows HTTP method GET, POST or DELETE for all (`**`) paths under `/v1`, this allows all entities to read and updated:

```json
{
    "name": "admin",
    "displayName": "Administrator",
    "allows": [
        {
            "methods": [
                "GET",
                "POST",
                "DELETE"
            ],
            "paths": [
                "/v1/**"
            ]
        }
        ]
}
```

## Default role

The above role `admin` is created when `dbadmin` is invoked with `createschema`.

## Fields specification

| fieldname      | optional  | purpose                                                                              |
| -------------- | --------- | ------------------------------------------------------------------------------------ |
| name           | mandatory | Name (cannot be updated afterwards)                                                  |
| displayName    | optional  | Friendly name                                                                        |
| allows         | mandatory | Array with allowed methods & paths                                                   |
| allows.methods | optional  | Allowed methods                                                                      |
| allows.paths   | optional  | Allowed paths, supports [doublestar](https://github.com/bmatcuk/doublestar#patterns) |
|                |           | `*` matches any sequence of non-path-separators                                      |
|                |           | `**` matches any sequence of characters, including path separators                   |

## Example roles configurations

Role `infra_readonly` allowing reading (GET) all listeners, routes and clusters configuration:

```json
{
    "name": "infra_readonly",
    "displayName": "Infra read only role",
    "allows": [
        {
            "methods": [
                "GET",
            ],
            "paths": [
                "/v1/listeners",
                "/v1/routes",
                "/v1/clusters"
            ]
        }
        ]
}
```

Role `route_update` allowing route `ticketshop` to be retrieved, and just the attribute `Cluster` to be updated.
This allows this role to update just the destination cluster of the this route:

```json
{
    "name": "route_update",
    "displayName": "",
    "allows": [
        {
            "methods": [
                "GET",
            ],
            "paths": [
                "/v1/routes/ticketshop"
            ]
        },
        {
            "methods": [
                "POST",
            ],
            "paths": [
                "/v1/routes/ticketshop/attributes/Cluster"
            ]
        }
        ]
}
```

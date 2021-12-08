# User

A user defines an user in the system. `managementserver` uses both `user` and `role` to determine whether an API call to any endpoint under `/v1/**` endpoint should be allowed or not.

## Supported operations

| Method | Path                 | What                     |
| ------ | -------------------- | ------------------------ |
| GET    | /v1/users            | retrieve all users       |
| POST   | /v1/users            | creates a new user       |
| GET    | /v1/users/_username_ | retrieve a user          |
| POST   | /v1/users/_username_ | updates an existing user |
| DELETE | /v1/users/_username_ | delete user              |

_For POST content-type: application/json is required._

## Example user entity

User `admin` with roles `admin` and password 'passwd':

```json
{
    "name": "admin",
    "displayName": "Administrator",
    "password": "passwd",
    "status": "active",
    "roles": [
        "admin",
    ]
}
```

## Default user

The above role `admin` is created when `managementserver` is invoked with `createschema`.

## Fields specification

| fieldname   | optional  | purpose                                               |
| ----------- | --------- | ----------------------------------------------------- |
| name        | mandatory | Name (cannot be updated afterwards)                   |
| displayName | optional  | Friendly name                                         |
| password    | mandatory | Password (will only be updated when field is present) |
| status      | mandatory | Active                                                |
| roles       | mandatory | Array of [roles](role.md) that this user has          |

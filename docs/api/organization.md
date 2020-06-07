# Organization

An organization is a collection of developers.

## Supported methods and paths

| Method | Path                     | What                             |
| ------ | ------------------------ | -------------------------------- |
| GET    | /v1/organizations        | retrieve all organizations       |
| POST   | /v1/organizations        | creates a new organization       |
| GET    | /v1/organizations/_name_ | retrieve an organization         |
| POST   | /v1/organizations/_name_ | updates an existing organization |
| DELETE | /v1/organizations/_name_ | deletes an organization          |

* For POST content-type: application/json is required.

## Example organization definition

```json
{
    "name": "petstore",
    "displayName": "Pet Store Inc"
}
```

## Fields specification

| fieldname   | optional  | purpose                             |
| ----------- | --------- | ----------------------------------- |
| name        | mandatory | name (cannot be updated afterwards) |
| displayName | optional  | friendly name                       |

## Background

Organizations are used to group developers together, they do not have any functional use.

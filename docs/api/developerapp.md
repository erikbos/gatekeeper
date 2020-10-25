# Developer application

A developer application holds the attribut.

## Supported methods and paths

| Method | Path                                                                           | What                                     |
| ------ | ------------------------------------------------------------------------------ | ---------------------------------------- |
| GET    | /v1/developers/_developer_/apps                             | retrieve all apps of developer           |
| POST   | /v1/developers/_developer_/apps                             | creates a new developer app              |
| GET    | /v1/developers/_developer_/apps/_appname_                   | retrieve one developer app               |
| POST   | /v1/developers/_developer_/apps/_appname_                   | updates an existing developer app        |
| DELETE | /v1/developers/_developer_/apps/_appname_                   | deletes a developer app                  |
| GET    | /v1/developers/_developer_/apps/_appname_/attributes        | retrieve all attributes of developer app |
| POST   | /v1/developers/_developer_/apps/_appname_/attributes        | update all attribute of developer app    |
| GET    | /v1/developers/_developer_/apps/_appname_/attributes/_name_ | retrieve attribute of developer app      |
| POST   | /v1/developers/_developer_/apps/_appname_/attributes/_name_ | update attribute of developer app        |
| DELETE | /v1/developers/_developer_/apps/_appname_/attributes/_name_ | deletes attribute of developer app       |

* For POST content-type: application/json is required.

## Example developer app definition

```json
{
  "name": "teleporter",
  "displayName": "Teleportrrrrr",
  "status": "active",
  "attributes": [
    {
        "name": "people_quotaPerSecond",
        "value": "20"
    },
    {
      "name": "IPAccessList",
      "value": "10.0.0.0/8,192.168.178.0/24"
    },
    {
      "name": "Referer",
      "value": "example.org,*example.com"
    }
  ]
}
```

## Fields specification

| fieldname        | optional  | purpose                                                   |
| ---------------- | --------- | --------------------------------------------------------- |
| name             | mandatory | name (cannot be updated afterwards)                       |
| displayName      | optional  | friendly name                                             |
| status           | mandatory | status, call will be rejected in case not set to "active" |
| attributes       | optional  | specific attributes                                       |

## Attribute specification

| attribute name                | purpose                                                       | possible values                |
| ----------------------------- | ------------------------------------------------------------- | ------------------------------ |
| IPAccessList                  | source ip request access list                                 | 10.0.0.0/8, 192.168.42.0/24    |
| Referer                       | HTTP Referer hostname access list                             | *.example.com, www.example.net |
| _productname_ _quotaPerSecond | Set a specific quota per second rate for a particular product | 50                             |

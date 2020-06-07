# Developer application

A developer application holds the attribut.

## Supported methods and paths

| Method | Path                                                                           | What                                     |
| ------ | ------------------------------------------------------------------------------ | ---------------------------------------- |
| GET    | /v1/organization/_org_/developers/_developer_/apps                             | retrieve all apps of developer           |
| POST   | /v1/organization/_org_/developers/_developer_/apps                             | creates a new developer app              |
| GET    | /v1/organization/_org_/developers/_developer_/apps/_appname_                   | retrieve one developer app               |
| POST   | /v1/organization/_org_/developers/_developer_/apps/_appname_                   | updates an existing developer app        |
| DELETE | /v1/organization/_org_/developers/_developer_/apps/_appname_                   | deletes a developer app                  |
| GET    | /v1/organization/_org_/developers/_developer_/apps/_appname_/attributes        | retrieve all attributes of developer app |
| POST   | /v1/organization/_org_/developers/_developer_/apps/_appname_/attributes        | update all attribute of developer app    |
| GET    | /v1/organization/_org_/developers/_developer_/apps/_appname_/attributes/_name_ | retrieve attribute of developer app      |
| POST   | /v1/organization/_org_/developers/_developer_/apps/_appname_/attributes/_name_ | update attribute of developer app        |
| DELETE | /v1/organization/_org_/developers/_developer_/apps/_appname_/attributes/_name_ | deletes attribute of developer app       |

* For POST content-type: application/json is required.

## Example developer app definition

```json
{
  "name": "teleporter",
  "displayName": "Teleportrrrrr",
  "organizationName": "petstore",
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

| fieldname        | optional  | purpose                             |
| ---------------- | --------- | ----------------------------------- |
| name             | mandatory | name (cannot be updated afterwards) |
| displayName      | optional  | friendly name                       |
| organizationName | mandatory | last name                           |
| status           | mandatory | wheter                              |
| attributes       | optional  | specific attributes                 |

## Attribute specification

| attribute name | purpose                        | possible values |
| -------------- | ------------------------------ | --------------- |
| IPAccessList   | source ip request access list  |                 |
| Referer        | HTTP Referer check access list |                 |

## Inspiration

Definition of developer app is based on work [created and shared by Google](https://docs.apigee.com/reference/apis/apigee/rest/v1/organizations.developers.apps) and used according to terms described in the [Creative Commons 4.0 Attribution License](https://creativecommons.org/licenses/by/4.0/)

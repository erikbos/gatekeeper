# developer

A developer is an API user and has defined.

## Supported methods and paths

| Method | Path                                                                | What                                 |
| ------ | ------------------------------------------------------------------- | ------------------------------------ |
| GET    | /v1/organization/_org_/developers                                   | retrieve all developers              |
| POST   | /v1/organization/_org_/developers                                   | creates a new developer              |
| GET    | /v1/organization/_org_/developers/_developername_                   | retrieve a developer                 |
| POST   | /v1/organization/_org_/developers/_developername_                   | updates an existing developer        |
| DELETE | /v1/organization/_org_/developers/_developername_                   | deletes a developer                  |
| GET    | /v1/organization/_org_/developers/_developername_/attributes        | retrieve all attributes of developer |
| POST   | /v1/organization/_org_/developers/_developername_/attributes        | update all attribute of developer    |
| GET    | /v1/organization/_org_/developers/_developername_/attributes/_name_ | retrieve one attribute of developer  |
| POST   | /v1/organization/_org_/developers/_developername_/attributes/_name_ | update an attribute of developer     |
| DELETE | /v1/organization/_org_/developers/_developername_/attributes/_name_ | deletes attribute of developer       |

* For POST content-type: application/json is required.

## Example developer definition

```json
{
    "email": "john@example.com",
    "firstName": "John",
    "lastName": "Smith",
    "userName": "john",
        "attributes": [
    {
        "name": "Shoesize",
        "value": "42"
    },
    {
        "name": "CustomerGroup",
        "value": "VIP"
    }
    ]
}
```

## Fields specification

| fieldname  | optional  | purpose             |
| ---------- | --------- | ------------------- |
| email      | mandatory | name                |
| firstName  | mandatory | first name          |
| lastName   | mandatory | last name           |
| userName   | mandatory | user name           |
| attributes | optional  | specific attributes |

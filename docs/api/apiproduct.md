# APIproduct

An apiproduct defines a set paths which are allowed to be accessed. Policies can be set to add additional operations to be before forwarding a request upstream. Attributes can be added to provide settings for policies.

## Supported methods and paths

| Method | Path                                                               | What                                  |
| ------ | ------------------------------------------------------------------ | ------------------------------------- |
| GET    | /v1/organization/_org_/apiproducts                                 | retrieve all apiproducts              |
| POST   | /v1/organization/_org_/apiproducts                                 | creates a new apiproduct              |
| GET    | /v1/organization/_org_/apiproducts/_productname_                   | retrieve an apiproduct                |
| POST   | /v1/organization/_org_/apiproducts/_productname_                   | updates an existing apiproduct        |
| DELETE | /v1/organization/_org_/apiproducts/_productname_                   | deletes an apiproduct                 |
| GET    | /v1/organization/_org_/apiproducts/_productname_/attributes        | retrieve all attributes of apiproduct |
| POST   | /v1/organization/_org_/apiproducts/_productname_/attributes        | update all attribute of apiproduct    |
| GET    | /v1/organization/_org_/apiproducts/_productname_/attributes/_name_ | retrieve one attribute of apiproduct  |
| POST   | /v1/organization/_org_/apiproducts/_productname_/attributes/_name_ | update one attribute of apiproduct    |
| DELETE | /v1/organization/_org_/apiproducts/_productname_/attributes/_name_ | deletes attribute of apiproduct       |

* For POST content-type: application/json is required.

## Example developer definition

```json
{
    "name": "VIPTicket",
    "displayName": "TicketService VIP Inc",
    "routeGroup": "routes_443",
    "paths": [
        "/ticketservice/basic/*",
        "/ticketservice/vip/*"
    ],
    "attributes": [
    {
        "name": "VIPTicket_quotaPerSecond",
        "value": "42"
    }
    ],
    "policies": "checkIPAccessList,checkReferer,qps,sendAPIKey,sendDeveloperEmail,sendDeveloperID,sendDeveloperAppID"
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
| policies   | optional  | policies to apply   |

## Attribute specification

| attribute name                | purpose                              | possible values |
| ----------------------------- | ------------------------------------ | --------------- |
| _productname_ _quotaPerSecond | Set a specific quota per second rate |                 |

## Policy specification

The policies field can contain a comma separate list of policies which will be applied before sending the request upstream to a backend.

| attribute name       | purpose                                                                  | possible values |
| -------------------- | ------------------------------------------------------------------------ | --------------- |
| checkIPAccessList    | Validate source ip address against developerapp attribute _IPAccessList_ |                 |
| checkReferer         | Validate Host header against developerapp attribute _Referer_            |                 |
| sendAPIKey           | send apikey used to upstream                                             |                 |
| sendDeveloperEmail   | send developer email to upstream                                         |                 |
| sendDeveloperID      | send developer id to upstream                                            |                 |
| sendDeveloperAppID   | send developer app id to upstream                                        |                 |
| sendDeveloperAppName | send developer app name to upstream                                      |                 |

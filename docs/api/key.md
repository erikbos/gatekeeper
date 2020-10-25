# Keys

Keys are attached to a [developer application](developerapp.md) and are used to grant access to a set of [APIProducts](apiproduct.md). Multiple can be added to a developer application.

## Supported methods and paths

| Method | Path                                                                    | What                               |
| ------ | ----------------------------------------------------------------------- | ---------------------------------- |
| GET    | /v1/developers/_developer_/apps/_appname_/keys       | retrieve all keys of developer app |
| POST   | /v1/developers/_developer_/apps/_appname_/keys       | creates new key for developer app  |
| GET    | /v1/developers/_developer_/apps/_appname_/keys/_key_ | retrieve key of developer app      |
| POST   | /v1/developers/_developer_/apps/_appname_/keys/_key_ | updates key of developer app       |
| DELETE | /v1/developers/_developer_/apps/_appname_/keys/_key_ | deletes key of developer app       |

For POST:

* content-type: application/json is required.
* consumerKey & consumerSecret can be provided to imported existing keys.

## Example key definition

```json
{
    "consumerKey": "4DrmtHuaA9ywu4rGTr2C0CFcgr1iLPbu",
    "consumerSecret": "4SOMItkaLErzH4n2",
    "apiProducts" : [
        {
            "apiproduct" : "people",
            "status" : "approved"
        }, {
            "apiproduct" : "prem1iumfish",
            "status" : "approved"
        }
    ],
    "status": "approved"
}
```

## Fields specification

| fieldname      | optional  | purpose                                                       |
| -------------- | --------- | ------------------------------------------------------------- |
| consumerKey    | mandatory | api key, used in apikey-based authentication                  |
| consumerSecret | mandatory | api key secret, used in OAuth2 authentication                 |
| apiProducts    | mandatory | allowed [APIProducts](apiproducts.md)                         |
| status         | mandatory | status, requests will not be allowed if not set to "approved" |

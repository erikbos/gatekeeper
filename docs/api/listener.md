# Listener

A listener defines tcp listening port, virtual host and TLS configuration parameters.

## Supported operations

| Method | Path                                       | What                             |
| ------ | ------------------------------------------ | -------------------------------- |
| GET    | /v1/listeners                              | Retrieve all listeners           |
| POST   | /v1/listeners                              | Creates a new listener           |
| GET    | /v1/listeners/_listener_                   | Retrieve a listener              |
| POST   | /v1/listeners/_listener_                   | Updates an existing listener     |
| DELETE | /v1/listeners/_listener_                   | Deletes a listener               |
| GET    | /v1/listeners/_listener_/attributes        | Retrieve all listener attributes |
| POST   | /v1/listeners/_listener_/attributes        | Update all listener attributes   |
| GET    | /v1/listeners/_listener_/attributes/_name_ | Retrieve one listener attribute  |
| POST   | /v1/listeners/_listener_/attributes/_name_ | Udate one listenerattribute      |
| DELETE | /v1/listeners/_listener_/attributes/_name_ | Delete one listener attribute    |

_For POST content-type: application/json is required._

## Fields specification

| fieldname        | optional  | purpose                                           |
| ---------------- | --------- | ------------------------------------------------- |
| name             | mandatory | Name (cannot be updated afterwards)               |
| displayName      | optional  | Friendly name                                     |
| virtualHosts     | mandatory | Array of virtal hostnames                         |
| port             | mandatory | Port Envoy needs to listen on                     |
| organizationName | mandatory | Organization name                                 |
| routeGroup       | mandatory | Indicate which http routing table will be applied |
| attributes       | optional  | Specific configuration to apply                   |

## Attribute specification

| attribute name       | purpose                                              | possible values              |
| -------------------- | ---------------------------------------------------- | ---------------------------- |
| HTTPProtocol         | Highest HTTP protocol to support                     | HTTP/1.1, HTTP/2, HTTP/3     |
| TLSEnable            | Whether to enable TLS or not, HTTP/2 always uses TLS | true, false                  |
| TLSCertificate       | Certificate to use for TLS                           |                              |
| TLSCertificateKey    | Key of certificate                                   |                              |
| TLSMinimumVersion    | Minimum version of TLS to use                        | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSMaximumVersion    | Maximum version of TLS to use                        | TLSv10,TLSv11, TLSv12 TLSv13 |
| TLSCipherSuites      | Allowed TLS cipher suite                             |                              |
| AccessLogFile        | File for storing access logs                         |                              |
| AccessLogCluster     | Cluster to send access logs to                       |                              |

All attributes listed above are mapped onto configuration properties of [Envoy listener API specifications](https://www.envoyproxy.io/docs/envoy/latest/api-v3/api/v3/listener.proto#listener) for detailed explanation of purpose and allowed value of each attribute.

The listener options exposed this way are a subset of Envoy's capabilities, in general any listener configuration option Envoy supports can be exposed  this way. Feel free to open an issue if you need more of Envoy's functionality exposed.

## Policy specification

The policies field can contain a comma separate list of policies which will be evaluated.

| attribute name       | purpose                                                                  |
| -------------------- | ------------------------------------------------------------------------ |
| checkAPIKey          | Verify apikey                                                            |
| checkOAuth2          | Verify OAuth2 accesstoken                                                |
| removeAPIKeyFromQP   | Remove apikey from query parameters                                      |
| lookupGeoIP          | Set country and state of connecting ip address as [Dynamic Metadata](https://www.envoyproxy.io/docs/envoy/latest/configuration/advanced/well_known_dynamic_metadata) |

## Envoycp control plane

Envoycp monitors database for changed listeners at `xds.configcompileinterval` interval. In case of changes envoycp will compile a new Envoy configuration and notify all envoyproxy instances.

## Example listener configurations

HTTP listener on port `80` mapping incoming requests for http virtual host `www.petstore.com` to routegroup `routes_80`

```json
{
    "name": "example_80",
    "displayName": "Example Inc.",
    "virtualHosts": [
         "www.petstore.com"
    ],
    "port": 80,
    "organizationName": "petstore",
    "routeGroup": "routes_80"
}
```

One listener with two virtual hosts sharing a TLS certificate:

```json
{
    "name": "example_443_1",
    "displayName": "example secure",
    "virtualHosts": [
        "www.example.com",
        "test.com"
    ],
    "port": 443,
    "routeGroup": "routes_443",
    "policies": "lookupGeoIP,checkAPIKey",
    "organizationName": "petstore",
    "attributes": [
        {
            "name": "HTTPProtocol",
            "value": "HTTP/2"
        },
        {
            "name": "TLSCertificate",
            "value": "-----BEGIN CERTIFICATE-----\nMIIDgDCCAmgCCQCN5+Z6gKrj5zANBgkqhkiG9w0BAQUFADCBgTELMAkGA1UEBhMCVVMxCzAJBgNVBAgMAk5ZMREwDwYDVQQHDAhOZXcgWW9yazEWMBQGA1UECgwNUGV0U3RvcmUgSW5jLjEYMBYGA1UEAwwPd3d3LmV4YW1wbGUuY29tMSAwHgYJKoZIhvcNAQkBFhFpbmZvQHBldHN0b3JlLmNvbTAeFw0yMDA1MjkxOTQ3NDVaFw0yMTEwMTExOTQ3NDVaMIGBMQswCQYDVQQGEwJVUzELMAkGA1UECAwCTlkxETAPBgNVBAcMCE5ldyBZb3JrMRYwFAYDVQQKDA1QZXRTdG9yZSBJbmMuMRgwFgYDVQQDDA93d3cuZXhhbXBsZS5jb20xIDAeBgkqhkiG9w0BCQEWEWluZm9AcGV0c3RvcmUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyGwduQHZsQ9ToI9iXJY+QxC6QrF9Wf5QLiCg00cyN/8cDmPuZa/apVzb9u+7z4L/T9eS1CM7MpyLqyDThlH/aYDmBcMz04goiSINDdMwlntfzGvn8MgILKDy/isaG+TmHP1hb2BzqQw+ipFE+7BARuOo+9rLxbczE4ioydRzi9ua2C10VpUy/S2D65RITbsD1FUUPZvA/Z36bQyORiSKTKXqe1nUERoXWRrOnEgyBjbtZm64Fk0+7jfst1kAr3I1G3ssbxTZa6q839r6Pbqi9qIgLcZG5sFZUvMT3JfOwIrJkKUdBiYPBfsWG9od0L2NRTtYTe/+xMMHYwWTTYAgxwIDAQABMA0GCSqGSIb3DQEBBQUAA4IBAQAqSHDgee7fy6lDi2mWZt1HkXzFZxYADm1xRgIgxq2O+Benw98FTu149uswxtDaPPlGXCuwCZmPL5GMhFvw/L5X/JWsy5yugH5/v//jSKvUEhIOkHKHqNmRgFbm7wt9mv5Ca/CKB6qgIVBAVeDYTLTQJ3t5jz3ZJ16H8ObYpMGFGrPZojzgbwbglDaoxYOXjfK1fVe0kpIHvmOkWeTgVU5eetAxOhL9x2KddTomacN/DtFaFSeD2zwKjcbmzU7ggO3eiPrSjQrX4bEq7J3bw5BboDXAL7829a7tGe/hal5kN8H4rXUt8LHEHngh2Epqx1mYBDC6qEPNj5kMPpN7EQ5s\n-----END CERTIFICATE-----"
        },
        {
            "name": "TLSCertificateKey",
            "value": "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEAyGwduQHZsQ9ToI9iXJY+QxC6QrF9Wf5QLiCg00cyN/8cDmPuZa/apVzb9u+7z4L/T9eS1CM7MpyLqyDThlH aYDmBcMz04goiSINDdMwlntfzGvn8MgILKDy/isaG+TmHP1hb2BzqQw+ipFE+7BARuOo+9rLxbczE4ioydRzi9ua2C10VpUy/S2D65RITbsD1FUUPZvA/Z36bQyORiSKTKXqe1nUERoXWRrOnEgyBjbtZm64Fk0+7jfst1kAr3I1G3ssbxTZa6q839r6Pbqi9qIgLcZG5sFZUvMT3JfOwIrJkKUdBiYPBfsWG9od0L2NRTtYTe/+xMMHYwWTTYAgxwIDAQABAoIBAQCAxQ43vtuaKknFwDonWJS6TDYQAa+TMZVcfbQ26uh2F99z03rpNJpbYpUlTBQ0GGtnZg89Y0F2nCQUmCuvgmGC7MFddHSI9VNuAEW42za9iJkdYzsLdcniuqpE6XaF84Rxnc6LW8IUG/zW1M0olK5HnaAF6SbBappTc5tWybxPX24TzGFEJ49SYXwYWtyPZnFwlChsYPD2idx7vsDcDr/k0eRBQgOL/MzJw25gm5XZ8sWtQ9rpNx7ZtlmPOlMS0S5+4gQ9QTRKCtySbrqLFSLXfP92F8PBOpTg20/EPn7c8taYQXNtRQo28P1tq32pEeQMHudHy4otJpHMGN2aaw2JAoGBAPfyQzo2XcORKLae0qIVKb5DEEXNNLr71xiy4RHYqYHN9pSHeJ60ZtocdK7mFuRwE1Op7DtkybsFS3T+UWysX+aBrI0t/RUJuNEppnQX/87DbhejvIXRPRSuWQUe20mV8x337Hb+9VdW8DG5T0f5fH8hTVP+j/BxFWby3dd17lnNAoGBAM7urd3t5UCbtaDpnT6JC4rlC0kvmi0oWQ6ZIY3cBNZKeSu6Vt77JHOazHAGYnxGlO3pB0rEHhk8T/R85UAE94L9V3be9PgrHF80+nAwP+1fWWsEuFEhr7Bl1VNEylUtHH3EUBQrp19/QrO02bfSW0hV7IAjXXEczdAvNRGHQIDjAoGANjvQhqgjpEZZEHD3A6r7YXmL0qjLEudJKkbeQigRE6p4eA6VzKkLIkQ9JZCAi2EUaSVu3aLzGSxUT/fc2Zdutp3An3TiubpRqbahiR1Cv/gxWYxgDSkyYhastBkGwDbDYde76l9kTMFgco/lDoo8uBYRswCofWBO3SDcc2eBRjUCgYEAshELRp0vGIClM9mzuRte9l+QbaLrzg4ZTImTKSp6cxhU2r8Xf/um61/6qi+kUgK+p1dOMhU/PUH8H4vWDlf30R1GRYEoVeFrIbZKB35NlGrnXEMMhKwzLd0DTAs2/UK7cLIcoq7J8VBmSpPGgfsfF8jwoXdNMkeyB4KH7RRw+jcCgYAzbVOOPQhUvDTLEL8fk8UBiR989jKtkX3vGGVO+ATApeJwFwgymx/YrPwImj1XOswAf+tbxFPDmpg40mU36J+1EzoQb6G4GhAQqw6qcRYjynqVl5gNEBIIY0p5johx3tmEf/oV+ROD1A5uNhgIk+/vfrNpvUByVJwzDUgnUH3ljQ==\n-----END RSA PRIVATE KEY-----"
        },
        {
            "name": "TLSCipherSuites",
            "value": "[ECDHE-RSA-CHACHA20-POLY1305|ECDHE-RSA-AES256-GCM-SHA384|ECDHE-RSA-AES128-GCM-SHA256]"
        },
        {
            "name": "TLSMinimumVersion",
            "value": "TLSv1_2"
        }
    ]
}
```

# Introduction

This directory contains example API calls to run.

Easiest way to use these is to:
1. Install Visual Code plugin [REST Client](https://github.com/Huachao/vscode-restclient) this will allow you to edit and execute API calls directly from Visual Code.
2. Update Visual Code's settings **settings.json** to configure a hostname & port for the REST Client:

```
"rest-client.environmentVariables": {

        "localhost": {
            "hostname": "localhost:7777",
        },
        "production": {
            "hostname": "dbadmin.example.com:7777",
        }
    }
}
```

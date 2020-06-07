# Introduction

This directory contains some example API calls.

A convenient way to execute these is using Visual Code's [REST Client](https://github.com/Huachao/vscode-restclient).
This will allow you to edit and execute API calls directly from within Visual Code.

You might want to put the hostname and port of *dbadmin* in Visual Code's settings **settings.json**:

```json
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

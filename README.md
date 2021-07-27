# vault-client

A simple golang application that creates/gets a secret tied to an organization.

## Getting Started
The easiest way to get started is with Docker using docker-compose

```shell
docker-compose up
```
This will start both the vault server and the client API.

The client API runs on PORT `9200` by default

## APIs

`GET /secrets/{organization_id}`
Get organization secrets

`PUT  /secrets/{organization_id}`
Create organization secrets
```json
{
  "AWS_ACCESS_KEY_ID": "value",
  "SECRET_KEY": "value"
}
```


## Limitations
You cannot append new values to previous secret values (This can be improved by implementing secret versions)

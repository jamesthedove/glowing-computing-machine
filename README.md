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


`POST  /initialize/{organization_id}`
This will create and mount an AWS path for the organization 


`POST  /configure/{organization_id}`
This configures the root AWS access key or the organization

Request Body
```json
{
  "aws_secret_key": "value",
  "secret_key": "value",
  "region": "us-east-1"
}
```


`POST  /generate-credentials/{organization_id}`
This generates a temporary credential for the organization
```json
{
  "access_key": "AKI***",
  "secret_key": "**",
  "security_token": null
}
```


`POST  /run-tf/{organization_id}`
This applies the terraform file in the directory
Note that this currently doesn't work with docker

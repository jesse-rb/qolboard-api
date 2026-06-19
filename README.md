# qolboard-api

## 1. Getting started with development

### 1.1 Docker & Docker Compose

This project uses docker to containerize the development environment (and one day production) for convenience. Everything required for local development is containerized, so you should not need to install anything other than docker and docker-compose binaries to get started.
+ Docker (If on macos, linux, or windows/wsl2 [Docker Desktop](https://www.docker.com/products/docker-desktop/) is a convenient way to to install docker and docker-compose binaries)

(optional) Destroy local containerized env (helpful to start with a clean slate):
```
make docker-compose-destroy
```

### 1.2 Makefil

This project uses a Makefile for some convenient commands to get up and running, and for ongoing development.

### 1.3 env

A local `.env.example` file is provided, which should have the correct defaults for local development.

Set up .env:
```
cp .env.example .env
```

### 1.4 Database

This project uses a postgres database. For convenience, the docker-compose file includes a postgers service which can be used for development.

Start db:
```
make db-up
```

Stop db:
```
make db-down
```

Run DB migrations:
```
make migrations-up
```

Check DB migrations status:
```
make migrations-status
```

(optional) For convenience, the docker-compose also includes adminer (a convenient web based DB client), but you can use any db client.
```
make adminer-up
```

```
make adminer-down
```

### 1.5 Golang API

Run the Golang API:
```
make api-run
```

## API design & architecture

### Requesting related resources

This API uses a pattern which allows the client to request additional related resources via a `with[]=resourceA.resourceB` query param. (Similar to with GraphQL, batch loading resources is optimized to avoid N+1 query problems, and a maximum depth limit is enforced)

e.g. requesting user with all their canvases, and canvases shared with them
```
GET /user?with[]=canvas_shared_accesses.canvas&with[]=canvases
```

### Responses

Responses are mostly consistent. An `errors` array is always included, and can be empty if there are no specific errors.
For endpoints that return data, if the request was successful then either a `data` object or array is included, depending on whether or not
the data is a list of resources, or a single resource.

Example error response:
```
{
    "errors": [
        {
            "message": "Unauthorized",
            "field": "",
            "value": ""
        }
    ]
}
```

Example successful response:
```
{
    "data": {
        "canvas_shared_accesses": null,
        "canvases": [],
        "email": "test@gmail.com",
        "id": "b965790f-a11a-4cbf-a9c2-1f5763099de0"
    },
    "errors": []
}
```

### Auth

Outside of the websocket connection, this is a stateless API which uses JWT authentication, with refresh tokens which allow active users to remain authenticated without being interrupted too often.

User registration and login is passwordless, users must verify their email via an email link, and users login via email one time passwords (OTP).

### Email

When running the API locally, any email that would ordinarily be sent in production is instead simply logged to stdout.

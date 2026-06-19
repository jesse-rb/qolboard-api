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
make local-api-run
```

## API design & architecture

...docs in progress

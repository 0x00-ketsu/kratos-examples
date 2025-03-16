# Usermanage

A simple usermanage application sample developed by using `go-kratos`.

## Features

Use `JWT` to authenticate the user.

- Auth
    - [x] Login
    - [x] Logout
    - [x] Change Password
    - [x] Get User Info
- Usermanage (admin oriented)
    - [x] List Users
    - [x] Get User
    - [x] Create User
    - [x] Update User Partially
    - [x] Update User Replace
    - [x] Delete User
    - [x] Reset Password
- Health
    - [x] Health Check
    - [x] Health Probe

## System Initialization

- Automatically create database if it does not exist
- Migrate database tables
- Create admin account if it does not exist

    ```json
    {"msg": "admin account created successfully", "credential": "*9Ja1CwDQNxiU5NZ"}
    ```

## Rrequirements

- `go` 1.24
- `buf`
- `wire`
- `kratos` v2.x
- `MySQL` or `PostgreSQL`
- `Redis`
- `Opentelemetry` (optional)
    - `jaeger`
- `docker` and `docker-compose` (for demo)

## Initialize

```bash
make init
```

## Build

### Build the application

```bash
make build
```

### Build the image

```bash
MODULE_PREFIX=usermanage make build-image
```

## Opentelemetry

### Jaeger

Start the `jaeger` with docker.

```bash
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:1.67.0
```

Visit the `jaeger` dashboard at `http://localhost:16686`.

## Demo

> **NOTE**
>
> `docker` and `docker-compose` are required.

- HTTP server listens on `8000`
- GRPC server listens on `9000`

```bash
# Build image
MODULE_PREFIX=usermanage make build-image

# Run the application
docker compose -f examples/docker-compose.yml up
```

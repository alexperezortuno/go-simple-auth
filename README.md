# SIMPLE AUTH

## Packages

```text
go get github.com/gin-gonic/gin
go get github.com/golang-jwt/jwt/v5
go get gorm.io/gorm
go get gorm.io/driver/sqlite
go get golang.org/x/crypto/bcrypt
```

## Run

```shell
go run main.go
```

## Build

```shell
GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o $(pwd)/dist/go_auth_$(go env GOOS)_$(go env GOARCH)
```

```shell
cd dist

export export JWT_TOKEN=li7lasdasiy3eliufdifsydfsdyfiuskh3kjlk2df89s; export PORT=8081 && ./go_auth_$(go env GOOS)_$(go env GOARCH)
```

## Endpoints

### Create token

```shell
curl -X POST http://localhost:8080/login -d '{"username": "admin", "password": "admin"}' -H "Content-Type: application/json" | jq .token
```

### Renew token

```shell
curl -X POST http://localhost:8080/api/renew -H "Content-Type: application/json" -H "Authorization: <TOKEN>" | jq .
```

### Validate token
```shell
curl -X POST http://localhost:8080/api/validate -H "Content-Type: application/json" -H "Authorization: <TOKEN>" | jq .
```

### Health check

```shell
curl -X GET http://localhost:8080/health | jq .
```

---

## Docker

Create image
```shell
docker build -t go_auth:dev .
```

Run container
```shell
docker run --rm --name go_auth -d -p 8080:8080 -e JWT_TOKEN=mysecret -e PORT=8080  go-simple-auth:dev
```

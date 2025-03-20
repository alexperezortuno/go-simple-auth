FROM golang:1.24-alpine AS builder

RUN apk add --no-cache build-base

COPY . /go/src/github.com/alexperezortuno/go-simple-auth

WORKDIR /go/src/github.com/alexperezortuno/go-simple-auth

RUN go mod tidy
RUN go env
RUN go version
RUN CGO_ENABLED=1 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o /go/bin/go-simple-auth

FROM nginx:1.27.4-alpine-slim

RUN apk add --no-cache sqlite-libs sqlite-dev sqlite

COPY --from=builder /go/bin/go-simple-auth /usr/local/bin/go-simple-auth

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

COPY nginx.conf /etc/nginx/nginx.conf

RUN chmod +x /usr/local/bin/go-simple-auth

CMD ["/usr/local/bin/entrypoint.sh"]

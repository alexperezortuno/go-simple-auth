FROM golang:1.24-alpine AS builder

RUN apk add --no-cache build-base git

RUN git clone https://github.com/alexperezortuno/go-simple-auth.git /go/src/github.com/alexperezortuno/go-simple-auth --depth 1

WORKDIR /go/src/github.com/alexperezortuno/go-simple-auth

RUN go mod tidy
RUN go env
RUN go version
RUN CGO_ENABLED=1 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build -o /go/bin/go-simple-auth

FROM alpine:3.21.3

RUN apk add --no-cache sqlite bash

COPY --from=builder /go/bin/go-simple-auth /usr/local/bin/go-simple-auth

COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

CMD ["/usr/local/bin/entrypoint.sh"]

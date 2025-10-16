FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o oauth-server ./cmd/server/

FROM alpine

WORKDIR /server
COPY --from=builder /build/oauth-server ./oauth-server

EXPOSE 8080
EXPOSE 9090

CMD ["/server/oauth-server"]

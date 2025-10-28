FROM golang:1.25 AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN GO111MODULE=on go mod download
RUN go mod vendor

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o hydros .

RUN make install-goose

FROM alpine

WORKDIR /server
COPY Makefile ./
RUN apk add --no-cache make curl tar bash

COPY --from=builder /build/hydros ./hydros
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations

RUN echo goose -h

EXPOSE 8080
EXPOSE 9090

CMD ["/server/hydros"]

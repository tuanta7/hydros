ENV_FILE=.env
MIGRATIONS_FOLDER=migrations
PROTO_FOLDER=protobuf/proto
BUF_VERSION?=1.58.0

build:
	CGO_ENABLED=0 GOOS=linux go build -o hydros .

create-client:
	./hydros create-client

env-example:
	awk -F'=' 'BEGIN {OFS="="} \
    	/^[[:space:]]*#/ {print; next} \
    	/^[[:space:]]*$$/ {print ""; next} \
    	NF>=1 {gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); print $$1"="}' .env > .env.example
	echo ".env.example generated successfully."

install-goose:
	go install github.com/pressly/goose/v3/cmd/goose@latest
	ls "$(shell go env GOPATH)/bin/" | grep goose

migrate-sql:
	goose -dir=$(MIGRATIONS_FOLDER)/postgres create $(NAME) sql

migrate-go:
	goose -dir=$(MIGRATIONS_FOLDER)/go create $(NAME) go

migrate-up:
	goose -env $(ENV_FILE) up

migrate-down:
	goose -env $(ENV_FILE) down

install-buf:
	go install github.com/bufbuild/buf/cmd/buf@v${BUF_VERSION}

buf-dev:
	buf dep update
	buf export buf.build/bufbuild/protovalidate --output=.

buf-gen:
	buf dep update
	buf generate

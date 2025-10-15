MIGRATIONS_FOLDER=migrations
ENV_FILE=.env

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

env-example:
	awk -F'=' 'BEGIN {OFS="="} \
    	/^[[:space:]]*#/ {print; next} \
    	/^[[:space:]]*$$/ {print ""; next} \
    	NF>=1 {gsub(/^[[:space:]]+|[[:space:]]+$$/, "", $$1); print $$1"="}' .env > .env.example
	echo ".env.example generated successfully."
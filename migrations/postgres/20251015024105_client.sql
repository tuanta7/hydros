-- +goose Up
CREATE TABLE IF NOT EXISTS client
(
    id                              VARCHAR(255),
    name                            VARCHAR(255)            NOT NULL,
    description                     TEXT                    NOT NULL,
    secret                          TEXT                    NOT NULL,
    scope                           TEXT                    NOT NULL,
    redirect_uris                   TEXT                    NOT NULL,
    grant_types                     TEXT                    NOT NULL,
    response_types                  TEXT                    NOT NULL,
    audience                        TEXT                    NOT NULL,
    request_uris                    TEXT                    NOT NULL,
    jwks                            TEXT                    NOT NULL,
    jwks_uri                        TEXT                    NOT NULL,
    token_endpoint_auth_method      VARCHAR(25)             NOT NULL,
    token_endpoint_auth_signing_alg VARCHAR(10)             NOT NULL,
    created_at                      TIMESTAMP DEFAULT now() NOT NULL,
    updated_at                      TIMESTAMP DEFAULT now() NOT NULL,
    PRIMARY KEY (id)
);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE client;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- +goose Up
CREATE TABLE IF NOT EXISTS access_token
(
    signature    VARCHAR(255) NOT NULL PRIMARY KEY,
    request_id   VARCHAR(40)  NOT NULL,
    requested_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    client_id    VARCHAR(255) NOT NULL REFERENCES client (id) ON DELETE CASCADE,
    subject      VARCHAR(255) NOT NULL DEFAULT '',
    active       BOOLEAN      NOT NULL DEFAULT true,
    UNIQUE (request_id)
);

CREATE TABLE IF NOT EXISTS refresh_token
(
    signature              VARCHAR(255) NOT NULL PRIMARY KEY,
    access_token_signature VARCHAR(255)          DEFAULT NULL,
    request_id             VARCHAR(40)  NOT NULL,
    requested_at           TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    client_id              VARCHAR(255) NOT NULL REFERENCES client (id) ON DELETE CASCADE,
    subject                VARCHAR(255) NOT NULL DEFAULT '',
    active                 BOOLEAN      NOT NULL DEFAULT true,
    UNIQUE (request_id)
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE access_token;
DROP TABLE refresh_token;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

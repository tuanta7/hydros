-- +goose Up
CREATE TABLE IF NOT EXISTS jwk
(
    kid        VARCHAR(255) PRIMARY KEY,
    set        VARCHAR(255)            NOT NULL,
    key        TEXT                    NOT NULL,
    algorithm  TEXT                    NOT NULL,
    use        VARCHAR(10)             NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS jwk;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- +goose Up
CREATE TABLE IF NOT EXISTS jwk
(
    kid        VARCHAR(255),
    sid        VARCHAR(255)            NOT NULL,
    key        TEXT                    NOT NULL,
    active     BOOLEAN   DEFAULT FALSE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    PRIMARY KEY (kid)
);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS jwk;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

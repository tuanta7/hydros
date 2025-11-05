-- +goose Up
CREATE TABLE IF NOT EXISTS login_session
(
    id                           VARCHAR(40)           NOT NULL,
    authenticated_at             TIMESTAMP,
    subject                      VARCHAR(255)          NOT NULL,
    remember                     BOOLEAN DEFAULT false NOT NULL,
    identity_provider_session_id VARCHAR(40),
    PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS login_session_sub_idx ON login_session (subject);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS login_session_sub_idx;
DROP TABLE IF EXISTS login_session;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

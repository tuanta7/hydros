-- +goose Up
CREATE TABLE IF NOT EXISTS code
(
    signature        VARCHAR(255)                               NOT NULL,
    request_id       VARCHAR(40)                                NOT NULL,
    requested_at     TIMESTAMP    DEFAULT now()                 NOT NULL,
    client_id        VARCHAR(255)                               NOT NULL,
    scope            TEXT                                       NOT NULL,
    granted_scope    TEXT                                       NOT NULL,
    audience         TEXT         DEFAULT ''::TEXT,
    granted_audience TEXT         DEFAULT ''::TEXT,
    form_data        TEXT                                       NOT NULL,
    session_data     TEXT                                       NOT NULL,
    subject          VARCHAR(255) DEFAULT ''::CHARACTER VARYING NOT NULL,
    active           BOOLEAN      DEFAULT TRUE                  NOT NULL,
    challenge        VARCHAR(40), -- foreign key to a flow login challenge
    PRIMARY KEY (signature)
);

CREATE INDEX IF NOT EXISTS code_requested_at_idx ON code (requested_at);
CREATE INDEX IF NOT EXISTS code_client_id_idx ON code (client_id);
CREATE INDEX IF NOT EXISTS code_client_id_subject_idx ON code (client_id, subject);
CREATE INDEX IF NOT EXISTS code_request_id_idx ON code (request_id);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS code_requested_at_idx;
DROP INDEX IF EXISTS code_client_id_idx;
DROP INDEX IF EXISTS code_client_id_subject_idx;
DROP INDEX IF EXISTS code_request_id_idx;
DROP TABLE IF EXISTS code;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

-- +goose Up
CREATE TABLE IF NOT EXISTS flow
(
    id                            VARCHAR(40)                                NOT NULL,
    acr                           TEXT         DEFAULT ''::TEXT              NOT NULL,
    amr                           JSONB        DEFAULT '[]'::JSONB,
    login_skip                    BOOLEAN                                    NOT NULL,
    login_csrf                    VARCHAR(40)                                NOT NULL,
    login_remember                BOOLEAN      DEFAULT false                 NOT NULL,
    login_remember_for            INTEGER                                    NOT NULL,
    login_extend_session_lifetime BOOLEAN      DEFAULT false                 NOT NULL,
    login_was_handled             BOOLEAN      DEFAULT false                 NOT NULL,
    login_error                   TEXT,
    login_authenticated_at        TIMESTAMP,
    login_session_id              VARCHAR(40),
    subject                       VARCHAR(255)                               NOT NULL,
    forced_subject_identifier     VARCHAR(255) DEFAULT ''::CHARACTER VARYING NOT NULL,
    identity_provider_session_id  VARCHAR(40),

    consent_skip                  BOOLEAN      DEFAULT false                 NOT NULL,
    consent_csrf                  VARCHAR(40),
    consent_remember              BOOLEAN      DEFAULT false                 NOT NULL,
    consent_remember_for          INTEGER,
    consent_was_handled           BOOLEAN      DEFAULT false                 NOT NULL,
    consent_error                 TEXT,
    consent_handled_at            TIMESTAMP,

    requested_at                  TIMESTAMP    DEFAULT now()                 NOT NULL,
    request_url                   TEXT                                       NOT NULL,
    requested_scope               JSONB                                      NOT NULL,
    granted_scope                 JSONB,
    requested_audience            JSONB        DEFAULT '[]'::JSONB,
    granted_audience              JSONB        DEFAULT '[]'::JSONB,
    client_id                     VARCHAR(255)                               NOT NULL,

    context                       JSONB        DEFAULT '{}'::JSONB           NOT NULL,
    oidc_context                  JSONB        DEFAULT '{}'::JSONB           NOT NULL,
    state                         INTEGER                                    NOT NULL,

    PRIMARY KEY (id),
    CONSTRAINT flow_client_id_fk FOREIGN KEY (client_id) REFERENCES client ON DELETE CASCADE,
    CONSTRAINT flow_login_session_id_fk FOREIGN KEY (login_session_id) REFERENCES login_session ON DELETE SET NULL,
    CONSTRAINT flow_check CHECK (
        (state = 128) OR (state = 129) OR (state = 1)
            OR (
            (state = 2) AND (
                (login_remember IS NOT NULL) AND
                (login_remember_for IS NOT NULL) AND
                (login_error IS NOT NULL) AND
                (acr IS NOT NULL) AND
                (login_was_handled IS NOT NULL) AND
                (conTEXT IS NOT NULL) AND
                (amr IS NOT NULL)
                ))
            OR (
            (state = 3) AND (
                (login_remember IS NOT NULL) AND
                (login_remember_for IS NOT NULL) AND
                (login_error IS NOT NULL) AND
                (acr IS NOT NULL) AND
                (login_was_handled IS NOT NULL) AND
                (context IS NOT NULL) AND
                (amr IS NOT NULL)
                ))
            OR (
            (state = 4) AND (
                (login_remember IS NOT NULL) AND
                (login_remember_for IS NOT NULL) AND
                (login_error IS NOT NULL) AND
                (acr IS NOT NULL) AND
                (login_was_handled IS NOT NULL) AND
                (context IS NOT NULL) AND
                (amr IS NOT NULL) AND
                (consent_skip IS NOT NULL) AND
                (consent_csrf IS NOT NULL)
                ))
            OR (
            (state = 5) AND (
                (login_remember IS NOT NULL) AND
                (login_remember_for IS NOT NULL) AND
                (login_error IS NOT NULL) AND
                (acr IS NOT NULL) AND
                (login_was_handled IS NOT NULL) AND
                (context IS NOT NULL) AND
                (amr IS NOT NULL) AND
                (consent_skip IS NOT NULL) AND
                (consent_csrf IS NOT NULL)
                ))
            OR (
            (state = 6) AND (
                (login_remember IS NOT NULL) AND
                (login_remember_for IS NOT NULL) AND
                (login_error IS NOT NULL) AND
                (acr IS NOT NULL) AND
                (login_was_handled IS NOT NULL) AND
                (context IS NOT NULL) AND
                (amr IS NOT NULL) AND
                (consent_skip IS NOT NULL) AND
                (consent_csrf IS NOT NULL) AND
                (granted_scope IS NOT NULL) AND
                (consent_remember IS NOT NULL) AND
                (consent_remember_for IS NOT NULL) AND
                (consent_error IS NOT NULL) AND
                (consent_was_handled IS NOT NULL)
                ))
        )
);

CREATE INDEX IF NOT EXISTS flow_client_id_subject_idx ON flow (client_id, subject);
CREATE INDEX IF NOT EXISTS flow_client_id_idx ON flow (client_id);
CREATE INDEX IF NOT EXISTS flow_login_session_id_idx ON flow (login_session_id);
CREATE INDEX IF NOT EXISTS flow_sub_idx ON flow (subject);
CREATE INDEX IF NOT EXISTS flow_previous_consents_idx ON flow (subject, client_id, consent_skip, consent_error, consent_remember);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP INDEX IF EXISTS flow_previous_consents_idx;
DROP INDEX IF EXISTS flow_sub_idx;
DROP INDEX IF EXISTS flow_login_session_id_idx;
DROP INDEX IF EXISTS flow_client_id_idx;
DROP INDEX IF EXISTS flow_client_id_subject_idx;
DROP TABLE IF EXISTS flow;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

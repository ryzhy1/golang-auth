-- +goose Up
-- +goose StatementBegin
CREATE TABLE tokens
(
    id                 SERIAL PRIMARY KEY,
    user_id            UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    access_token       VARCHAR(255) NOT NULL,
    refresh_token      UUID         NOT NULL,
    access_expired_at  TIMESTAMP    NOT NULL,
    refresh_expired_at TIMESTAMP    NOT NULL,
    is_active          BOOLEAN      NOT NULL DEFAULT TRUE,
    is_banned          BOOLEAN      NOT NULL DEFAULT FALSE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tokens;
-- +goose StatementEnd

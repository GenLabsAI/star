-- +goose Up
-- +goose StatementBegin
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
    content,
    session_id UNINDEXED,
    message_id UNINDEXED,
    role UNINDEXED,
    content_rowid=rowid,
    tokenize = 'porter unicode61'
);
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS messages_fts;
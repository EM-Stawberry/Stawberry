-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_offers_created_at ON offers(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_offers_created_at;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE businesses
    ADD COLUMN link text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE businesses
    DROP COLUMN IF EXISTS link;
-- +goose StatementEnd
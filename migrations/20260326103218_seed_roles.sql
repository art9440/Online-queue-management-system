-- +goose Up
-- +goose StatementBegin
INSERT INTO roles (name)
VALUES
    ('super_admin'),
    ('business_admin'),
    ('manager'),
    ('employee')
ON CONFLICT (name) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM roles
WHERE name IN ('super_admin', 'business_admin', 'manager', 'employee');
-- +goose StatementEnd
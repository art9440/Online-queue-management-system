-- +goose Up
-- +goose StatementBegin

INSERT INTO users (login, password_hash, role_id, business_id)
SELECT 
  'admin@test.com',
  '$2a$10$qT0gybEHAxeWguDHtQZmc.WxeaDC7r0lvqgIKa4NE.61IOp.TgvH.',
  r.id,
  1
FROM roles r
WHERE r.name = 'business_admin'
AND NOT EXISTS (
  SELECT 1 FROM users WHERE login = 'admin@test.com'
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DELETE FROM users
WHERE login = 'admin@test.com';

-- +goose StatementEnd
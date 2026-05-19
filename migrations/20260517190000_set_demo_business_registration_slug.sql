-- +goose Up
-- +goose StatementBegin
UPDATE businesses
SET registration_slug = 'demo-business'
WHERE name = 'Demo Business'
  AND type = 'service_company'
  AND registration_slug IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE businesses
SET registration_slug = NULL
WHERE name = 'Demo Business'
  AND type = 'service_company'
  AND registration_slug = 'demo-business';
-- +goose StatementEnd

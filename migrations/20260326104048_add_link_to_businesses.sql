-- +goose Up
ALTER TABLE businesses
ADD COLUMN registration_slug TEXT;

-- ссылка должна быть уникальной (у каждого бизнеса своя)
CREATE UNIQUE INDEX businesses_registration_slug_uindex
ON businesses (registration_slug)
WHERE registration_slug IS NOT NULL;

-- +goose Down
DROP INDEX businesses_registration_slug_uindex;

ALTER TABLE businesses
DROP COLUMN registration_slug;
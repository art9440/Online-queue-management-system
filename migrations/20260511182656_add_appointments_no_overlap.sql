-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE appointments
ADD CONSTRAINT appointments_no_employee_overlap
EXCLUDE USING gist (
    employee_id WITH =,
    tstzrange(start_time, end_time, '[)') WITH &&
)
WHERE (status IN ('pending', 'confirmed'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE appointments
DROP CONSTRAINT IF EXISTS appointments_no_employee_overlap;

-- +goose StatementEnd

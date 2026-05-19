-- +goose Up
-- +goose StatementBegin
DELETE FROM employee_schedules es
USING employees e
JOIN branches br ON br.id = e.branch_id
JOIN businesses b ON b.id = br.business_id
WHERE es.employee_id = e.id
  AND b.name = 'Demo Business'
  AND b.type = 'service_company';

INSERT INTO employee_schedules (employee_id, starts_at, ends_at)
SELECT e.id, slot.starts_at, slot.ends_at
FROM employees e
JOIN branches br ON br.id = e.branch_id
JOIN businesses b ON b.id = br.business_id
JOIN LATERAL (
    VALUES
        (
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '1 day 09 hours') AT TIME ZONE br.timezone,
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '1 day 13 hours') AT TIME ZONE br.timezone
        ),
        (
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '1 day 14 hours') AT TIME ZONE br.timezone,
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '1 day 18 hours') AT TIME ZONE br.timezone
        ),
        (
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '2 day 10 hours') AT TIME ZONE br.timezone,
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '2 day 18 hours') AT TIME ZONE br.timezone
        ),
        (
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '3 day 09 hours') AT TIME ZONE br.timezone,
            (date_trunc('day', now() AT TIME ZONE br.timezone) + interval '3 day 17 hours') AT TIME ZONE br.timezone
        )
) AS slot(starts_at, ends_at) ON true
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM employee_schedules es
USING employees e
JOIN branches br ON br.id = e.branch_id
JOIN businesses b ON b.id = br.business_id
WHERE es.employee_id = e.id
  AND b.name = 'Demo Business'
  AND b.type = 'service_company';
-- +goose StatementEnd

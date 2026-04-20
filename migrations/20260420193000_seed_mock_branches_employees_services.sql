-- +goose Up
-- +goose StatementBegin
INSERT INTO businesses (name, type)
SELECT 'Demo Business', 'service_company'
WHERE NOT EXISTS (
    SELECT 1
    FROM businesses
    WHERE name = 'Demo Business'
      AND type = 'service_company'
);

INSERT INTO branches (business_id, name, address, timezone)
SELECT b.id, v.name, v.address, v.timezone
FROM businesses b
CROSS JOIN (
    VALUES
        ('Central Branch', '10 Lenina street, Novosibirsk', 'Asia/Novosibirsk'),
        ('Left Bank Branch', '25 Karl Marx avenue, Novosibirsk', 'Asia/Novosibirsk'),
        ('Academpark Branch', '8 Nikolaeva street, Novosibirsk', 'Asia/Novosibirsk')
) AS v(name, address, timezone)
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company'
  AND NOT EXISTS (
      SELECT 1
      FROM branches br
      WHERE br.business_id = b.id
        AND br.name = v.name
  );

INSERT INTO employees (branch_id, name, surname, position)
SELECT br.id, v.name, v.surname, v.position
FROM branches br
JOIN businesses b ON b.id = br.business_id
JOIN (
    VALUES
        ('Central Branch', 'Anna', 'Petrova', 'Senior specialist'),
        ('Central Branch', 'Dmitry', 'Sokolov', 'Specialist'),
        ('Left Bank Branch', 'Elena', 'Smirnova', 'Senior specialist'),
        ('Left Bank Branch', 'Ivan', 'Kozlov', 'Specialist'),
        ('Academpark Branch', 'Maria', 'Volkova', 'Senior specialist'),
        ('Academpark Branch', 'Pavel', 'Orlov', 'Specialist')
) AS v(branch_name, name, surname, position) ON v.branch_name = br.name
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company'
  AND NOT EXISTS (
      SELECT 1
      FROM employees e
      WHERE e.branch_id = br.id
        AND e.name = v.name
        AND e.surname = v.surname
  );

INSERT INTO services (branch_id, name, duration_minutes, price)
SELECT br.id, v.name, v.duration_minutes, v.price
FROM branches br
JOIN businesses b ON b.id = br.business_id
JOIN (
    VALUES
        ('Central Branch', 'Consultation', 30, 1500.00::numeric),
        ('Central Branch', 'Diagnostics', 45, 2200.00::numeric),
        ('Central Branch', 'Extended Session', 60, 3200.00::numeric),
        ('Left Bank Branch', 'Consultation', 30, 1400.00::numeric),
        ('Left Bank Branch', 'Diagnostics', 45, 2100.00::numeric),
        ('Left Bank Branch', 'Express Session', 30, 1800.00::numeric),
        ('Academpark Branch', 'Consultation', 30, 1600.00::numeric),
        ('Academpark Branch', 'Diagnostics', 45, 2300.00::numeric),
        ('Academpark Branch', 'Extended Session', 60, 3400.00::numeric)
) AS v(branch_name, name, duration_minutes, price) ON v.branch_name = br.name
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company'
  AND NOT EXISTS (
      SELECT 1
      FROM services s
      WHERE s.branch_id = br.id
        AND s.name = v.name
  );

INSERT INTO employee_services (employee_id, service_id)
SELECT e.id, s.id
FROM employees e
JOIN branches br ON br.id = e.branch_id
JOIN businesses b ON b.id = br.business_id
JOIN services s ON s.branch_id = br.id
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company'
  AND (
      (e.name = 'Anna' AND s.name IN ('Consultation', 'Diagnostics', 'Extended Session')) OR
      (e.name = 'Dmitry' AND s.name IN ('Consultation', 'Diagnostics')) OR
      (e.name = 'Elena' AND s.name IN ('Consultation', 'Diagnostics', 'Express Session')) OR
      (e.name = 'Ivan' AND s.name IN ('Consultation', 'Express Session')) OR
      (e.name = 'Maria' AND s.name IN ('Consultation', 'Diagnostics', 'Extended Session')) OR
      (e.name = 'Pavel' AND s.name IN ('Consultation', 'Diagnostics'))
  )
  AND NOT EXISTS (
      SELECT 1
      FROM employee_services es
      WHERE es.employee_id = e.id
        AND es.service_id = s.id
  );

INSERT INTO employee_schedules (employee_id, starts_at, ends_at)
SELECT e.id, slot.starts_at, slot.ends_at
FROM employees e
JOIN branches br ON br.id = e.branch_id
JOIN businesses b ON b.id = br.business_id
JOIN LATERAL (
    VALUES
        (date_trunc('day', now()) + interval '1 day 09 hours', date_trunc('day', now()) + interval '1 day 13 hours'),
        (date_trunc('day', now()) + interval '1 day 14 hours', date_trunc('day', now()) + interval '1 day 18 hours'),
        (date_trunc('day', now()) + interval '2 day 10 hours', date_trunc('day', now()) + interval '2 day 18 hours'),
        (date_trunc('day', now()) + interval '3 day 09 hours', date_trunc('day', now()) + interval '3 day 17 hours')
) AS slot(starts_at, ends_at) ON true
WHERE b.name = 'Demo Business'
  AND b.type = 'service_company'
  AND NOT EXISTS (
      SELECT 1
      FROM employee_schedules es
      WHERE es.employee_id = e.id
        AND es.starts_at = slot.starts_at
        AND es.ends_at = slot.ends_at
  );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM businesses
WHERE name = 'Demo Business'
  AND type = 'service_company';
-- +goose StatementEnd

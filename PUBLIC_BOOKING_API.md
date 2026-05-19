# Public Booking REST API

Документ описывает публичный сценарий записи клиента по `registration_slug` бизнеса.

`registration_slug` хранится в таблице `businesses.registration_slug` и уникален среди непустых значений:

```sql
CREATE UNIQUE INDEX businesses_registration_slug_uindex
ON businesses (registration_slug)
WHERE registration_slug IS NOT NULL;
```

Публичный flow:

1. Клиент открывает ссылку бизнеса: `/public/{registrationSlug}` на frontend.
2. Frontend получает услуги бизнеса.
3. Клиент выбирает услугу.
4. Frontend получает филиалы, где эта услуга доступна.
5. Клиент выбирает филиал.
6. Frontend получает мастеров, которые делают услугу в выбранном филиале.
7. Клиент выбирает мастера.
8. Frontend получает свободные слоты мастера на дату.
9. Клиент выбирает время и отправляет данные для записи.

## Base URLs

В `docker-compose.yml` сервисы сейчас опубликованы так:

```text
branches service: http://localhost:8083
booking service:  http://localhost:8084
```

Public endpoints не требуют JWT.

## 1. Получить услуги бизнеса

```http
GET /public/{registrationSlug}/services
Host: localhost:8083
```

Пример:

```bash
curl "http://localhost:8083/public/beautiful-salon/services"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "branch_id": 1,
    "name": "Haircut",
    "duration_minutes": 30,
    "price": 1500
  }
]
```

Ошибки:

```text
400 invalid registration slug
500 internal/server repository error
```

## 2. Получить филиалы для услуги

```http
GET /public/{registrationSlug}/services/{serviceId}/branches
Host: localhost:8083
```

Пример:

```bash
curl "http://localhost:8083/public/beautiful-salon/services/1/branches"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "business_id": 100,
    "name": "Main branch",
    "address": "Lenina 1"
  }
]
```

Ошибки:

```text
400 invalid registration slug
400 invalid service id
500 internal/server repository error
```

## 3. Получить мастеров для услуги и филиала

```http
GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees
Host: localhost:8083
```

Пример:

```bash
curl "http://localhost:8083/public/beautiful-salon/services/1/branches/1/employees"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "branch_id": 1,
    "name": "Anna",
    "surname": "Ivanova",
    "position": "Master"
  }
]
```

Ошибки:

```text
400 invalid registration slug
400 invalid service id
400 invalid branch id
500 internal/server repository error
```

## 4. Получить свободные слоты мастера

Этот endpoint добавлен для выбора даты и времени перед созданием записи.

```http
GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees/{employeeId}/slots?date=YYYY-MM-DD
Host: localhost:8084
```

Пример:

```bash
curl "http://localhost:8084/public/beautiful-salon/services/1/branches/1/employees/1/slots?date=2026-05-18"
```

Ответ `200 OK`:

```json
[
  {
    "start_time": "2026-05-18T09:00:00+07:00",
    "end_time": "2026-05-18T09:30:00+07:00"
  },
  {
    "start_time": "2026-05-18T09:15:00+07:00",
    "end_time": "2026-05-18T09:45:00+07:00"
  }
]
```

Как считаются слоты:

- проверяется, что `registrationSlug`, `serviceId`, `branchId` и `employeeId` связаны с одним бизнесом;
- длительность берется из `services.duration_minutes`;
- рабочее время берется из `employee_schedules.starts_at` / `employee_schedules.ends_at`;
- занятые интервалы из `appointments` со статусами `pending` и `confirmed` исключаются;
- прошлые слоты не возвращаются;
- шаг генерации сейчас `15 minutes`.

Ошибки:

```text
400 invalid date format, expected YYYY-MM-DD
400 invalid registration slug
400 invalid service id
400 invalid branch id
400 invalid employee id
500 internal/server repository error
```

## 5. Создать запись клиента

Для публичной записи по ссылке используйте slug-route:

```http
POST /public/{registrationSlug}/appointments
Host: localhost:8084
Content-Type: application/json
```

Пример:

```bash
curl -X POST "http://localhost:8084/public/beautiful-salon/appointments" \
  -H "Content-Type: application/json" \
  -d '{
    "client": {
      "name": "Ivan",
      "surname": "Petrov",
      "phone": "+79990000000",
      "email": "ivan@example.com",
      "tg_username": "ivan_petrov"
    },
    "branch_id": 1,
    "employee_id": 1,
    "service_id": 1,
    "start_time": "2026-05-18T09:00:00+07:00",
    "comment": "First visit"
  }'
```

Обязательные поля:

```text
client.name
client.surname
client.phone
branch_id
employee_id
service_id
start_time
```

`start_time` должен быть в RFC3339.

Ответ `201 Created`:

```json
{
  "appointment_id": 10,
  "client_id": 25,
  "branch": {
    "id": 1,
    "name": "Main branch"
  },
  "employee": {
    "id": 1,
    "name": "Anna",
    "surname": "Ivanova"
  },
  "service": {
    "id": 1,
    "name": "Haircut"
  },
  "start_time": "2026-05-18T09:00:00+07:00",
  "end_time": "2026-05-18T09:30:00+07:00",
  "status": "pending",
  "comment": "First visit"
}
```

Ошибки:

```text
400 invalid request body
400 invalid start_time format, expected RFC3339
400 invalid registration slug
400 invalid branch id
400 invalid employee id
400 invalid service id
400 invalid client
400 invalid client contact
409 time slot is already busy
409 appointment is not available for selected employee, service, branch or time
500 internal/server repository error
```

`409 appointment is not available...` означает, что выбранная комбинация slug/филиал/услуга/мастер/время не прошла проверку. Например, услуга не принадлежит филиалу, мастер не работает в филиале, мастер не оказывает услугу, slug не принадлежит бизнесу филиала или время вне расписания.

## Совместимость

Старый endpoint создания записи остается доступен:

```http
POST /appointments
Host: localhost:8084
```

Для публичного frontend-сценария лучше использовать:

```http
POST /public/{registrationSlug}/appointments
```

Так backend дополнительно проверяет связь записи с бизнесом из public-ссылки.

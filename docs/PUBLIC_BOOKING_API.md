# API публичной записи

Документ описывает публичный сценарий записи клиента по `registration_slug` бизнеса.

`registration_slug` хранится в поле `businesses.registration_slug` и уникален среди непустых значений:

```sql
CREATE UNIQUE INDEX businesses_registration_slug_uindex
ON businesses (registration_slug)
WHERE registration_slug IS NOT NULL;
```

Сценарий записи:

1. Клиент открывает ссылку бизнеса на фронтенде: `/public/{registrationSlug}`.
2. Фронтенд получает услуги бизнеса.
3. Клиент выбирает услугу.
4. Фронтенд получает филиалы, где эта услуга доступна.
5. Клиент выбирает филиал.
6. Фронтенд получает мастеров, которые оказывают услугу в выбранном филиале.
7. Клиент выбирает мастера.
8. Фронтенд получает свободные слоты мастера на дату.
9. Клиент выбирает время и отправляет данные для записи.

## Базовые адреса

В `docker-compose.yml` сервисы опубликованы так:

```text
branches: http://localhost:8083
booking:  http://localhost:8084
```

Публичные endpoints не требуют JWT.

## 1. Получить услуги бизнеса

```http
GET /public/{registrationSlug}/services
Host: localhost:8083
```

Пример:

```bash
curl "http://localhost:8083/public/demo-business/services"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
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
curl "http://localhost:8083/public/demo-business/services/1/branches"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
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
curl "http://localhost:8083/public/demo-business/services/1/branches/1/employees"
```

Ответ `200 OK`:

```json
[
  {
    "id": 1,
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

```http
GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees/{employeeId}/slots?date=YYYY-MM-DD
Host: localhost:8084
```

Пример:

```bash
curl "http://localhost:8084/public/demo-business/services/1/branches/1/employees/1/slots?date=2026-05-18"
```

Ответ `200 OK`:

```json
[
  {
    "start_time": "2026-05-18T09:00:00+07:00",
    "end_time": "2026-05-18T09:30:00+07:00",
    "timezone": "Asia/Novosibirsk"
  },
  {
    "start_time": "2026-05-18T09:15:00+07:00",
    "end_time": "2026-05-18T09:45:00+07:00",
    "timezone": "Asia/Novosibirsk"
  }
]
```

Как считаются слоты:

- проверяется, что `registrationSlug`, `serviceId`, `branchId` и `employeeId` относятся к одному бизнесу;
- длительность берется из `services.duration_minutes`;
- рабочее время берется из `employee_schedules.starts_at` и `employee_schedules.ends_at`;
- интервалы, занятые записями со статусами `pending` и `confirmed`, исключаются;
- прошедшие слоты не возвращаются;
- шаг сетки - 15 минут;
- время считается в часовом поясе филиала.

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

```http
POST /public/{registrationSlug}/appointments
Host: localhost:8084
Content-Type: application/json
```

Пример:

```bash
curl -X POST "http://localhost:8084/public/demo-business/appointments" \
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

`start_time` передается в формате RFC3339.

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
  "comment": "First visit",
  "ics_download_url": "/public/appointments/10/calendar.ics?token=...",
  "google_calendar_export_url": "/public/appointments/10/google-calendar/auth-url?token=..."
}
```

`ics_download_url` возвращается всегда, `google_calendar_export_url` - только если настроен Google OAuth (см. раздел 6).

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

`409 appointment is not available...` означает, что выбранная комбинация slug, филиала, услуги, мастера и времени не прошла проверку. Например, услуга не принадлежит филиалу, мастер не работает в филиале, мастер не оказывает услугу, slug не принадлежит бизнесу филиала или время вне расписания.

Записи без slug не создаются: старый endpoint `POST /appointments` удален, потому что позволял записаться в любой бизнес без проверки публичной ссылки. Пустой или чужой slug отклоняется.

## 6. Добавление записи в календарь

После создания записи на странице доступны три способа:

1. **Скачать файл `.ics`** - `GET /public/appointments/{id}/calendar.ics?token=...` (ссылка из `ics_download_url`). Файл открывается в Google Календаре, Apple Calendar, Outlook. Ключи Google не нужны.
2. **Открыть в Google Календаре** - ссылку вида `https://calendar.google.com/calendar/render?action=TEMPLATE&...` формирует фронтенд, она открывает форму нового события с заполненными полями. Ключи Google не нужны.
3. **Сохранить в Google Календарь автоматически** - через OAuth: пользователь входит в Google, событие создается в его календаре через Google Calendar API, после чего пользователя возвращает на `/public/{registrationSlug}?calendar=success|error`. Кнопка показывается, только если заданы все `GOOGLE_*`.

### Как включить Google OAuth

1. Google Cloud Console: создать проект, в **APIs & Services -> Library** включить **Google Calendar API**.
2. **OAuth consent screen**: тип *External*, добавить scope `https://www.googleapis.com/auth/calendar.events`, в **Test users** добавить Google-аккаунты для проверки (пока приложение не прошло верификацию, войти могут только они).
3. **Credentials -> Create credentials -> OAuth client ID**, тип *Web application*, в **Authorized redirect URIs** добавить `http://localhost:8084/google-calendar/callback`.
4. Прописать в `.env`:

```env
GOOGLE_CLIENT_ID=...apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=...
GOOGLE_REDIRECT_URL=http://localhost:8084/google-calendar/callback
FRONTEND_URL=http://localhost:5173
```

5. Перезапустить booking: `docker compose up -d booking`. Если какой-то из `GOOGLE_*` не задан, booking пишет в лог `google calendar oauth export is disabled`, а кнопка автоматического сохранения скрыта.

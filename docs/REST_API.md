# REST API Online Queue Management System

Описание HTTP API Go-сервисов. Маршруты, тела запросов и ответов, правила доступа сверены с исходным кодом.

## Базовые адреса

Порты по умолчанию в Docker Compose:

| Сервис | Адрес |
| --- | --- |
| Auth | `http://localhost:8082` |
| Registration | `http://localhost:8081` |
| Branches | `http://localhost:8083` |
| Booking | `http://localhost:8084` |
| Bot | `http://localhost:8085` |

## Аутентификация

Защищенные endpoints читают JWT из HTTP-only cookie:

```http
Cookie: access_token=<jwt>
```

`POST /auth/login` устанавливает две cookie:

- `access_token`, path `/`
- `refresh_token`, path `/auth`

Роли в системе:

- `super_admin`
- `business_admin`
- `manager`
- `employee`

Большинство endpoints кабинета доступны только `business_admin` и/или `manager`.

Общие ошибки авторизации:

```text
401 unauthorized
403 forbidden
```

## Сервис Auth

### `GET /health`

Возвращает `200 OK` с телом `ok`.

### `POST /auth/login`

Аутентифицирует пользователя и устанавливает cookie.

Запрос:

```json
{
  "login": "owner@example.com",
  "password": "password"
}
```

Ответ `200 OK`:

```json
{
  "message": "ok"
}
```

Ошибки:

- `400` некорректный JSON или нет login/password
- `401` неверный логин или пароль
- `405` метод не поддерживается
- `500` внутренняя ошибка

### `POST /auth/refresh`

Обновляет cookie по `refresh_token`.

Ответ `200 OK`:

```json
{
  "message": "ok"
}
```

Ошибки:

- `401` refresh-токен отсутствует или недействителен
- `405` метод не поддерживается
- `500` внутренняя ошибка

### `POST /auth/logout`

Удаляет cookie. Отсутствие refresh-cookie ошибкой не считается.

Ответ `200 OK`:

```json
{
  "message": "ok"
}
```

### `GET /auth/me`

Требует `access_token`.

Ответ `200 OK`:

```json
{
  "user_id": 1,
  "login": "owner@example.com",
  "role_id": 2,
  "business_id": 7,
  "branch_id": 11
}
```

Поле `branch_id` отсутствует, если пользователь не привязан к филиалу.

## Сервис Registration

### `GET /health`

Возвращает `200 OK`.

### `GET /ping`

Возвращает `pong`. В коде помечен как тестовый endpoint.

### `POST /register`

Начинает регистрацию бизнеса и пользователя, отправляет код подтверждения на почту.

Запрос:

```json
{
  "email": "owner@example.com",
  "password": "password",
  "business_name": "Beautiful Salon",
  "business_type": "salon"
}
```

Ответ `200 OK`:

```json
{
  "status": "pending",
  "registration_id": "registration-id"
}
```

Ошибки:

- `400` некорректный JSON
- `500` ошибка валидации, репозитория, Redis или очереди писем

### `POST /verify`

Подтверждает незавершенную регистрацию.

Запрос:

```json
{
  "registration_id": "registration-id",
  "code": "123456"
}
```

Ответ `200 OK`:

```json
{
  "status": "verified"
}
```

Ошибки:

- `400` некорректный JSON, неверный или пустой код, регистрация истекла или не найдена

### `POST /resend`

Повторно отправляет код регистрации.

Запрос:

```json
{
  "registration_id": "registration-id"
}
```

Ответ `200 OK`:

```json
{
  "status": "resended"
}
```

Ошибки:

- `400` некорректный JSON или ошибка валидации

### `POST /password-recovery`

Начинает восстановление пароля.

Запрос:

```json
{
  "email": "owner@example.com"
}
```

Ответ `200 OK`:

```json
{
  "status": "password_recovery_pending",
  "recovery_id": "recovery-id"
}
```

Ошибки:

- `400` некорректный JSON или нет email
- `500` ошибка сервиса, репозитория или отправки письма

### `POST /password-recovery/confirm`

Подтверждает код восстановления пароля. Временный пароль приходит на почту.

Запрос:

```json
{
  "recovery_id": "recovery-id",
  "code": "123456"
}
```

Ответ `200 OK`:

```json
{
  "status": "password_recovery_completed"
}
```

Ошибки:

- `400` некорректный JSON, не заполнены поля, неверный код, восстановление истекло

## Сервис Branches

Защищенные endpoints требуют `access_token`.

Форматы объектов в ответах:

Филиал:

```json
{
  "id": 1,
  "name": "Central",
  "address": "Main street"
}
```

Сотрудник:

```json
{
  "id": 1,
  "name": "Anna",
  "surname": "Ivanova",
  "position": "Master"
}
```

Услуга:

```json
{
  "id": 1,
  "name": "Haircut",
  "duration_minutes": 30,
  "price": 1500
}
```

### `GET /health`

Возвращает `200 OK`.

### `GET /branches`

Возвращает филиалы, доступные текущему пользователю.

Доступ:

- `business_admin`: все филиалы своего бизнеса
- `manager`: только свой филиал

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Ошибки:

- `401` не авторизован
- `500` ошибка сервиса или репозитория, запрещенная роль, филиал не найден

### `GET /branches/{id}/clients`

Возвращает клиентов, у которых есть записи в филиале.

Доступ:

- `business_admin`: филиал должен принадлежать его бизнесу
- `manager`: `{id}` должен совпадать с филиалом менеджера

Ответ `200 OK`:

```json
[
  {
    "id": 5,
    "email": "client@example.com",
    "phone": "+79990000000",
    "name": "Alex",
    "surname": "Stone",
    "tg_username": "alex_stone",
    "created_at": "2026-05-18T09:00:00Z"
  }
]
```

Ошибки:

- `400` некорректный id филиала
- `401` не авторизован
- `403` доступ запрещен
- `404` филиал не найден
- `500` ошибка репозитория

### `GET /branches/{id}/bookings?date=YYYY-MM-DD`

Возвращает записи филиала на дату.

Доступ такой же, как у `/branches/{id}/clients`.

Ответ `200 OK`:

```json
[
  {
    "id": 21,
    "branch_id": 11,
    "client": {
      "id": 5,
      "name": "Alex",
      "surname": "Stone",
      "created_at": "2026-05-18T09:00:00Z"
    },
    "employee_id": 3,
    "employee_name": "Anna",
    "employee_surname": "Ivanova",
    "service_id": 9,
    "service_name": "Haircut",
    "start_time": "2026-05-18T09:00:00Z",
    "end_time": "2026-05-18T09:30:00Z",
    "status": "pending",
    "comment": "First visit",
    "created_at": "2026-05-18T08:00:00Z"
  }
]
```

Ошибки:

- `400` некорректный id филиала или дата
- `401` не авторизован
- `403` доступ запрещен
- `404` филиал не найден
- `500` ошибка репозитория

### `GET /branches/{id}/employees`

Возвращает сотрудников филиала.

Доступ:

- `business_admin`: филиал должен принадлежать его бизнесу
- `manager`: только свой филиал

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

- `400` некорректный id филиала
- `401` не авторизован
- `500` ошибка сервиса или репозитория, запрещенная роль, филиал не найден

### `GET /services`

Возвращает услуги текущего бизнеса.

Доступ:

- только `business_admin`

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

- `401` не авторизован
- `500` ошибка сервиса или репозитория, запрещенная роль

### `GET /businesses/{id}/registration-slug`

Возвращает `registration_slug` бизнеса (часть публичной ссылки на запись).

Доступ:

- только `business_admin`
- `{id}` должен совпадать с `business_id` пользователя

Ответ `200 OK`:

```json
{
  "business_id": 7,
  "registration_slug": "beautiful-salon"
}
```

Ошибки:

- `400` некорректный id бизнеса
- `401` не авторизован
- `403` доступ запрещен
- `404` бизнес не найден или slug не задан
- `500` ошибка репозитория или внутренняя ошибка

### `GET /services/{serviceId}/branches`

Возвращает филиалы, в которых доступна услуга.

Доступ:

- только `business_admin`

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Ошибки:

- `400` некорректный id услуги
- `401` не авторизован
- `500` ошибка сервиса или репозитория, запрещенная роль

### `GET /services/{serviceId}/branches/{branchId}/employees`

Возвращает сотрудников филиала, которые оказывают услугу.

Доступ:

- `business_admin`: филиал должен принадлежать его бизнесу
- `manager`: только свой филиал

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

- `400` некорректный id услуги или филиала
- `401` не авторизован
- `500` ошибка сервиса или репозитория, запрещенная роль, филиал не найден

### `GET /public/{registrationSlug}/services`

Публичный endpoint. Возвращает услуги бизнеса по его slug.

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

- `400` некорректный slug
- `500` ошибка репозитория или неизвестный slug

### `GET /public/{registrationSlug}/services/{serviceId}/branches`

Публичный endpoint. Возвращает филиалы, где оказывается услуга.

Ответ `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Ошибки:

- `400` некорректный id услуги или slug
- `500` ошибка репозитория или неизвестный slug

### `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees`

Публичный endpoint. Возвращает сотрудников филиала, которые оказывают услугу.

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

- `400` некорректный id услуги, филиала или slug
- `500` ошибка репозитория, филиал чужого бизнеса, филиал не найден, неизвестный slug

## Сервис Booking

Статусы записи:

- `pending`
- `confirmed`
- `completed`
- `cancelled`

### `GET /health`

Возвращает `200 OK`.

### `POST /public/{registrationSlug}/appointments`

Публичная запись клиента к бизнесу, без авторизации. Филиал, сотрудник и услуга должны принадлежать бизнесу с этим `registrationSlug`; пустой или чужой slug отклоняется.

Запрос:

```json
{
  "client": {
    "email": "client@example.com",
    "phone": "+79990000000",
    "name": "Ivan",
    "surname": "Petrov",
    "tg_username": "ivan_petrov"
  },
  "branch_id": 1,
  "employee_id": 1,
  "service_id": 1,
  "start_time": "2026-05-18T09:00:00+07:00",
  "comment": "First visit"
}
```

Обязательные поля:

- `client.name`
- `client.surname`
- `client.phone`
- `branch_id`
- `employee_id`
- `service_id`
- `start_time` в формате RFC3339

Ответ `201 Created`:

```json
{
  "appointment_id": 10,
  "client_id": 25,
  "branch": {
    "id": 1,
    "name": "Central"
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

`ics_download_url` возвращается всегда. `google_calendar_export_url` - только если настроен Google OAuth (`GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`). Обе ссылки используют один токен экспорта (Redis, TTL 7 дней).

Ошибки:

- `400` некорректное тело, id, данные клиента, контакт, slug или время начала
- `409` слот занят или недоступен
- `500` ошибка репозитория или внутренняя ошибка

### `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees/{employeeId}/slots?date=YYYY-MM-DD`

Публичный endpoint. Возвращает свободные слоты сотрудника на дату.

Ответ `200 OK`:

```json
[
  {
    "start_time": "2026-05-18T09:00:00+07:00",
    "end_time": "2026-05-18T09:30:00+07:00",
    "timezone": "Asia/Novosibirsk"
  }
]
```

Ошибки:

- `400` некорректный id услуги, филиала, сотрудника, дата, slug или время
- `500` ошибка репозитория или внутренняя ошибка

### `GET /employees/{id}/appointments`

Требует `access_token`.

Доступ:

- `business_admin`
- `manager`

Ответ `200 OK`:

```json
[
  {
    "id": 21,
    "client_id": 5,
    "branch_id": 11,
    "employee_id": 3,
    "service_id": 9,
    "start_time": "2026-05-18T09:00:00Z",
    "end_time": "2026-05-18T09:30:00Z",
    "status": "pending",
    "comment": "First visit"
  }
]
```

Ошибки:

- `400` некорректный id сотрудника
- `401` не авторизован
- `403` доступ запрещен
- `500` ошибка репозитория или внутренняя ошибка

### `GET /appointments/{id}`

Требует `access_token`.

Доступ:

- `business_admin`: запись должна принадлежать его бизнесу
- `manager`: запись должна быть в филиале менеджера

Ответ `200 OK`:

```json
{
  "id": 21,
  "client_id": 5,
  "branch_id": 11,
  "employee_id": 3,
  "service_id": 9,
  "start_time": "2026-05-18T09:00:00Z",
  "end_time": "2026-05-18T09:30:00Z",
  "status": "pending",
  "comment": "First visit"
}
```

Ошибки:

- `400` некорректный id записи
- `401` не авторизован
- `403` доступ запрещен
- `404` запись не найдена
- `500` ошибка репозитория или внутренняя ошибка

### `PATCH /appointments/{id}/cancel`

Отменяет запись. Требует `access_token`, правила доступа как у `GET /appointments/{id}`.

Ответ:

```text
204 No Content
```

Ошибки:

- `400` некорректный id записи
- `401` не авторизован
- `403` доступ запрещен
- `404` запись не найдена
- `409` запись уже отменена, завершена или недоступна
- `500` ошибка репозитория или внутренняя ошибка

### `GET /google-calendar/auth-url`

Требует `access_token`.

Создает OAuth state в Redis и возвращает ссылку на вход в Google.

Ответ `200 OK`:

```json
{
  "url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

Ошибки:

- `401` не авторизован
- `503` интеграция с Google Календарем отключена
- `500` ошибка Redis или внутренняя ошибка

### `GET /google-calendar/callback?state=...&code=...`

Публичный callback для Google OAuth. Его открывает браузер пользователя, поэтому он всегда делает редирект обратно на фронтенд (`FRONTEND_URL`):

- публичная запись: `302` на `{FRONTEND_URL}/public/{registrationSlug}?calendar=success|error`;
- кабинет или неизвестный state: `302` на `{FRONTEND_URL}/admin?calendar=success|error`.

`calendar=error` означает отказ на экране согласия, истекший state, отключенную интеграцию или ошибку Google API (подробности пишутся в лог).

### `GET /public/appointments/{id}/calendar.ics?token=...`

Публичный endpoint. Отдает запись как файл iCalendar (RFC 5545), который можно импортировать в Google Календарь, Apple Calendar или Outlook. Работает без настроек Google OAuth. Доступ - по токену экспорта из `ics_download_url`.

Ответ `200 OK`: `Content-Type: text/calendar; charset=utf-8`, `Content-Disposition: attachment; filename="appointment-{id}.ics"`. В событии: название услуги, название и адрес филиала, сотрудник, комментарий и напоминание за час до начала.

Ошибки:

- `400` некорректный id записи
- `404` токен отсутствует, неизвестен или от другой записи
- `500` ошибка репозитория или внутренняя ошибка

### `POST /appointments/{id}/google-calendar`

Требует `access_token`. Выгружает запись в Google Календарь.

Доступ:

- `business_admin`: запись должна принадлежать его бизнесу
- `manager`: запись должна быть в филиале менеджера

Ответ `201 Created`:

```json
{
  "event_id": "google-event-id",
  "html_link": "https://calendar.google.com/calendar/event?..."
}
```

Ошибки:

- `400` некорректный id записи
- `401` не авторизован
- `403` доступ запрещен
- `404` запись не найдена
- `409` Google Календарь не привязан
- `503` интеграция с Google Календарем отключена
- `500` ошибка Redis, Google или внутренняя ошибка

## Сервис Bot

### `GET /health`

Возвращает `200 OK`.

### `POST /telegram/notifications`

Отправляет уведомление в Telegram по ранее привязанному телефону или username. Вызывается сервисом scheduler.

Запрос:

```json
{
  "phone": "+79990000000",
  "username": "client_username",
  "business": "Beautiful Salon",
  "service": "Haircut",
  "branch": "Central",
  "employee": "Anna Ivanova",
  "start_time": "2026-05-18T09:00:00+07:00",
  "description": "Appointment reminder"
}
```

Нужно указать хотя бы `phone` или `username`. `start_time`, если указан, должен быть в формате RFC3339.

Ответ:

```text
202 Accepted
```

Ошибки:

- `400` некорректный JSON, не указан получатель, некорректный start_time
- `502` ошибка отправки в Telegram или поиска привязки
- `503` бот отключен, потому что `TELEGRAM_BOT_TOKEN` пуст

## Сервис Scheduler

REST endpoints нет. Сервис опрашивает Redis, забирает напоминания, время которых наступило, и отправляет их по email и через сервис bot.

## Переменные окружения

Полный список переменных с комментариями - в `.env.example`, описание - в [README.md](../README.md#файл-env).

## Известные недочеты API

1. Часть маршрутов Registration и Branches зарегистрирована без указания метода. Обработчики Registration метод не проверяют. В Branches метод `GET` указан только для `/branches`, `/branches/{id}/clients`, `/branches/{id}/bookings` и `/businesses/{id}/registration-slug`; остальные маршруты филиалов и услуг принимают любой HTTP-метод.
2. Обработчики Branches часто отвечают `500` на доменные ошибки "доступ запрещен" и "не найдено" вместо `403`/`404`; у `/branches/{id}/clients` и `/branches/{id}/bookings` коды ошибок уже правильные.
3. Registration отвечает `500` на большинство ошибок сервиса в `/register` и `/password-recovery`, включая ошибки валидации.

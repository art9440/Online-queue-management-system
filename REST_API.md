# Online Queue Management System REST API

This document describes the HTTP API exposed by the current Go services.

Verification status:

- `GOCACHE=/tmp/go-build-cache go test ./...` passes.
- Routes, request DTOs, response DTOs, auth rules, and env names were checked against the source code.
- Runtime behavior with real Postgres, Redis, SMTP, Telegram, and Google Calendar still depends on valid `.env` values and reachable external services.

## Base URLs

Default Docker Compose ports:

| Service | Base URL |
| --- | --- |
| Auth | `http://localhost:8082` |
| Registration | `http://localhost:8081` |
| Branches | `http://localhost:8083` |
| Booking | `http://localhost:8084` |
| Bot | `http://localhost:8085` |

## Authentication

Authenticated endpoints read JWT from an HTTP-only cookie:

```http
Cookie: access_token=<jwt>
```

`POST /auth/login` sets both:

- `access_token`, path `/`
- `refresh_token`, path `/auth`

Roles used by the services:

- `super_admin`
- `business_admin`
- `manager`
- `employee`

Most business endpoints currently allow only `business_admin` and/or `manager`.

Common auth errors:

```text
401 unauthorized
403 forbidden
```

## Auth Service

### `GET /health`

Returns `200 OK` with body `ok`.

### `POST /auth/login`

Authenticates a user and sets auth cookies.

Request:

```json
{
  "login": "owner@example.com",
  "password": "password"
}
```

Response `200 OK`:

```json
{
  "message": "ok"
}
```

Errors:

- `400` invalid JSON or missing login/password
- `401` bad credentials
- `405` method not allowed
- `500` internal error

### `POST /auth/refresh`

Refreshes auth cookies using `refresh_token`.

Response `200 OK`:

```json
{
  "message": "ok"
}
```

Errors:

- `401` missing/invalid refresh token
- `405` method not allowed
- `500` internal error

### `POST /auth/logout`

Clears auth cookies. Missing refresh cookie is tolerated.

Response `200 OK`:

```json
{
  "message": "ok"
}
```

### `GET /auth/me`

Requires `access_token`.

Response `200 OK`:

```json
{
  "user_id": 1,
  "login": "owner@example.com",
  "role_id": 2,
  "business_id": 7,
  "branch_id": 11
}
```

`branch_id` is omitted when absent.

## Registration Service

### `GET /health`

Returns `200 OK`.

### `GET /ping`

Returns `pong`. This is marked in code as a test endpoint.

### `POST /register`

Starts business/user registration and sends a verification code.

Request:

```json
{
  "email": "owner@example.com",
  "password": "password",
  "business_name": "Beautiful Salon",
  "business_type": "salon"
}
```

Response `200 OK`:

```json
{
  "status": "pending",
  "registration_id": "registration-id"
}
```

Errors:

- `400` invalid JSON
- `500` validation, repository, Redis, or email queue error

### `POST /verify`

Verifies a pending registration.

Request:

```json
{
  "registration_id": "registration-id",
  "code": "123456"
}
```

Response `200 OK`:

```json
{
  "status": "verified"
}
```

Errors:

- `400` invalid JSON, invalid/missing code, expired or missing registration

### `POST /resend`

Resends the registration code.

Request:

```json
{
  "registration_id": "registration-id"
}
```

Response `200 OK`:

```json
{
  "status": "resended"
}
```

Errors:

- `400` invalid JSON or service validation error

### `POST /password-recovery`

Starts password recovery.

Request:

```json
{
  "email": "owner@example.com"
}
```

Response `200 OK`:

```json
{
  "status": "password_recovery_pending",
  "recovery_id": "recovery-id"
}
```

Errors:

- `400` invalid JSON or missing email
- `500` service/repository/email error

### `POST /password-recovery/confirm`

Confirms password recovery code.

Request:

```json
{
  "recovery_id": "recovery-id",
  "code": "123456"
}
```

Response `200 OK`:

```json
{
  "status": "password_recovery_completed"
}
```

Errors:

- `400` invalid JSON, missing fields, invalid code, expired recovery

## Branches Service

Authenticated endpoints require `access_token`.

Response shapes:

Branch:

```json
{
  "id": 1,
  "name": "Central",
  "address": "Main street"
}
```

Employee:

```json
{
  "id": 1,
  "name": "Anna",
  "surname": "Ivanova",
  "position": "Master"
}
```

Service:

```json
{
  "id": 1,
  "name": "Haircut",
  "duration_minutes": 30,
  "price": 1500
}
```

### `GET /health`

Returns `200 OK`.

### `GET /branches`

Returns branches visible to the current user.

Access:

- `business_admin`: all branches of own business
- `manager`: own branch only

Response `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Errors:

- `401` unauthorized
- `500` service/repository error, forbidden role, branch not found

### `GET /branches/{id}/clients`

Returns clients who have appointments in the branch.

Access:

- `business_admin`: branch must belong to own business
- `manager`: `{id}` must equal manager branch

Response `200 OK`:

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

Errors:

- `400` invalid branch id
- `401` unauthorized
- `403` forbidden
- `404` branch not found
- `500` repository error

### `GET /branches/{id}/bookings?date=YYYY-MM-DD`

Returns appointments for a branch and date.

Access is the same as `/branches/{id}/clients`.

Response `200 OK`:

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

Errors:

- `400` invalid branch id or invalid date
- `401` unauthorized
- `403` forbidden
- `404` branch not found
- `500` repository error

### `GET /branches/{id}/employees`

Returns employees for a branch.

Access:

- `business_admin`: branch must belong to own business
- `manager`: branch must be own branch

Response `200 OK`:

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

Errors:

- `400` invalid branch id
- `401` unauthorized
- `500` service/repository error, forbidden role, branch not found

### `GET /services`

Returns services for the current business.

Access:

- `business_admin` only

Response `200 OK`:

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

Errors:

- `401` unauthorized
- `500` service/repository error or forbidden role

### `GET /businesses/{id}/registration-slug`

Returns `registration_slug` for a business.

Access:

- `business_admin` only
- `{id}` must equal the authenticated user's `business_id`

Response `200 OK`:

```json
{
  "business_id": 7,
  "registration_slug": "beautiful-salon"
}
```

Errors:

- `400` invalid business id
- `401` unauthorized
- `403` forbidden
- `404` business not found or registration slug is not set
- `500` repository/internal error

### `GET /services/{serviceId}/branches`

Returns branches where a service is available.

Access:

- `business_admin` only

Response `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Errors:

- `400` invalid service id
- `401` unauthorized
- `500` service/repository error or forbidden role

### `GET /services/{serviceId}/branches/{branchId}/employees`

Returns employees who can perform a service in a branch.

Access:

- `business_admin`: branch must belong to own business
- `manager`: branch must be own branch

Response `200 OK`:

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

Errors:

- `400` invalid service id or branch id
- `401` unauthorized
- `500` service/repository error, forbidden role, branch not found

### `GET /public/{registrationSlug}/services`

Public endpoint. Returns services for a business registration slug.

Response `200 OK`:

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

Errors:

- `400` invalid registration slug
- `500` repository error or unknown slug

### `GET /public/{registrationSlug}/services/{serviceId}/branches`

Public endpoint. Returns branches for a service.

Response `200 OK`:

```json
[
  {
    "id": 1,
    "name": "Central",
    "address": "Main street"
  }
]
```

Errors:

- `400` invalid service id or registration slug
- `500` repository error or unknown slug

### `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees`

Public endpoint. Returns employees for service and branch.

Response `200 OK`:

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

Errors:

- `400` invalid service id, branch id, or registration slug
- `500` repository error, forbidden branch, branch not found, or unknown slug

## Booking Service

Appointment statuses:

- `pending`
- `confirmed`
- `completed`
- `cancelled`

### `GET /health`

Returns `200 OK`.

### `POST /appointments`

Creates an appointment without a `registrationSlug`. In the current code this endpoint is not authenticated.

Request:

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

Required:

- `client.name`
- `client.surname`
- `client.phone`
- `branch_id`
- `employee_id`
- `service_id`
- `start_time` as RFC3339

Response `201 Created`:

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
  "comment": "First visit"
}
```

Errors:

- `400` invalid body, IDs, client data, contact, registration slug, start time
- `409` busy or unavailable appointment slot
- `500` repository/internal error

### `POST /public/{registrationSlug}/appointments`

Public client booking endpoint for a business slug. Request/response are the same as `POST /appointments`.

Errors are the same as `POST /appointments`.

### `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees/{employeeId}/slots?date=YYYY-MM-DD`

Public endpoint. Returns available slots for one employee on one date.

Response `200 OK`:

```json
[
  {
    "start_time": "2026-05-18T09:00:00+07:00",
    "end_time": "2026-05-18T09:30:00+07:00",
    "timezone": "Asia/Novosibirsk"
  }
]
```

Errors:

- `400` invalid service id, branch id, employee id, date, registration slug, or start time
- `500` repository/internal error

### `GET /employees/{id}/appointments`

Requires `access_token`.

Access:

- `business_admin`
- `manager`

Response `200 OK`:

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

Errors:

- `400` invalid employee id
- `401` unauthorized
- `403` forbidden
- `500` repository/internal error

### `GET /appointments/{id}`

Requires `access_token`.

Access:

- `business_admin`: appointment must belong to own business
- `manager`: appointment branch must equal manager branch

Response `200 OK`:

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

Errors:

- `400` invalid appointment id
- `401` unauthorized
- `403` forbidden
- `404` appointment not found
- `500` repository/internal error

### `PATCH /appointments/{id}/cancel`

Requires `access_token`. Access rules are the same as `GET /appointments/{id}`.

Response:

```text
204 No Content
```

Errors:

- `400` invalid appointment id
- `401` unauthorized
- `403` forbidden
- `404` appointment not found
- `409` appointment already cancelled/completed or unavailable
- `500` repository/internal error

### `GET /google-calendar/auth-url`

Requires `access_token`.

Creates an OAuth state in Redis and returns a Google OAuth URL.

Response `200 OK`:

```json
{
  "url": "https://accounts.google.com/o/oauth2/auth?..."
}
```

Errors:

- `401` unauthorized
- `503` Google Calendar integration disabled
- `500` Redis/internal error

### `GET /google-calendar/callback?state=...&code=...`

Public callback endpoint for Google OAuth.

Response:

```text
204 No Content
```

Errors:

- `409` missing state/code or Google Calendar not linked
- `503` Google Calendar integration disabled
- `500` Redis/Google/internal error

### `POST /appointments/{id}/google-calendar`

Requires `access_token`. Exports an appointment to Google Calendar.

Access:

- `business_admin`: appointment must belong to own business
- `manager`: appointment branch must equal manager branch

Response `201 Created`:

```json
{
  "event_id": "google-event-id",
  "html_link": "https://calendar.google.com/calendar/event?..."
}
```

Errors:

- `400` invalid appointment id
- `401` unauthorized
- `403` forbidden
- `404` appointment not found
- `409` Google Calendar is not linked
- `503` Google Calendar integration disabled
- `500` Redis/Google/internal error

## Bot Service

### `GET /health`

Returns `200 OK`.

### `POST /telegram/notifications`

Sends a Telegram notification to a previously bound phone or username.

Request:

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

At least one of `phone` or `username` is required. `start_time`, when provided, must be RFC3339.

Response:

```text
202 Accepted
```

Errors:

- `400` invalid JSON, missing recipient, invalid start_time
- `502` Telegram send/binding lookup error
- `503` Telegram bot is disabled because `TELEGRAM_BOT_TOKEN` is empty

## Scheduler Service

The scheduler service does not expose REST endpoints. It polls Redis for due notifications and dispatches through email and bot HTTP integration.

## Required Environment

Common database env:

| Variable | Required by |
| --- | --- |
| `DB_HOST` | DB-backed services, migrate |
| `DB_PORT` | DB-backed services, migrate |
| `POSTGRES_DB` | DB-backed services, migrate, Postgres container |
| `DB_USER` | DB-backed services, migrate |
| `DB_PASSWORD` | DB-backed services, migrate |
| `DB_SSLMODE` | optional, default `disable` |
| `DB_DSN` | optional override for constructed DSN |
| `POSTGRES_USER` | Postgres container |
| `POSTGRES_PASSWORD` | Postgres container |

Common Redis env:

| Variable | Required by |
| --- | --- |
| `REDIS_ADDR` | auth, registration, booking, bot, scheduler |
| `REDIS_PASSWORD` | auth optional, registration/bot/scheduler currently require non-empty, booking optional |
| `REDIS_DB` | auth, registration, booking, bot, scheduler |

Ports and JWT:

| Variable | Required by |
| --- | --- |
| `AUTH_PORT` | auth, default `8082` |
| `REGISTRATION_PORT` | registration |
| `BRANCHES_PORT` | branches |
| `BOOKING_PORT` | booking |
| `BOT_PORT` | bot, default `8085` |
| `JWT_ACCESS_SECRET` | auth, branches, booking |
| `JWT_REFRESH_SECRET` | auth |
| `ACCESS_TOKEN_TTL` | auth, duration like `15m` |
| `REFRESH_TOKEN_TTL` | auth, duration like `168h` |
| `COOKIE_SECURE` | auth, default `false` |

Email:

| Variable | Required by |
| --- | --- |
| `SMTP_HOST` | registration, scheduler |
| `SMTP_PORT` | registration, scheduler |
| `SMTP_USER` | registration, scheduler |
| `SMTP_PASS` | registration, scheduler |
| `EMAIL_TIMEOUT` | optional, default `20s` |

Registration queue:

| Variable | Required by |
| --- | --- |
| `NUM_WORKERS` | optional, default `10` |
| `RATE_LIMIT` | optional, default `30s` |
| `WRK_TIMEOUT` | optional, default `10s` |
| `APP_ENV` | optional |

Booking Google Calendar:

| Variable | Required by |
| --- | --- |
| `GOOGLE_CLIENT_ID` | optional; empty value makes auth URL/export fail at runtime |
| `GOOGLE_CLIENT_SECRET` | optional; empty value makes OAuth exchange/export fail at runtime |
| `GOOGLE_REDIRECT_URL` | optional; should match Google OAuth settings |

Bot and scheduler:

| Variable | Required by |
| --- | --- |
| `TELEGRAM_BOT_TOKEN` | optional; empty disables bot polling and sending |
| `TELEGRAM_POLLING_TIMEOUT` | optional, default `30` |
| `TELEGRAM_POLLING_INTERVAL` | optional, default `1s` |
| `SCHEDULER_POLL_INTERVAL` | optional, default `30s` |
| `SCHEDULER_BATCH_SIZE` | optional, default `100` |
| `BOT_URL` | optional, default `http://bot:8085` |

## Correctness Notes

The code compiles and tests pass, but these points matter before production use:

1. Several Registration and Branches routes are registered without method-specific patterns. Registration handlers also do not check HTTP method. Branches has method-specific `GET` patterns only for `/branches`, `/branches/{id}/clients`, and `/branches/{id}/bookings`; other branch/service routes currently accept any HTTP method and execute the GET handler.
2. `POST /appointments` in Booking is unauthenticated in the current router. If this is meant for internal/admin use, it should be protected or removed in favor of `POST /public/{registrationSlug}/appointments`.
3. Branches handlers often map domain errors like forbidden/not found to `500` instead of `403`/`404`; `/branches/{id}/clients` and `/branches/{id}/bookings` already have better error mapping.
4. Registration maps most service errors from `/register` and `/password-recovery` to `500`, including validation-style errors. That may be acceptable for now but is not ideal API behavior.
5. `.env` currently has no Google Calendar or Telegram variables listed. Google Calendar endpoints and Telegram sending will not fully work until those values are added where needed.
6. The Auth and Branches apps log `JWT_ACCESS_SECRET`. That should be removed before real deployment.

# Online Queue Management System — обзор проекта

Сводный документ по архитектуре проекта: сервисы, их REST API, взаимодействие между собой, схема PostgreSQL, использование Redis, устройство фронтенда, порты и запуск.

> Не описаны намеренно: метрики (Prometheus/Grafana), логирование через Elasticsearch/Logstash/Kibana, трейсинг (Jaeger/OpenTelemetry), strace — это отдельная тема, вынесенная за скобки этого обзора.

---

## 1. Общая архитектура

Проект — **микросервисный монорепозиторий на Go** (модуль `Online-queue-management-system`, Go 1.26). Каждый сервис лежит в `services/<name>`, имеет свою точку входа `cmd/main.go`, свой Dockerfile и свою секцию в корневом `docker-compose.yml`. Общий код (JWT, конфиги, логгер, HTTP-клиент, middleware, email, Redis-клиент) вынесен в `libs/*` и подключается всеми сервисами как shared-библиотеки одного go-модуля.

Роутинг во всех сервисах — стандартный `net/http.ServeMux` (Go 1.22+, с шаблонами `"GET /branches/{id}"`), сторонние роутеры (chi/gin/echo) не используются.

Предметная область: SaaS для онлайн-записи клиентов в очередь/на приём (салоны, сервисные компании и т.п.) — филиалы, сотрудники, услуги, клиенты, бронирования (appointments), с ролями `super_admin / business_admin / manager / employee` и публичной страницей самозаписи по уникальной ссылке бизнеса.

### Список сервисов

| Сервис | Назначение | Путь | Порт (compose) |
|---|---|---|---|
| **auth** | Аутентификация, выдача/рефреш JWT, сессии в Redis | `services/auth` | 8082 |
| **registration** | Регистрация бизнеса/владельца, подтверждение email, восстановление пароля | `services/registration` | 8081 |
| **branches** | Филиалы, сотрудники, услуги, клиенты бизнеса, публичный каталог для записи | `services/branches` | 8083 |
| **booking** | Создание/просмотр/отмена записей, свободные слоты, постановка email-напоминаний | `services/booking` | 8084 |
| **scheduler** | Фоновый воркер: вычитывает «созревшие» напоминания из Redis и рассылает их по email | `services/scheduler` | — (без HTTP) |
| **migrate** | Прогон goose-миграций Postgres при старте окружения (job, не долгоживущий) | `services/migrate` | — |
| **notification** | Пустая заготовка (`package notification`), функционал не используется | `services/notification` | — |

---

## 2. Как сервисы общаются между собой

Message broker (Kafka/RabbitMQ/NATS) не используется. gRPC — тоже нет (в зависимостях есть только как транзитивная часть OTel-трейсинга). Взаимодействие построено на трёх механизмах:

1. **REST/HTTP синхронно** — только фронтенд → сервисы; сервисы друг друга по HTTP не вызывают.
2. **Redis как общая асинхронная шина/очередь** — `booking` (producer) кладёт уведомления о записи в Redis Sorted Set, `scheduler` (consumer, poll по тикеру) вычитывает «созревшие» по времени и рассылает по email. Прямой связи booking↔scheduler по HTTP нет.
3. **Общий Postgres** — `auth`, `branches`, `booking`, `registration` читают/пишут одну БД, но каждый работает со своим набором таблиц/репозиториев; единого API-gateway между сервисами нет.

Дополнительно: **registration/scheduler → SMTP** (`libs/email`).

Важный момент по авторизации между сервисами: **auth только выпускает JWT**, а остальные сервисы (branches, booking) валидируют токен самостоятельно через общую библиотеку `libs/auth` — отдельного auth-gateway/middleware-сервиса нет, шлюза перед сервисами тоже нет (клиент обращается к каждому сервису напрямую по своему порту).

---

## 3. REST API по сервисам

### Auth — порт 8082
| Метод и путь | Доступ | Описание |
|---|---|---|
| `GET /health` | публично | health-check |
| `POST /auth/login` | публично | логин, ставит cookies `access_token` (path `/`) и `refresh_token` (path `/auth`) |
| `POST /auth/refresh` | по `refresh_token` cookie | обновление access-токена |
| `POST /auth/logout` | публично | чистит cookies |
| `GET /auth/me` | по `access_token` cookie | возвращает `user_id, login, role_id, business_id, branch_id?` |

### Registration — порт 8081
| Метод и путь | Описание |
|---|---|
| `GET /health`, `GET /ping` | health-check |
| `POST /register` | email+пароль+бизнес → pending-регистрация в Redis + код на email |
| `POST /verify` | код подтверждения → создание пользователя и бизнеса в Postgres, генерация `registration_slug` |
| `POST /resend` | повторная отправка кода |
| `POST /password-recovery`, `POST /password-recovery/confirm` | восстановление пароля (временный пароль на почту) |

> Примечание: часть маршрутов зарегистрирована без строгой привязки к HTTP-методу (технический долг, см. раздел 8).

### Branches — порт 8083
Приватные (JWT-cookie, доступ проверяется по `business_id`/`branch_id` из токена):
- `GET /branches`, `GET /branches/{id}/clients`, `GET /branches/{id}/bookings?date=YYYY-MM-DD`, `GET /branches/{id}/employees`
- `GET /services` (только business_admin), `GET /services/{serviceId}/branches`, `GET /services/{serviceId}/branches/{branchId}/employees`
- `GET /businesses/{id}/registration-slug` (только business_admin, только свой бизнес)

Публичные (по `registrationSlug`, без JWT — для клиентского флоу записи):
- `GET /public/{registrationSlug}/services`
- `GET /public/{registrationSlug}/services/{serviceId}/branches`
- `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees`

### Booking — порт 8084
| Метод и путь | Доступ | Описание |
|---|---|---|
| `POST /public/{registrationSlug}/appointments` | публично | единственный способ создания записи; slug обязателен, филиал должен принадлежать бизнесу slug |
| `GET /public/{registrationSlug}/services/{serviceId}/branches/{branchId}/employees/{employeeId}/slots?date=` | публично | свободные слоты мастера (шаг 15 мин) |
| `GET /employees/{id}/appointments` | приватно | записи сотрудника |
| `GET /appointments/{id}` | приватно | запись по id |
| `PATCH /appointments/{id}/cancel` | приватно | отмена записи |
| `GET /health` | публично | health-check |

Статусы записи: `pending → confirmed → completed`, либо `cancelled`.

### Scheduler
REST-эндпоинтов нет — фоновый воркер (тикер с интервалом `SCHEDULER_POLL_INTERVAL`), вычитывает уведомления из Redis пачками (`SCHEDULER_BATCH_SIZE`) и рассылает по email. Задачи с другим `channel` (остатки старых telegram-уведомлений) удаляются без отправки.

---

## 4. Что может делать пользователь

Роли (`libs/auth/roles.go` на бэке, `frontend/src/constants/roles.js` на фронте): `super_admin`, `business_admin`, `manager`, `employee` (последняя роль пока не имеет своих защищённых маршрутов на фронте — задел на будущее).

- **business_admin (владелец/админ бизнеса)**
  - Регистрирует бизнес и себя (`POST /register` → код на email → `POST /verify`), получает `registration_slug` — уникальную ссылку для клиентской онлайн-записи.
  - Видит все филиалы своего бизнеса, все услуги бизнеса, сотрудников и клиентов по филиалам, записи по филиалу и дате.
  - Открывает конкретную запись, отменяет её.
- **manager (менеджер филиала)** — те же операции просмотра, но строго в рамках своего `branch_id` из JWT; попытка обратиться к чужому филиалу отклоняется и на бэке, и на фронте (см. раздел 7).
- **Клиент без аккаунта** — по публичной ссылке `/public/{slug}` (описано в `docs/PUBLIC_BOOKING_API.md`): выбирает услугу → филиал, где она доступна → мастера → свободный слот на дату → отправляет данные (имя, фамилия, телефон, email) → получает подтверждение записи; если указан email — напоминания за 24ч и за 1ч.
- **Восстановление доступа** — любой пользователь инициирует `POST /password-recovery`, подтверждает код, получает временный пароль на почту.

Ключевые сущности: `Business` (со `registration_slug`), `Branch` (с `timezone`), `Employee` (с `employee_schedules`), `Service` (`duration_minutes`, `price`), `Client` (email/телефон), `Appointment` (клиент+филиал+сотрудник+услуга+время+статус).

---

## 5. Авторизация и аутентификация

- **JWT (HS256)**, `github.com/golang-jwt/jwt/v5`, две пары секретов: `JWT_ACCESS_SECRET` (`ACCESS_TOKEN_TTL`, в `.env` = 15m) и `JWT_REFRESH_SECRET` (`REFRESH_TOKEN_TTL` = 168h).
- Токены передаются через **HTTP-only cookies** (`access_token` на path `/`, `refresh_token` на path `/auth`), не через заголовок `Authorization`. Флаг `Secure` управляется `COOKIE_SECURE`.
- `auth`-сервис хранит refresh-сессии в Redis (whitelist-паттерн, инвалидация по logout).
- Claims: `user_id, login, role_id, role_name, business_id, branch_id (nullable)`.
- Общая библиотека `libs/auth` (не сервис, а shared-пакет) подключена к `auth`, `branches`, `booking`: `TokenParser.ParseAccessToken` для валидации, `auth.Middleware(parser)` — HTTP-middleware, кладёт claims в контекст, `auth.FromContext(ctx)` — чтение в хендлерах.
- Middleware применяется выборочно (не глобально) — публичные маршруты (`/public/...`, `/health`, регистрация) регистрируются без него.
- **RBAC не через middleware**, а вручную в каждом application-сервисе (`switch auth.Role(...)`) + сверка `business_id`/`branch_id` из токена с запрошенным ресурсом в БД → мультитенантная изоляция данных.

---

## 6. PostgreSQL: миграции и схема

### Миграции
Каталог `migrations/` в корне репозитория, инструмент — **goose** (`pressly/goose/v3`), файлы `<timestamp>_описание.sql` с директивами `-- +goose Up/Down`. Применяются отдельным сервисом `services/migrate` при старте окружения (ждёт готовности Postgres, затем `goose.Up`), остальные сервисы стартуют только после успешного завершения `migrate` (`depends_on: service_completed_successfully`).

| # | Миграция | Суть |
|---|---|---|
| 1 | `20260326103157_init.sql` | Базовая схема: `businesses, roles, branches, users, employees, employee_schedules (legacy-формат), services, employee_services, clients, appointments` + индексы |
| 2 | `20260326103218_seed_roles.sql` | Сид 4 ролей |
| 3 | `20260326104048_add_link_to_businesses.sql` | `businesses.registration_slug` + частичный уникальный индекс |
| 4 | `20260420161500_update_employee_schedules_to_timestamptz.sql` | `branches.timezone`; старая `employee_schedules` → переименована в `employee_schedules_legacy` (не удаляется!); создана новая `employee_schedules` на `timestamptz` |
| 5 | `20260420193000_seed_mock_branches_employees_services.sql` | Демо-бизнес, 3 филиала в Новосибирске, 6 сотрудников, услуги, расписания |
| 6 | `20260421171635_add_mock_user.sql` | Демо-пользователь `admin@test.com` (business_admin) |
| 7 | `20260511182656_add_appointments_no_overlap.sql` | Расширение `btree_gist` + EXCLUDE-constraint против пересечения записей одного сотрудника (pending/confirmed) |
| 8 | `20260517190000_set_demo_business_registration_slug.sql` | `registration_slug='demo-business'` для демо-бизнеса |
| 9 | `20260517193000_fix_demo_business_schedule_timezone.sql` | Пересчёт демо-расписаний с учётом таймзоны филиала |

Enum-типов в БД нет — статусы реализованы как `text` + `CHECK`.

### Итоговая схема (после всех миграций)

- **businesses**(id PK, name, type, registration_slug UNIQUE частично)
- **roles**(id PK, name UNIQUE) — `super_admin, business_admin, manager, employee`
- **branches**(id PK, business_id FK→businesses CASCADE, name, address, timezone DEFAULT 'UTC')
- **users**(id PK, login UNIQUE, password_hash, role_id FK→roles RESTRICT, branch_id FK→branches SET NULL nullable, business_id FK→businesses CASCADE, created_at) — аккаунты для входа в панель управления
- **employees**(id PK, branch_id FK→branches CASCADE, name, surname, position, created_at) — сотрудники, оказывающие услуги (отдельно от `users`!)
- **employee_schedules**(id PK, employee_id FK→employees CASCADE, starts_at/ends_at timestamptz, CHECK starts_at<ends_at, UNIQUE(employee_id,starts_at,ends_at))
- **employee_schedules_legacy** — «осиротевшая» таблица (day_of_week/start_time/end_time), физически существует, но нигде в коде не используется — технический долг от рефакторинга миграции №4
- **services**(id PK, branch_id FK→branches CASCADE, name, duration_minutes CHECK>0, price numeric(10,2) CHECK>=0, created_at)
- **employee_services**(employee_id, service_id — составной PK, M2M employees↔services)
- **clients**(id PK, email/phone UNIQUE nullable, name, surname, password_hash nullable, created_at)
- **appointments**(id PK, client_id FK→clients CASCADE, branch_id FK→branches CASCADE, employee_id FK→employees RESTRICT, service_id FK→services RESTRICT, start_time/end_time timestamptz, status CHECK IN (pending,confirmed,completed,cancelled), comment, created_at, CHECK start_time<end_time, **EXCLUDE USING gist** против пересечения слотов одного сотрудника для активных статусов)

### ER-схема словами
```
businesses ──< branches ──< employees ──< employee_schedules
    │              │            └──< employee_services >── services
    │              ├──< services
    │              └──< users (business_id обязателен)
roles ──< users (role_id)
clients ──< appointments >── branches / employees / services  (центральная факт-таблица)
```

### Доступ к данным
ORM/query-builder не используется — чистый SQL + repository pattern (Clean/DDD: `domain` / `application` / `infrastructure`). Разные сервисы используют разные драйверы поверх одного DSN: **auth** — `pgx/pgxpool`, **booking/branches/registration** — `database/sql` + `lib/pq`, **migrate** — `pgx/stdlib`.

DSN собирается в `libs/config/dbConfig.go`: либо берётся готовый `DB_DSN`, либо строится как `postgres://DB_USER:DB_PASSWORD@DB_HOST:DB_PORT/POSTGRES_DB?sslmode=DB_SSLMODE`.

---

## 7. Redis: как устроено общение

Клиент — `github.com/redis/go-redis/v9`, обёртка в `libs/redisclient/`. Каждый сервис (кроме `branches`, который Redis не использует) создаёт свой собственный `*redis.Client` при старте, но все ходят в **один и тот же инстанс и одну и ту же логическую БД (DB=0)** — разделение только по префиксам ключей. Конфиг: `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`.

Distributed lock и rate-limiting в коде не реализованы. Pub/Sub не используется — только строки/hash/sorted set.

| Сервис | Назначение | Ключи | TTL |
|---|---|---|---|
| **auth** | Refresh-сессии (whitelist) | `auth:refresh:<jti>` → userID | `REFRESH_TOKEN_TTL` (168h) |
| **registration** | Pending-регистрация | `registration:pending:<id>` → JSON | 10 мин |
| **registration** | Восстановление пароля | `registration:recovery:<id>` → JSON | 10 мин |
| **booking** | Отложенные уведомления (producer) | `notification:<id>` (payload) + Sorted Set `notifications:scheduled` (score=unix-время отправки) | без TTL у payload |
| **scheduler** | Consumer отложенных уведомлений | читает `notifications:scheduled` по `ZRangeByScore(-inf, now)`, `ZREM` после выборки | — |

Механика проверки кода регистрации в `registration` — атомарный **Lua-скрипт** (`GET` + сравнение кода + `DEL` при совпадении), защита от повторного использования и гонок.

Связка **booking (producer) → Redis Sorted Set → scheduler (poll-consumer, тикер)** — это самодельная delay-queue для напоминаний по email (за 24ч и за 1ч до визита). Это единственное подобие «очереди сообщений» в проекте.

**Важно:** сама предметная «очередь» (расписание записей, слоты) хранится в PostgreSQL, а не в Redis, несмотря на название проекта — Redis решает только вспомогательные задачи (сессии, TTL-состояния, delay-queue уведомлений).

---

## 8. Фронтенд (идейно)

- **React 19** + **Vite 8**, роутинг — `react-router-dom v7` (`BrowserRouter`).
- Серверное состояние/кэш — **@tanstack/react-query v5** (`src/libs/queryClient.js`).
- Клиентское состояние авторизации — React Context (`src/context/AuthContext.jsx`), без Redux/Zustand.
- Стили — **Tailwind CSS v4**, utility-классы прямо в JSX.
- Чистый JS/JSX (не TypeScript), типизация пропсов через `prop-types`.

### Структура `frontend/src/`
- `api/` — по axios-инстансу на каждый бэкенд-сервис (`registrationApi`, `authorizationApi`, `branchesApi`, `bookingApi`), все с `withCredentials: true` (авторизация — через HttpOnly cookie, не localStorage).
- `components/ui` — примитивы (Button, Input, Select, FieldError); `components/layouts` — каркас дашборда (Sidebar, DashboardTopBar); `components/dashboard` — карточки предметной области (BookingCard, ClientCard, EmployeeCard, ServiceCard, EmployeeSchedule и т.д.); `ProtectedRoute.jsx` — гейт по ролям.
- `pages/` — по странице на маршрут: LoginPage, RegistrationPage, VerifyPage, PublicBookingPage, AdminPage, ManagerPage, ClientsPage, ServicesPage, EmployeesPage, SettingsPage.
- `context/AuthContext.jsx`, `hooks/useFormValidation.js`, `constants/roles.js` + `schedule.js`, `mocks/*` (заглушки для локальной разработки UI).

### Общение с бэкендом
В `src/api/axios.js` реализован **single-flight auto-refresh**: на 401 (кроме самого login/refresh/logout) интерцептор вызывает `POST /auth/refresh` через общий shared-промис (чтобы не плодить параллельные рефреши), повторяет исходный запрос; при неудаче рефреша кидает DOM-событие `auth:unauthorized`, на которое подписан `AuthContext` и сбрасывает пользователя.

URL сервисов задаются через `VITE_REGISTRATION_URL`, `VITE_AUTH_URL`, `VITE_BRANCHES_URL`, `VITE_BOOKING_URL` (env), с дефолтами на `localhost:808{1,2,3,4}` — совпадают с портами из docker-compose, поэтому для локальной разработки против локального стека отдельный `frontend/.env` не требуется.

### Роли и маршруты
- Публично: `/register`, `/verify`, `/login`, `/public/:registrationSlug`.
- `/admin*` — `super_admin`/`business_admin`.
- `/manager*` — `manager`, с проверкой, что запрошенный `branchId` совпадает с `user.branch_id` (иначе редирект на свой филиал).
- Роль `employee` пока не имеет отдельных защищённых маршрутов (задел на будущее).

### Тесты
Cypress e2e (`frontend/cypress/e2e/`): `auth.cy.js` (регистрация→подтверждение→логин), `registration.cy.js` (флоу регистрации бизнеса, дубли email), `dashboard-pages.cy.js` (дашборд поверх демо-данных из миграций). Тесты обращаются к реальной тестовой БД напрямую через `cy.task` (не моки).

---

## 9. Docker Compose, порты, запуск

### Сервисы в корневом `docker-compose.yml`
Сеть — `online_queue_network`, единственный volume — `postgres_data`.

| Сервис | Порт (host:container) | Зависимости | Примечание |
|---|---|---|---|
| postgres | 5432:5432 | — | image `postgres:15`, healthcheck `pg_isready` |
| redis | 6379:6379 | — | image `redis:7`, `--requirepass`, healthcheck `redis-cli ping` |
| migrate | — | postgres (healthy) | одноразовый job (`restart: "no"`), прогоняет goose-миграции |
| auth | 8082:8082 | migrate (completed), redis (healthy) | |
| registration | 8081:8081 | migrate (completed), redis (healthy) | |
| branches | 8083:8083 | migrate (completed) | Redis не требует |
| booking | 8084:8084 | migrate (completed), redis (healthy) | |
| scheduler | — (не публикуется) | redis (healthy) | фоновый воркер |

Все Go-сервисы (auth/registration/branches/booking) имеют `GET /health`.

Сервис `services/notification` **не подключён** ни к одному docker-compose файлу — код есть, контейнера нет.

Observability (Prometheus, ELK, Jaeger) вынесены в отдельные compose-файлы `observability/{prometheus,elk,jaeger}/docker-compose.yml`, поднимаются отдельно поверх уже созданной сети `online_queue_network` — не входят в обязательный запуск.

### Как поднять проект локально
1. В корне уже лежит заполненный `.env` (не `.env.example`) с dev-значениями (Postgres, Redis, JWT-секреты, SMTP). При необходимости — заменить секреты на свои.
2. `docker compose up -d --build` — соберёт образы, поднимет postgres/redis, дождётся healthy, прогонит `migrate` (включая seed демо-данных), затем поднимет `auth/registration/branches/booking` и `scheduler`.
   - Повторно накатить миграции вручную: `docker compose run --rm migrate`.
3. Проверка: `docker compose ps`, `curl http://localhost:8081/health` (и т.д. по портам).
4. **Фронтенд поднимается отдельно** (в корневом compose контейнера для него нет):
   ```bash
   cd frontend
   npm install
   npm run dev   # http://localhost:5173 (Vite, strictPort)
   ```
   Дефолтные адреса API в `axios.js` (`localhost:8081-8084`) уже совпадают с портами из compose — доп. настройка не нужна. Для других окружений — `frontend/.env` с `VITE_*_URL`.
5. Остановка: `docker compose down` (данные Postgres сохраняются в volume; `-v` — для полной очистки).

### Dockerfile'ы
Все сервисы (кроме migrate) — multi-stage: `golang:1.26` builder → `debian:bookworm-slim` (booking/scheduler — с закреплёнными digest-версиями, non-root `USER 65532:65532`). `migration.Dockerfile` копирует бинарь **и** папку `migrations/`.

Makefile и отдельных стартовых shell-скриптов в проекте нет.

---

## 10. Известные технические долги (зафиксировано агентами при анализе)

- `employee_schedules_legacy` — таблица-сирота после рефакторинга (миграция №4), не используется приложением, кандидат на удаление.
- В registration/branches часть маршрутов не строго проверяет HTTP-метод.
- В некоторых местах branches ошибки forbidden/not-found маппятся в 500 вместо 403/404.
- `auth`/`branches` местами логируют `JWT_ACCESS_SECRET` — нужно убрать перед проду.
- `.env` в корне репозитория содержит реальные dev-секреты (включая SMTP app-password) — стоит проверить .gitignore.
- Опечатка в конфиге React Query: `refetchOnWindowsFocus` вместо `refetchOnWindowFocus` (не действует).
- `services/notification` — пустая нереализованная заглушка.

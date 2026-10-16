# Online Queue Management System

Сервис онлайн-записи к бизнесам сферы услуг. Клиент записывается по публичной ссылке без регистрации, бизнес ведет филиалы и записи в личном кабинете, система напоминает клиенту о визите по email и в Telegram.

Бэкенд - Go-микросервисы, фронтенд - React SPA, хранилища - PostgreSQL и Redis. Все запускается одной командой через Docker Compose.

## Требования

- Docker с Compose v2 (команда `docker compose`).
- Свободные порты на хосте: 5173, 8081, 8082, 8083, 8084, 8085.

Go и Node.js для запуска не нужны: все собирается внутри контейнеров.

## Быстрый запуск

```bash
cp .env.example .env
# отредактировать .env (см. раздел ниже), как минимум пароли, JWT-секреты и SMTP
docker compose up -d --build
```

Первая сборка занимает несколько минут. Миграции БД применяются автоматически: сервис `migrate` запускается перед остальными и завершается.

Проверить, что все поднялось:

```bash
docker compose ps
```

Все сервисы должны быть в состоянии `running` (или `healthy`), `migrate` - `exited (0)`.

## Адреса

| Что | Адрес |
|---|---|
| Публичная страница записи (демо-бизнес) | http://localhost:5173/public/demo-business |
| Вход в кабинет | http://localhost:5173/login |
| Регистрация бизнеса | http://localhost:5173/register |
| API registration | http://localhost:8081 |
| API auth | http://localhost:8082 |
| API branches | http://localhost:8083 |
| API booking | http://localhost:8084 |
| API bot | http://localhost:8085 |

Демо-пользователь (создается миграциями): `admin@test.com` / `123456`.

У каждого API есть `GET /health`.

## Файл .env

Все сервисы читают настройки только из переменных окружения. `docker compose` передает в контейнеры файл `.env` из корня проекта. Шаблон со всеми ключами - `.env.example`. Файл `.env` в git не попадает.

### Обязательно заполнить

| Переменная | Что это |
|---|---|
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | пользователь, пароль и имя БД, с которыми создается контейнер Postgres |
| `DB_USER`, `DB_PASSWORD` | с ними сервисы подключаются к БД; должны совпадать с `POSTGRES_USER` и `POSTGRES_PASSWORD` |
| `DB_HOST`, `DB_PORT` | адрес БД; в compose это `postgres` и `5432`, менять не нужно |
| `DB_SSLMODE` | режим TLS к БД, значения по умолчанию нет; для локального запуска `disable` |
| `REDIS_PASSWORD` | пароль Redis; задается и контейнеру Redis, и сервисам |
| `REDIS_ADDR` | адрес Redis; в compose это `redis:6379` |
| `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET` | секреты подписи токенов; длинные случайные строки, разные |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS` | почтовый сервер для кодов регистрации и напоминаний |

Сгенерировать секрет можно так: `openssl rand -hex 32`.

Для Gmail: `SMTP_HOST=smtp.gmail.com`, `SMTP_PORT=587`, `SMTP_USER` - адрес почты, `SMTP_PASS` - пароль приложения (Google Account -> Security -> App passwords), а не пароль от аккаунта. Без рабочего SMTP сервисы стартуют, но письма не уходят и повторяются в Redis.

### Можно оставить как в шаблоне

| Переменная | Что это |
|---|---|
| `REGISTRATION_PORT`, `AUTH_PORT`, `BRANCHES_PORT`, `BOOKING_PORT`, `BOT_PORT` | порты сервисов внутри контейнеров; если менять, то и в `ports` в `docker-compose.yml` |
| `REGISTRATION_URL`, `AUTH_URL`, `BRANCHES_URL`, `BOOKING_URL` | адреса API, по которым ходит браузер; фронтенд получает их при старте контейнера |
| `CORS_ALLOWED_ORIGINS` | с каких адресов браузеру можно обращаться к API (через запятую) |
| `FRONTEND_URL` | куда вернуть пользователя после входа в Google |
| `ACCESS_TOKEN_TTL`, `REFRESH_TOKEN_TTL` | время жизни токенов (15m и 168h) |
| `COOKIE_SECURE` | `true`, только если сайт открыт по HTTPS |
| `EMAIL_TIMEOUT`, `NUM_WORKERS`, `RATE_LIMIT`, `WRK_TIMEOUT`, `EMAIL_RETRY_AFTER`, `EMAIL_MAX_ATTEMPTS` | очередь писем: число воркеров, таймауты, повторы |
| `APP_ENV` | пусто; значение `test` делает код регистрации всегда `123456` и отключает отправку письма с кодом |
| `BOT_URL` | адрес сервиса bot для scheduler; в compose это `http://bot:8085` |
| `SCHEDULER_*` | интервал опроса, размер пачки, аренда задач и повторы напоминаний |
| `TELEGRAM_POLLING_TIMEOUT`, `TELEGRAM_POLLING_INTERVAL` | параметры опроса Telegram |

### Необязательные интеграции

- **Telegram.** `TELEGRAM_BOT_TOKEN` - токен бота от @BotFather. Если пусто, сервис bot стартует, но ничего не отправляет. Напоминание в Telegram придет, только если клиент сам написал боту и отправил телефон или username.
- **Google Календарь.** `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL`. Если не заданы, кнопка автоматического сохранения в Google скрыта, а скачивание `.ics` и ссылка-шаблон Google Календаря работают. Как получить ключи: `docs/PUBLIC_BOOKING_API.md`, раздел "Как включить Google OAuth".

## Полезные команды

```bash
docker compose logs -f booking            # логи сервиса (JSON в stdout)
docker compose up -d --build booking      # пересобрать и перезапустить один сервис
docker compose run --rm migrate           # применить миграции вручную
docker compose exec postgres psql -U admin -d online_queue_db
docker compose exec redis redis-cli       # пароль берется из REDISCLI_AUTH
docker compose up -d --scale scheduler=2 --scale bot=2   # несколько реплик воркеров
docker compose down                       # остановить
docker compose down -v                    # остановить и удалить данные БД
```

Если менялись `POSTGRES_*` после первого запуска или БД в странном состоянии, выполните `docker compose down -v` и запустите заново: Postgres создает пользователя и базу только при пустом томе.

## Разработка фронтенда

Фронтенд с горячей перезагрузкой (нужен Node.js 22):

```bash
docker compose stop frontend
cd frontend
npm ci
npm run dev        # http://localhost:5173
```

## Тесты

```bash
go test ./...
```

Юнит-тесты не требуют запущенного стека: Redis в них заменен на `miniredis`.

## Структура

```
services/        Go-сервисы: registration, auth, branches, booking, bot, scheduler, migrate
libs/            общие пакеты: конфиг, логгер, JWT, HTTP-middleware, email-очередь, Redis
migrations/      SQL-миграции (goose) и демо-данные
frontend/        React SPA и nginx-образ для нее
docs/            описание API и обзор проекта
Отчёт.md         отчет: доменная область, стек, сущности, 12 факторов
```



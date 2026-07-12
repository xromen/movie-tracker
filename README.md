# Movie Tracker

Монорепозиторий с фронтендом на Next.js и backend API на Go. Приложение ищет фильмы и сериалы через TMDB-совместимое API, показывает каталоги, детали, трейлеры, рекомендации и хранит пользовательский список просмотра.

## Состав проекта

```text
apps/
  web/  Next.js 16 App Router приложение
  api/  Go API, миграции, интеграции с PostgreSQL, Redis и TMDB
compose.yaml
.github/workflows/ci-cd.yml
```

## Быстрый запуск через Docker Compose

1. Создайте корневой `.env` из `.env.example`.
2. Заполните как минимум `TMDB_BEARER_TOKEN`, `JWT_SECRET` и `DB_PASSWORD`.
3. Убедитесь, что внешний volume `movie-tracker-go_postgres_data` существует, либо создайте его:

```bash
docker volume create movie-tracker-go_postgres_data
```

4. Запустите стек:

```bash
docker compose up --build
```

По умолчанию:

- web: `http://localhost:3000`
- API: `http://localhost:8080/api`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`

Порты на хосте переопределяются переменными `WEB_PORT`, `API_PORT`, `POSTGRES_PORT` и `REDIS_PORT`.

## Переменные окружения

Корневой `.env` используется Docker Compose. Основные значения:

```env
WEB_PORT=3000
API_PORT=8080
POSTGRES_PORT=5432
REDIS_PORT=6379

NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000

PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=change-me
DB_NAME=movietracker
DB_SSLMODE=disable
DB_STATEMENT_TIMEOUT=15s

TMDB_BASE_URL=https://api.themoviedb.org/3
TMDB_IMAGES_BASE_URL=https://api.themoviedb.org
TMDB_BEARER_TOKEN=your_tmdb_v4_bearer_token
TMDB_TIMEOUT=10s

JWT_SECRET=change-me-in-production

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_DISABLED=false
```

В Docker Compose frontend ходит к backend по внутреннему адресу `http://api:${PORT:-8080}/api`. При локальном запуске web вне Docker значение по умолчанию для `MOVIE_TRACKER_API_URL` - `http://localhost:8080/api`.

## Локальная разработка web

```bash
cd apps/web
npm install
npm run dev
```

При необходимости создайте `apps/web/.env.local`:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

## Локальная разработка API

Для инфраструктуры можно поднять PostgreSQL и Redis из корневого compose:

```bash
docker compose up -d postgres redis
```

API автоматически применяет миграции при старте. В Docker-образе миграции лежат в `/migrations`; текущий локальный `go run` также ищет `file:///migrations`, поэтому для полностью локального запуска нужно подготовить этот путь к `apps/api/migrations` или запускать API через Docker Compose.

```bash
cd apps/api
go run ./cmd/api/
```

## Проверки

Frontend:

```bash
cd apps/web
npm run lint
npm run typecheck
npm run build
```

API:

```bash
cd apps/api
go test ./...
```

Compose build:

```bash
docker compose build
```

## CI/CD

GitHub Actions запускает frontend lint/typecheck/build, проверяет сборку Docker Compose и на push в `main`/`master` деплоит стек на self-hosted runner через `docker compose up --build --detach --remove-orphans`.

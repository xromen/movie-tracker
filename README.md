# Movie Tracker

Монорепозиторий с фронтендом на Next.js и backend API на Go.

## Структура

```text
apps/
  web/  Next.js приложение
  api/  Go API, миграции, интеграции с PostgreSQL/Redis/TMDB
compose.yaml
.github/workflows/ci-cd.yml
```

## Запуск через Docker Compose

1. Создайте корневой `.env` из `.env.example` и заполните `TMDB_BEARER_TOKEN`, `JWT_SECRET`, пароли БД при необходимости.
2. Запустите весь стек:

```bash
docker compose up --build
```

Фронтенд будет доступен на `http://localhost:3000`, API на `http://localhost:8080`.

Внутри Docker фронтенд обращается к API по имени сервиса:

```env
MOVIE_TRACKER_API_URL=http://api:8080/api
```

Снаружи Docker локальная разработка фронта по-прежнему использует API на:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
```

## Локальная разработка без Docker для фронта

```bash
cd apps/web
npm install
npm run dev
```

Если API запущен локально на другом адресе, задайте в `apps/web/.env.local`:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

## Локальная разработка API

```bash
cd apps/api
go run ./cmd/api/
```

Для зависимостей можно поднять только инфраструктуру из корневого compose:

```bash
docker compose up -d postgres redis
```

## Проверки

```bash
cd apps/web
npm run lint
npm run typecheck
npm run build
```

```bash
cd apps/api
go test ./...
```

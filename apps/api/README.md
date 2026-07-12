# MovieTracker API

MovieTracker API - backend-сервис на Go для поиска фильмов через TMDB-совместимое API и ведения пользовательских списков фильмов. Проект использует Gin для HTTP API, PostgreSQL для хранения пользователей и пользовательских списков, Redis для кэширования и JWT для авторизации.

## Возможности

- Регистрация и авторизация пользователей.
- JWT-аутентификация через заголовок `Authorization: Bearer <token>`.
- Поиск фильмов по названию.
- Получение подборок фильмов: популярные, топ рейтинга, сейчас в прокате и ожидаемые.
- Получение детальной информации о фильме, включая жанры, видео, рейтинг TMDB и принадлежность к коллекции.
- Получение рекомендаций по фильму.
- Получение информации о коллекции фильмов.
- Добавление фильмов в пользовательский список.
- Фильтрация пользовательского списка по статусу просмотра.
- Health-check эндпоинты для liveness/readiness.
- Кэширование ответов TMDB и пользовательских списков в Redis.

## Технологии

- Go 1.26.3
- Gin
- PostgreSQL 16
- Redis 7
- JWT
- Docker / Docker Compose
- golang-migrate

## Переменные окружения

Приложение читает переменные окружения из `.env` при локальном запуске.

```env
PORT=8080

DB_HOST=localhost
DOCKER_DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=movietracker
DB_SSLMODE=disable

TMDB_BASE_URL=https://api.themoviedb.org/3
TMDB_BEARER_TOKEN=your_tmdb_v4_bearer_token

JWT_SECRET=change-me-in-production

REDIS_ADDR=localhost:6379
DOCKER_REDIS_ADDR=redis:6379
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_DISABLED=false
```

Важно: в коде HTTP-порт читается из переменной `PORT`. Если в `.env` указать только `HTTP_PORT`, приложение возьмет значение по умолчанию `8080`.

### Описание переменных

| Переменная | Описание | Значение по умолчанию |
| --- | --- | --- |
| `PORT` | Порт HTTP-сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DOCKER_DB_HOST` | Хост PostgreSQL внутри Docker Compose сети | `postgres` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь PostgreSQL | `postgres` |
| `DB_PASSWORD` | Пароль PostgreSQL | пусто |
| `DB_NAME` | Имя базы данных | `movietracker` |
| `DB_SSLMODE` | SSL-режим подключения к PostgreSQL | `disable` |
| `TMDB_BASE_URL` | Base URL TMDB-совместимого API | `https://api.themoviedb.org/3` |
| `TMDB_BEARER_TOKEN` | Bearer token TMDB API v4 | пусто |
| `JWT_SECRET` | Секрет для подписи JWT | `change-me-in-production` |
| `REDIS_ADDR` | Адрес Redis | `localhost:6379` |
| `DOCKER_REDIS_ADDR` | Адрес Redis внутри Docker Compose сети | `redis:6379` |
| `REDIS_PORT` | Порт Redis, пробрасываемый на хост | `6379` |
| `REDIS_PASSWORD` | Пароль Redis | пусто |
| `REDIS_DB` | Номер базы Redis | `0` |
| `REDIS_DISABLED` | Отключить Redis-кэш | `false` |

## Запуск

### 1. Запуск инфраструктуры через Docker Compose

```bash
docker compose up -d postgres redis
```

### 2. Настройка `.env`

Создайте файл `.env` в корне проекта и заполните переменные окружения. Можно взять `.env.example` за основу.

Для локального запуска приложения вне Docker используйте:

```env
DB_HOST=localhost
REDIS_ADDR=localhost:6379
```

Для запуска API внутри Docker Compose используются отдельные docker-адреса сервисов:

```env
DOCKER_DB_HOST=postgres
DOCKER_REDIS_ADDR=redis:6379
```

### 3. Применение миграций

Проект использует SQL-миграции из директории `migrations`.

Установите `golang-migrate` CLI с поддержкой PostgreSQL:

```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Затем примените миграции:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/movietracker?sslmode=disable" up
```

Проверьте строку подключения: пользователь, пароль и база должны совпадать со значениями из вашего `.env` / `docker-compose.yml`.

### 4. Локальный запуск API

```bash
go run ./cmd/api/
```

После запуска API будет доступен по адресу:

```text
http://localhost:8080
```

### 5. Запуск всего стека

```bash
docker compose up --build
```

API будет доступен на `http://localhost:8080`.

## Эндпоинты

### Health

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `GET` | `/health/live` | Нет | Проверка, что процесс API запущен |
| `GET` | `/health/ready` | Нет | Проверка готовности API, PostgreSQL и Redis |

### Auth

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/register` | Нет | Регистрация пользователя |
| `POST` | `/api/v1/auth/login` | Нет | Авторизация пользователя |

#### Регистрация

```http
POST /api/v1/auth/register
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "username": "movie_fan",
  "password": "password123"
}
```

Ответ:

```json
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "username": "movie_fan"
  },
  "token": "jwt-token"
}
```

#### Логин

```http
POST /api/v1/auth/login
Content-Type: application/json
```

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

### Movies

| Метод | Путь | Авторизация | Query-параметры | Описание |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/movies/search` | Нет | `q`, `page` | Поиск фильмов |
| `GET` | `/api/v1/movies/now-playing` | Нет | `page` | Фильмы, которые сейчас в прокате |
| `GET` | `/api/v1/movies/popular` | Нет | `page` | Популярные фильмы |
| `GET` | `/api/v1/movies/top-rated` | Нет | `page` | Фильмы с высоким рейтингом |
| `GET` | `/api/v1/movies/upcoming` | Нет | `page` | Ожидаемые фильмы |
| `GET` | `/api/v1/movies/{id}` | Опционально | - | Детальная информация о фильме |
| `GET` | `/api/v1/movies/{id}/recommendations` | Нет | `page` | Рекомендации по фильму |
| `GET` | `/api/v1/movies` | Да | `status`, `page`, `per_page` | Список фильмов пользователя |
| `POST` | `/api/v1/movies/list` | Да | - | Добавить фильм в список пользователя |

Публичные списки фильмов возвращают ответ вида:

```json
{
  "movies": [
    {
      "id": 550,
      "title": "Бойцовский клуб",
      "overview": "...",
      "poster_path": "http://83.142.30.2:8002/image/t/p/w500/...",
      "release_date": "1999-10-15"
    }
  ],
  "total_pages": 10,
  "total_items": 200
}
```

#### Добавление фильма в список

```http
POST /api/v1/movies/list
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "id": 550,
  "status": "watched",
  "rating": 10
}
```

Допустимые значения `status`:

- `watched`
- `want_to_watch`
- `favorite`

`rating` опционален. На уровне БД рейтинг ограничен диапазоном от `1` до `10`.

#### Получение пользовательского списка

```http
GET /api/v1/movies?status=watched&page=1&per_page=20
Authorization: Bearer <token>
```

Параметр `status` опционален. Если он указан, значение должно быть одним из: `watched`, `want_to_watch`, `favorite`.

### Collections

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `GET` | `/api/v1/collections/{id}` | Нет | Детальная информация о коллекции фильмов |

Пример ответа:

```json
{
  "id": 10,
  "name": "Коллекция фильмов",
  "overview": "...",
  "poster_path": "http://83.142.30.2:8002/image/t/p/w500/...",
  "parts": [
    {
      "id": 11,
      "title": "Фильм",
      "overview": "...",
      "poster_path": "http://83.142.30.2:8002/image/t/p/w500/...",
      "media_type": "movie",
      "release_date": "2020-01-01"
    }
  ]
}
```

## Авторизация

Защищенные эндпоинты требуют JWT:

```http
Authorization: Bearer <token>
```

Токен возвращается при регистрации и логине. Время жизни access token в текущей конфигурации - 24 часа.

## Кэширование

Redis используется для кэширования:

- результатов поиска и публичных подборок фильмов на 24 часа;
- детальной информации о фильме на 24 часа;
- рекомендаций на 24 часа;
- пользовательских списков на 5 минут.

Для локальной разработки Redis можно отключить:

```env
REDIS_DISABLED=true
```

## Тесты и качество кода

Запуск тестов:

```bash
go test -race -count=1 ./...
```

Запуск тестов с coverage:

```bash
go test -race -count=1 -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Проверки линтера и `go vet`:

```bash
staticcheck ./...
go vet ./...
```

## Структура проекта

```text
cmd/api/                 Точка входа приложения
internal/config/         Загрузка конфигурации
internal/domain/         Доменные модели и ошибки
internal/handler/        HTTP handlers и middleware
internal/platform/       Интеграции: PostgreSQL, Redis, JWT, TMDB, hasher
internal/repository/     Работа с PostgreSQL
internal/service/        Бизнес-логика
migrations/              SQL-миграции
```

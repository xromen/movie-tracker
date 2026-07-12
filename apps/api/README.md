# MovieTracker API

Backend-сервис на Go для Movie Tracker. API работает с TMDB-совместимым источником данных, хранит пользователей, роли, refresh tokens, медиатеку пользователя и просмотренные эпизоды в PostgreSQL, а ответы TMDB кэширует в Redis.

## Возможности

- регистрация, вход, выход и обновление сессии;
- cookie-based auth через `access_token` и `refresh_token`;
- проверка `auth_version` пользователя, чтобы инвалидировать старые access tokens;
- поиск фильмов, сериалов и общий multi-search;
- публичные подборки фильмов: now playing, popular, top rated, upcoming;
- публичные подборки сериалов: airing today, popular, top rated, on the air;
- детали фильма или сериала с жанрами, видео, рейтингами, странами производства и датами релиза;
- сезоны и эпизоды сериалов, включая отметки просмотра для авторизованного пользователя;
- рекомендации по фильму или сериалу;
- коллекции фильмов;
- личный watch-list для фильмов и сериалов со статусами `watched`, `want_to_watch`, `favorite`;
- health-check эндпоинты для liveness/readiness;
- автоматическое применение SQL-миграций при старте.

## Технологии

- Go 1.26.3
- Gin
- PostgreSQL 16
- Redis 7
- JWT
- bcrypt
- golang-migrate
- Docker / Docker Compose

## Конфигурация

При локальном запуске приложение читает `.env` из текущей директории. В Docker Compose корневой `.env` передается в сервис `api`, а часть значений переопределяется через `compose.yaml`.

```env
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

### Переменные

| Переменная | Описание | Значение по умолчанию |
| --- | --- | --- |
| `PORT` | Порт HTTP-сервера внутри контейнера или локального процесса | `8080` |
| `DB_HOST` | Хост PostgreSQL | `localhost` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `DB_USER` | Пользователь PostgreSQL | `postgres` |
| `DB_PASSWORD` | Пароль PostgreSQL | пусто |
| `DB_NAME` | Имя базы данных | `movietracker` |
| `DB_SSLMODE` | SSL-режим подключения к PostgreSQL | `disable` |
| `DB_STATEMENT_TIMEOUT` | Таймаут SQL statement для pgx pool | `15s` |
| `TMDB_BASE_URL` | Base URL TMDB API | `https://api.themoviedb.org/3` |
| `TMDB_IMAGES_BASE_URL` | Base URL для TMDB-изображений | `https://api.themoviedb.org` |
| `TMDB_BEARER_TOKEN` | Bearer token TMDB API v4 | пусто |
| `TMDB_TIMEOUT` | Таймаут запросов к TMDB | `10s` |
| `JWT_SECRET` | Секрет для подписи access token | `change-me-in-production` |
| `REDIS_ADDR` | Адрес Redis | `localhost:6379` |
| `REDIS_PASSWORD` | Пароль Redis | пусто |
| `REDIS_DB` | Номер базы Redis | `0` |
| `REDIS_DISABLED` | Отключить Redis-кэш | `false` |

`apps/api/.env.example` также содержит старые подсказки `DOCKER_DB_HOST` и `DOCKER_REDIS_ADDR`, но текущий код читает именно `DB_HOST` и `REDIS_ADDR`. В Docker Compose они выставляются как `postgres` и `redis:6379`.

## Запуск

### Через Docker Compose

Из корня репозитория:

```bash
docker compose up --build api postgres redis
```

API будет доступен на `http://localhost:8080/api`, если `API_PORT` не переопределен. При старте сервис подключается к PostgreSQL, применяет миграции из `/migrations`, подключает Redis и запускает HTTP-сервер.

### Локально

1. Поднимите зависимости:

```bash
docker compose up -d postgres redis
```

2. Настройте `.env` в `apps/api` или экспортируйте переменные окружения.

3. Учитывайте, что текущий код миграций ищет `file:///migrations`. Для локального `go run` подготовьте этот путь к каталогу `apps/api/migrations` или запускайте API в контейнере.

```bash
cd apps/api
go run ./cmd/api/
```

## Эндпоинты

Все бизнес-эндпоинты находятся под `/api/v1`. Health-check находится под `/api/health`.

### Health

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `GET` | `/api/health/live` | нет | Проверка, что процесс API запущен |
| `GET` | `/api/health/ready` | роль `admin` | Проверка готовности API, PostgreSQL и Redis |

### Auth

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/register` | нет | Регистрация пользователя |
| `POST` | `/api/v1/auth/login` | нет | Вход пользователя |
| `POST` | `/api/v1/auth/logout` | нет | Очистка auth cookies |
| `POST` | `/api/v1/auth/refresh` | refresh cookie | Обновление access/refresh cookies |

Регистрация:

```json
{
  "email": "user@example.com",
  "username": "movie_fan",
  "password": "password123"
}
```

Логин:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

При успешной регистрации, логине или refresh API устанавливает cookies `access_token` и `refresh_token`. Тело ответа пустое.

### Movies

| Метод | Путь | Авторизация | Query-параметры | Описание |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/movie/search` | опционально | `q`, `page` | Поиск фильмов |
| `GET` | `/api/v1/movie/now-playing` | опционально | `page` | Фильмы в прокате |
| `GET` | `/api/v1/movie/popular` | опционально | `page` | Популярные фильмы |
| `GET` | `/api/v1/movie/top-rated` | опционально | `page` | Фильмы с высоким рейтингом |
| `GET` | `/api/v1/movie/upcoming` | опционально | `page` | Ожидаемые фильмы |
| `GET` | `/api/v1/movie/{id}` | опционально | нет | Детали фильма |
| `GET` | `/api/v1/movie/{id}/recommendations` | опционально | `page` | Рекомендации по фильму |

Ответ списков:

```json
{
  "results": [
    {
      "id": 550,
      "title": "Fight Club",
      "overview": "...",
      "poster_path": "https://api.themoviedb.org/image/t/p/w500/...",
      "release_date": "1999-10-15",
      "vote_average": 8.4,
      "vote_count": 30000,
      "watch_status": "watched"
    }
  ],
  "total_pages": 10,
  "total_items": 200
}
```

### TV

| Метод | Путь | Авторизация | Query-параметры | Описание |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/tv/search` | опционально | `q`, `page` | Поиск сериалов |
| `GET` | `/api/v1/tv/airing-today` | опционально | `page` | Серии, выходящие сегодня |
| `GET` | `/api/v1/tv/popular` | опционально | `page` | Популярные сериалы |
| `GET` | `/api/v1/tv/top-rated` | опционально | `page` | Сериалы с высоким рейтингом |
| `GET` | `/api/v1/tv/on-the-air` | опционально | `page` | Сериалы в эфире |
| `GET` | `/api/v1/tv/{id}` | опционально | нет | Детали сериала |
| `GET` | `/api/v1/tv/{id}/season/{season_number}` | опционально | `page` | Эпизоды сезона |
| `GET` | `/api/v1/tv/{id}/recommendations` | опционально | `page` | Рекомендации по сериалу |
| `PUT` | `/api/v1/tv/{id}/season/{season_number}/watched` | да | нет | Отметить сезон просмотренным |
| `DELETE` | `/api/v1/tv/{id}/season/{season_number}/watched` | да | нет | Снять отметку просмотра сезона |
| `PUT` | `/api/v1/tv/{id}/season/{season_number}/episode/{episode_number}/watched` | да | нет | Отметить эпизод просмотренным |
| `DELETE` | `/api/v1/tv/{id}/season/{season_number}/episode/{episode_number}/watched` | да | нет | Снять отметку просмотра эпизода |

### Watch List

| Метод | Путь | Авторизация | Query/body | Описание |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/watch-list` | да | `status`, `media_type`, `page`, `per_page` | Личный список пользователя |
| `GET` | `/api/v1/watch-list/status` | да | `media_id` | Статус конкретного media |
| `POST` | `/api/v1/watch-list/status` | да | JSON body | Установить статус |
| `DELETE` | `/api/v1/watch-list/status` | да | `media_id` | Удалить статус |

Запрос установки статуса:

```json
{
  "id": 550,
  "media_type": "movie",
  "watch_status": "watched"
}
```

Допустимые `media_type`: `movie`, `tv`.

Допустимые `watch_status`: `watched`, `want_to_watch`, `favorite`.

### Collections

| Метод | Путь | Авторизация | Описание |
| --- | --- | --- | --- |
| `GET` | `/api/v1/collections/{id}` | нет | Детали коллекции фильмов |

### Search

| Метод | Путь | Авторизация | Query-параметры | Описание |
| --- | --- | --- | --- | --- |
| `GET` | `/api/v1/search/multi` | нет | `query`, `page` | Общий поиск фильмов и сериалов |

## Авторизация

Основной сценарий авторизации - HTTP-only cookies. Middleware читает `access_token`, проверяет подпись JWT и сверяет `auth_version` пользователя в базе. Если access token истек, frontend proxy вызывает `/api/v1/auth/refresh` и прокидывает обновленные cookies.

Текущие TTL:

- access token: 15 секунд;
- refresh token: 7 дней.

## Кэширование

Redis используется для кэширования:

- поиска и публичных подборок фильмов;
- поиска и публичных подборок сериалов;
- деталей фильмов и сериалов;
- рекомендаций;
- коллекций;
- эпизодов сезона.

Для локальной разработки Redis можно отключить:

```env
REDIS_DISABLED=true
```

## Тесты и качество кода

```bash
go test ./...
go test -race -count=1 ./...
go test -race -count=1 -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go vet ./...
```

`staticcheck ./...` можно запускать дополнительно, если CLI установлен локально.

## Структура

```text
cmd/api/                 точка входа приложения
internal/config/         загрузка конфигурации
internal/domain/         доменные модели и ошибки
internal/handler/        HTTP handlers и middleware
internal/platform/       PostgreSQL, Redis, JWT, TMDB, hasher, refresh token
internal/repository/     работа с PostgreSQL
internal/service/        бизнес-логика
migrations/              SQL-миграции
```

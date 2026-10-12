# HTTP API

## Адреса и общий контракт

- прямой локальный адрес: `http://localhost:8080/api`;
- browser-facing proxy: `http://localhost:3000/api/backend`;
- бизнес-префикс: `/api/v1`;
- ошибка: JSON `{"error":"message"}`;
- Go API назначает каждому запросу `X-Request-ID` и пишет его в structured log. Next BFF сейчас не прокидывает этот response header браузеру.

Публичные movie/TV/company/collection endpoints используют optional auth: без валидной access cookie ответ остаётся публичным, а с валидной cookie включает пользовательский `watch_status` там, где он предусмотрен.

## Авторизация

| Метод | Путь | Body / условие | Результат |
| --- | --- | --- | --- |
| `POST` | `/api/v1/auth/register` | `email`, `username`, `password` | `201`, auth cookies |
| `POST` | `/api/v1/auth/login` | `email`, `password` | `200`, auth cookies |
| `POST` | `/api/v1/auth/refresh` | refresh cookie | `200`, rotated auth cookies |
| `POST` | `/api/v1/auth/logout` | — | `200`, cookies удалены |

Cookies называются `access_token` и `refresh_token`, имеют `HttpOnly`, `SameSite=Lax`, path `/`. Access token — JWT на 15 минут; refresh token — случайный токен на 7 дней, в базе хранится hash. Refresh rotation использует token family и отзывает скомпрометированное семейство при повторном использовании старого токена. `auth_version` пользователя позволяет инвалидировать access tokens.

Защищённые endpoints отвечают `401`, если access token отсутствует/истёк или версия auth не совпадает. Readiness требует роль `admin` и может вернуть `403`.

## Health и metrics

| Метод | Путь | Auth | Назначение |
| --- | --- | --- | --- |
| `GET` | `/api/health/live` | нет | только состояние HTTP-процесса |
| `GET` | `/api/health/ready` | `admin` | PostgreSQL, Redis, pool stats, uptime |
| `GET` | `/metrics` | нет | Prometheus metrics, путь вне `/api` |

Readiness возвращает `200` со статусом `ok` или `503` со статусом `degraded`.

## Movies

| Метод | Путь | Query | Auth |
| --- | --- | --- | --- |
| `GET` | `/api/v1/movie/search` | `q`, `page=1` | optional |
| `GET` | `/api/v1/movie/now-playing` | `page=1` | optional |
| `GET` | `/api/v1/movie/popular` | `page=1` | optional |
| `GET` | `/api/v1/movie/top-rated` | `page=1` | optional |
| `GET` | `/api/v1/movie/upcoming` | `page=1` | optional |
| `GET` | `/api/v1/movie/{id}` | — | optional |
| `GET` | `/api/v1/movie/{id}/recommendations` | `page=1` | optional |

## TV

| Метод | Путь | Query | Auth |
| --- | --- | --- | --- |
| `GET` | `/api/v1/tv/search` | `q`, `page=1` | optional |
| `GET` | `/api/v1/tv/airing-today` | `page=1` | optional |
| `GET` | `/api/v1/tv/popular` | `page=1` | optional |
| `GET` | `/api/v1/tv/top-rated` | `page=1` | optional |
| `GET` | `/api/v1/tv/on-the-air` | `page=1` | optional |
| `GET` | `/api/v1/tv/{id}` | — | optional |
| `GET` | `/api/v1/tv/{id}/recommendations` | `page=1` | optional |
| `GET` | `/api/v1/tv/{id}/season/{season}` | `page=1` | optional |
| `PUT`, `DELETE` | `/api/v1/tv/{id}/season/{season}/watched` | — | required |
| `PUT`, `DELETE` | `/api/v1/tv/{id}/season/{season}/episode/{episode}/watched` | — | required |

Season-level `PUT`/`DELETE` отвечают `204`; episode-level операции возвращают состояния эпизода и сезона.

## Watch list

Все endpoints требуют авторизацию.

| Метод | Путь | Параметры |
| --- | --- | --- |
| `GET` | `/api/v1/watch-list` | `status`, `media_type`, `page=1`, `per_page=20` |
| `GET` | `/api/v1/watch-list/status` | `media_id`, `media_type` |
| `POST` | `/api/v1/watch-list/status` | JSON body |
| `DELETE` | `/api/v1/watch-list/status` | `media_id`, `media_type` |

```json
{
  "id": 550,
  "media_type": "movie",
  "watch_status": "watched"
}
```

`media_type`: `movie` или `tv`. `watch_status`: `watched`, `want_to_watch` или `favorite`.

## Search, companies, collections и Telegram

| Метод | Путь | Query / auth | Назначение |
| --- | --- | --- | --- |
| `GET` | `/api/v1/search/multi` | `query`, `page=1` | общий поиск movie/TV |
| `GET` | `/api/v1/company/{company_id}` | optional auth | данные компании |
| `GET` | `/api/v1/company/{company_id}/movie` | `page=1`, optional auth | фильмы компании |
| `GET` | `/api/v1/company/{company_id}/tv` | `page=1`, optional auth | сериалы компании |
| `GET` | `/api/v1/collections/{id}` | optional auth | коллекция и её части |
| `GET` | `/api/v1/telegram/binding-url` | required auth | deep link, created/expiry timestamps |

## Формы списочных ответов

Movie/TV/search endpoints возвращают `results`, `total_pages`, `total_items`. Watch list возвращает `medias`, `total`, `page`, `per_page`. DTO определены рядом с handler: `apps/api/internal/handler/api/*.go`; frontend-типы — в `apps/web/src/lib/api/types.ts`.

## Запросы к TMDB

Общий HTTP transport повторяет GET без тела при сетевых timeout, EOF/обрыве соединения и HTTP `429`, `500`, `502`, `503`, `504`. Всего допускается 3 попытки. Паузы перед повторами — 250 и 500 мс со случайной добавкой до 50%; `Retry-After` для `429`/`503` учитывается как минимальное ожидание, поддерживаются секунды и HTTP-дата. Если пауза не помещается в deadline, возвращается последний ответ/ошибка без повтора.

`TMDB_TIMEOUT` ограничивает один исходящий HTTP-запрос вместе с ретраями, ожиданием лимитера и чтением ответа; более ранний deadline или отмена вызывающего контекста также останавливают запрос. Отмена контекста, исчерпание deadline и остальные HTTP-статусы не повторяются. Каждая попытка расходует токен общего для TMDB-клиента лимитера. Кэш и формат публичных API-ответов не изменены.

## Кэш

TMDB search, catalogs, details, recommendations, seasons, companies и collections кэшируются в Redis на 24 часа. Запись выполняется в фоне с timeout 3 секунды. Redis можно отключить: реализация превращается в no-op cache. Ключи централизованы в `internal/platform/cache/keys.go`.

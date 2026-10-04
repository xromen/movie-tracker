# Архитектура и карта кода

## Компоненты

| Компонент | Точка входа | Ответственность |
| --- | --- | --- |
| Next.js frontend | `apps/web/src/app` | UI, SSR, SEO, защита страниц, proxy к API |
| Go API | `apps/api/cmd/api/main.go` | HTTP API, auth, TMDB, PostgreSQL, Redis, миграции, метрики |
| Schedule/report worker | `apps/api/cmd/worker/main.go` | синхронизация дат релизов и эпизодов, формирование отчётов |
| Telegram bot/worker | `apps/api/cmd/bot/main.go` | привязка аккаунта, команды, отправка report messages |
| PostgreSQL | `apps/api/migrations` | постоянные данные и координация worker через advisory lock |
| Redis | `internal/platform/cache` | best-effort кэш данных TMDB |
| TMDB | `internal/platform/tmdb` | внешний каталог фильмов, сериалов, компаний и коллекций |

## Основные потоки

### Запрос из браузера

1. UI вызывает `fetchApi` из `apps/web/src/lib/api/client.ts`.
2. Запрос идёт на Next route `/api/backend/...`.
3. Route handler прокидывает method, query, body и cookies в Go API.
4. При наличии refresh cookie и отсутствии access cookie frontend один раз обновляет пару токенов; параллельные refresh объединяет `refresh-single-flight.ts`.
5. Go handler вызывает service, service — repository или TMDB client, после чего ответ и `Set-Cookie` возвращаются через proxy.

Серверные запросы Next также идут через собственный proxy. Запрос с cookies всегда `no-store`; публичный server-side GET по умолчанию имеет revalidation 300 секунд, если вызывающий код не указал другое значение.

### Сохранение media в watch list

`handler/api/watchlist.go` → `service/watchlist.go` → TMDB detail при необходимости → upsert `medias` → upsert `user_medias`. Внешний `tmdb_id` и `media_type` однозначно определяют media, но `user_medias.media_id` хранит внутренний `medias.id`.

### Фоновая синхронизация

`service/worker.go` запускается сразу и затем раз в час. PostgreSQL advisory lock `tmdb-schedule-sync` не даёт нескольким экземплярам выполнять цикл одновременно. Пакеты обрабатываются по 100 записей:

- коллекции обновляются раз в 7 дней;
- актуальные фильмы и сериалы обычно обновляются раз в сутки;
- завершённые/canceled сериалы и фильмы со старыми релизами — раз в 30 дней;
- ошибка откладывает повтор на 1 час;
- для пользователей с Telegram формируются периодические отчёты и строки `report_messages`.

Telegram worker опрашивает готовые сообщения, отправляет их как HTML и записывает `telegram_message_id`; ошибка отправки откладывается на час.

### Системный мониторинг

Node-exporter читает CPU, память, файловые системы и сеть Linux-хоста через host PID/network namespaces и read-only `/host`. Его HTTP listener привязан к Docker gateway; Prometheus опрашивает `host.docker.internal:9100` с `job="node"`. Grafana показывает метрики в `Movie Tracker System` и выполняет системные правила из `monitoring/grafana/provisioning/alerting/system.json`; уведомления идут через существующий contact point и Alertmanager в Telegram.


## Backend layers

```text
cmd/*               composition root и lifecycle процессов
internal/handler    HTTP/Telegram transport, validation, DTO
internal/service    use cases и бизнес-правила
internal/repository PostgreSQL queries
internal/domain     модели и общие ошибки
internal/platform   PostgreSQL, Redis, JWT, TMDB, hashing, logging, metrics
migrations          версия схемы данных
```

Новую бизнес-логику помещайте в service, SQL — в repository, преобразование HTTP DTO — в handler. Wiring и route registration остаются в `cmd/api/main.go`.

## Карта изменений

| Область | Файлы |
| --- | --- |
| Login/register/refresh/logout | `handler/api/user.go`, `service/user.go`, `repository/{user,refreshtoken}.go`, `platform/{jwt,refreshtoken}` |
| Middleware, roles, trace ID | `handler/api/middleware.go` |
| Movie/TV catalogs and details | `handler/api/{movie,tv}.go`, одноимённые service и `platform/tmdb` |
| Watch list | `handler/api/watchlist.go`, `service/watchlist.go`, `repository/media.go` |
| Companies/collections/search | одноимённые handler/service/TMDB файлы |
| Telegram binding | `handler/api/telegram.go`, `handler/telegram`, `service/telegram.go`, `repository/telegram.go` |
| Schedules and reports | `service/worker.go`, `repository/worker.go`, migrations `000008`–`000011` |
| API metrics/logging | `platform/metrics`, `platform/logger`, `handler/api/middleware.go` |
| System metrics/dashboard/alerts | `compose.yaml`, `monitoring/prometheus/prometheus.yml`, `monitoring/grafana/dashboards/system.json`, `monitoring/grafana/provisioning/alerting/system.json`, `monitoring/check_system.py` |
| Web auth/proxy | `src/proxy.ts`, `src/app/api/backend`, `src/lib/auth`, `src/lib/api/client.ts` |
| Web feature pages | `src/app/<route>`, `src/components`, `src/lib/api` |

## Важные текущие ограничения

- Локальный `go run ./cmd/api` не находит миграции без абсолютного `/migrations`; контейнер API содержит этот каталог.
- `apps/web/nginx.conf` не используется текущим Next standalone Dockerfile и сохранён как legacy-файл.
- Go job в GitHub Actions пока только устанавливает Go и восстанавливает cache; тесты и build выполняются локально, но не в CI.
- Refresh cookie всегда создаётся с `Secure=true`; production должен работать по HTTPS.

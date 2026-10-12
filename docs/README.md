# Техническая документация Movie Tracker

Этот каталог — основная карта проекта. Перед изменением кода сначала прочитайте этот файл, затем документ нужной области. Полное повторное сканирование репозитория обычно не требуется.

## Быстрая ориентация

| Задача | Сначала читать | Основные исходники |
| --- | --- | --- |
| Изменить архитектуру или связность сервисов | [architecture.md](architecture.md) | `compose.yaml`, `apps/*` |
| Добавить или изменить API endpoint | [api.md](api.md) | `apps/api/cmd/api/main.go`, `internal/handler/api`, `internal/service` |
| Изменить страницу или получение данных | [frontend.md](frontend.md) | `apps/web/src/app`, `components`, `lib/api` |
| Изменить таблицы или фоновые задания | [database.md](database.md) | `apps/api/migrations`, `repository`, `service/worker.go` |
| Добавить переменную окружения | [configuration-and-data.md](configuration-and-data.md) | `.env.example`, `compose.yaml`, `internal/config/config.go` |
| Запустить проверки или изменить delivery | [development-and-release.md](development-and-release.md) | Dockerfiles, `.github/workflows/ci-cd.yml` |
| Разобраться с метриками или логами | [monitoring.md](monitoring.md) | `monitoring`, `deploy/nginx` |

## Инварианты проекта

- Браузер обращается к Go API через Next route handler `/api/backend/[...path]`.
- Публичный префикс Go API — `/api`; бизнес-методы находятся под `/api/v1`.
- PostgreSQL — источник постоянных данных; Redis можно отключить через `REDIS_DISABLED=true`.
- SQL-миграции применяет только процесс API при старте и только из `file:///migrations`.
- Access token и refresh token передаются в HTTP-only cookies; frontend выполняет refresh до запроса, если access cookie отсутствует.
- Worker синхронизирует расписания и формирует отчёты, Telegram-процесс принимает команды и отправляет подготовленные сообщения.
- Frontend включён в Compose profile `frontend`, observability stack — в profile `monitoring`.

## Поддержание документации

Документация изменяется в том же изменении, что и код:

| Изменение кода | Что актуализировать |
| --- | --- |
| Пользовательская возможность, способ запуска | корневой `README.md` |
| Компонент, зависимость или поток данных | `architecture.md` |
| Endpoint, body/query/response, auth или TTL | `api.md` |
| Route, proxy, client caching или auth UI | `frontend.md` |
| Таблица, индекс, enum или миграция | `database.md` |
| Env, port, Compose service/profile/volume | `configuration-and-data.md` |
| Команда проверки, Dockerfile, workflow, deploy | `development-and-release.md` |
| Метрика, dashboard, scrape target или log path | `monitoring.md` |

Не копируйте большие фрагменты кода в документацию. Фиксируйте контракт, поток данных, точки расширения и точные пути к исходникам.

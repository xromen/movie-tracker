# Данные и миграции

PostgreSQL 16 — источник постоянных данных. Миграции находятся в `apps/api/migrations`, имеют парные `*.up.sql`/`*.down.sql` и автоматически применяются API-процессом при старте.

## Схема

| Таблица / тип | Назначение и ключевые связи |
| --- | --- |
| `users` | email/username, bcrypt hash, `auth_version`, уникальный `telegram_id` |
| `roles`, `user_roles` | роли пользователей; миграция создаёт роль `admin`, но не назначает её |
| `refresh_tokens` | hash, family, rotation/revocation, expiration; cascade по user |
| enum `media_type` | `movie`, `tv` |
| `medias` | внутренний id, `(tmdb_id, media_type)` unique, metadata, sync state, optional collection |
| enum `watch_status` | `watched`, `want_to_watch`, `favorite` |
| `user_medias` | user ↔ internal media, status/rating/watched time; unique `(user_id, media_id)` |
| `user_tv_episodes` | отмеченные просмотренными TMDB episodes; composite primary key |
| `movie_collections` | TMDB collection и состояние фоновой синхронизации |
| `movie_releases` | даты релиза фильма по региону/type |
| `tv_episodes` | TMDB episode, внутренний media id, season/episode/date |
| `telegram_binding_tokens` | токен привязки пользователя, TTL 5 минут |
| `reports` | период пользователя и дата следующего формирования |
| `report_messages` | части Telegram-отчёта, retry state и Telegram message id |

`medias.release_date` хранится как `VARCHAR(10)`, потому что TMDB может возвращать неполные/строковые даты. Изображения и часть внешнего payload нормализуются в service/domain, а исходные дополнительные данные могут находиться в `metadata JSONB`.

## История миграций

| Версия | Изменение |
| --- | --- |
| `000001` | users, medias, user_medias, enums и базовые индексы |
| `000002` | `release_date` преобразован в строку |
| `000003` | просмотренные TV episodes |
| `000004` | roles и `auth_version` |
| `000005` | refresh token families |
| `000006` | уникальность media по `(tmdb_id, media_type)`, внутренний FK user_medias |
| `000007` | Telegram id и binding tokens |
| `000008` | sync state, movie releases, TV episode schedules |
| `000009` | movie collections |
| `000010` | reports и report message queue |
| `000011` | сохранение Telegram message id |

## Правила изменения схемы

1. Не редактируйте уже применённую миграцию; добавляйте следующую пару файлов.
2. `up` и `down` должны быть взаимно обратимыми настолько, насколько позволяет сохранность данных.
3. Обновите этот документ и affected repository/domain models в том же изменении.
4. Проверьте миграцию на новой базе и upgrade существующей базы.
5. Учитывайте, что API запускает `m.Up()` до поднятия HTTP server, а worker и bot миграции не применяют.

## Persistence paths

- API repositories: `apps/api/internal/repository`;
- pool configuration and statement timeout: `internal/platform/database/postgres.go`;
- media/watch-list queries: `repository/media.go`;
- schedules/reports: `repository/worker.go`;
- Telegram queue/binding: `repository/telegram.go`;
- refresh rotation transaction: `repository/refreshtoken.go`.

Compose использует внешний volume `movie-tracker-go_postgres_data`. Его удаление не является обычной операцией разработки: в нём находятся все пользовательские данные.

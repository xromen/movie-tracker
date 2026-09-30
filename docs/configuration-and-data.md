# Конфигурация

Корневой `.env` читается Docker Compose и передаётся Go-сервисам через `env_file`. Локальный Go-процесс вызывает `godotenv.Load()` из своей текущей директории. Next получает только переменные, явно переданные Compose или shell.

## Host и observability ports

| Переменная | Default | Назначение |
| --- | --- | --- |
| `WEB_PORT` | `3000` | frontend host port |
| `API_PORT` | `8080` | API host port |
| `POSTGRES_PORT` | `5432` | PostgreSQL host port |
| `REDIS_PORT` | `6379` | Redis host port |
| `GRAFANA_PORT` | `3001` | Grafana host port |
| `PROMETHEUS_PORT` | `9091` | Prometheus host port |
| `LOKI_PORT` | `3100` | Loki host port |
| `TELEGRAM_ALERT_CHAT_ID` | пусто | ID Telegram-чата для оповещений Grafana; токен берётся из `TELEGRAM_BOT_TOKEN` |

Все опубликованные порты привязаны к `127.0.0.1`.

## Web

| Переменная | Default | Назначение |
| --- | --- | --- |
| `MOVIE_TRACKER_API_URL` | `http://localhost:8080/api` | server-side адрес Go API; Compose задаёт `http://api:8080/api` |
| `NEXT_PUBLIC_APP_URL` | `http://localhost:3000` | fallback origin server-side proxy |
| `NEXT_PUBLIC_SITE_URL` | `https://movietracker.ru` в коде | metadata base и structured data; Compose default — APP URL |
| `MOVIE_TRACKER_ACCESS_TOKEN_COOKIE` | `access_token` | серверное имя access cookie |
| `MOVIE_TRACKER_REFRESH_TOKEN_COOKIE` | `refresh_token` | серверное имя refresh cookie |
| `NEXT_PUBLIC_ACCESS_TOKEN_COOKIE` | — | legacy fallback для имени access cookie |
| `NEXT_PUBLIC_REFRESH_TOKEN_COOKIE` | — | legacy fallback для имени refresh cookie |

Cookie overrides сейчас не передаются frontend-сервису из `compose.yaml`; при их использовании измените Compose согласованно с API.

## Go HTTP, database и Redis

| Переменная | Default |
| --- | --- |
| `PORT` | `8080` |
| `DB_HOST` | `localhost`; Compose: `postgres` |
| `DB_PORT` | `5432` |
| `DB_USER` | `postgres` |
| `DB_PASSWORD` | пусто в коде; template: `change-me` |
| `DB_NAME` | `movietracker` |
| `DB_SSLMODE` | `disable` |
| `DB_STATEMENT_TIMEOUT` | `15s` |
| `REDIS_ADDR` | `localhost:6379`; Compose: `redis:6379` |
| `REDIS_PASSWORD` | пусто |
| `REDIS_DB` | `0` |
| `REDIS_DISABLED` | `false` |
| `JWT_SECRET` | `change-me-in-production` |
| `LOG_FILE_PATH` | пусто; Compose задаёт путь каждому Go-процессу |

DB pool фиксирован кодом: max 25, min 5, max lifetime 5 минут, idle time 10 минут. HTTP read/write timeout — 10 секунд, idle timeout — 60 секунд. Access/refresh TTL не настраиваются env: 15 минут и 7 дней.

## TMDB и Telegram

| Переменная | Default | Назначение |
| --- | --- | --- |
| `TMDB_BASE_URL` | `https://api.themoviedb.org/3` | TMDB API |
| `TMDB_IMAGES_BASE_URL` | `https://api.themoviedb.org` | prefix изображений в API config |
| `TMDB_BEARER_TOKEN` | пусто | обязательный TMDB v4 token |
| `TMDB_TIMEOUT` | `10s` | HTTP timeout |
| `TMDB_RPM` | `4` | rate limiter events per second в текущем config name |
| `TMDB_BURST` | `40` | burst limit |
| `TELEGRAM_BOT_TOKEN` | пусто | обязателен для `telegram-bot` |
| `TELEGRAM_BOT_USERNAME` | пусто | используется в binding deep link |
| `TELEGRAM_API_BASE_URL` | `https://api.telegram.org` | Telegram API endpoint |

## Compose services, profiles и volumes

Без профиля доступны `api`, `telegram-bot`, `worker`, `postgres`, `redis`. `frontend` имеет profile `frontend`. Prometheus/Grafana/Alertmanager/Loki/Promtail/exporters имеют profile `monitoring`.

Alertmanager получает `TELEGRAM_API_BASE_URL` (по умолчанию `https://api.telegram.org`) и `TELEGRAM_ALERT_CHAT_ID` из Compose, а `TELEGRAM_BOT_TOKEN` через Compose secret. Адрес API должен быть доступен из контейнера Alertmanager; `localhost` внутри контейнера не указывает на хост. Alertmanager доступен только в сети Compose на порту `9093`.
Для inline-конфигурации Alertmanager нужен Docker Compose не ниже 2.23.1.

Профиль `monitoring` запускается самостоятельно: Prometheus опрашивает `frontend:3000` только если запущен frontend. Системный nginx отдаёт `stub_status` только на `127.0.0.1:8081`; exporter читает его через host network и по умолчанию слушает `:9113` на хосте, поэтому доступ к `9113` нужно ограничить firewall. Postgres-exporter подключается к `postgres:5432` с `DB_USER`/`DB_PASSWORD` и доступен только внутри сети Compose на `9187`. Redis-exporter подключается к `redis:6379` с `REDIS_PASSWORD` и доступен только внутри сети Compose на `9121`.
Каталог `monitoring/prometheus` монтируется в контейнер Prometheus целиком, поэтому новая версия `prometheus.yml` видна после выкладки.

| Volume | Данные |
| --- | --- |
| `postgres_data` | внешний persistent PostgreSQL volume |
| `api_logs` | API file logs, читаемые Promtail |
| `prometheus_data` | time series |
| `grafana_data` | Grafana state |
| `loki_data` | logs |
| `promtail_positions` | offsets прочитанных log files |

## Секреты

Не коммитьте `.env`. Для production обязательно замените `DB_PASSWORD`, `JWT_SECRET`, Grafana password и Telegram/TMDB tokens. Production frontend и auth должны быть доступны по HTTPS из-за `Secure` refresh cookie.

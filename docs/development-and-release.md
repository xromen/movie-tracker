# Разработка, CI и deployment

## Требования

- Docker Engine с Compose plugin;
- Node.js 22 и npm для web;
- Go 1.26.3 для API;
- TMDB v4 bearer token;
- Telegram token/username только для bot/report flow.

## Рабочий цикл

### Frontend

```bash
cd apps/web
npm ci
npm run dev
```

По умолчанию proxy ожидает API на `http://localhost:8080/api`. Локальные значения можно поместить в `apps/web/.env.local`.

### API

```bash
docker compose up -d postgres redis
docker compose up --build api
```

Container run предпочтителен: API применяет миграции из `/migrations`. Для нативного запуска `go run ./cmd/api` нужно смонтировать/создать абсолютный `/migrations` с содержимым `apps/api/migrations`.

Worker и bot можно запускать нативно из `apps/api` после настройки `.env`:

```bash
go run ./cmd/worker
go run ./cmd/bot
```

Оба требуют PostgreSQL; worker также требует TMDB, bot — Telegram config.

## Минимальные проверки

После изменений web:

```bash
cd apps/web
npm run typecheck
npm run lint
npm run build
```

После изменения web-метрик дополнительно выполните регрессионную проверку агрегации ID/неизвестных путей и histogram (Node.js 22.6+):

```bash
cd apps/web
node --experimental-strip-types --test src/lib/metrics/prometheus.test.mjs
```

После изменений Go:

```bash
cd apps/api
go test ./...
go vet ./...
```

Для конкурентных изменений auth/repository/worker дополнительно:

```bash
go test -race -count=1 ./...
```

После Docker/config изменений:

```bash
docker compose config
docker compose build
```

После изменения системного мониторинга проверьте provisioning и PromQL на синтетических метриках (Python 3.9+ и официальный `promtool` из дистрибутива Prometheus, без Python-зависимостей):

```bash
python monitoring/check_system.py /path/to/promtool
promtool check config monitoring/prometheus/prometheus.yml
```

Проверка читает фактические JSON дашборда и алертов, проверяет все PromQL-запросы, нормальный режим, warning/critical, граничные значения, pending periods, OOM и исключение временных/read-only файловых систем. На Windows передайте путь к `promtool.exe`. Это не заменяет проверку `up{job="node"}` и статуса provisioned rules после выкладки.

## CI/CD

Workflow: `.github/workflows/ci-cd.yml`.

- `web`: Node 22, `npm ci`, lint, typecheck, build;
- `api`: checkout и setup Go/cache; на текущий момент не запускает `go test` или `go build`;
- `compose`: после web/api jobs собирает Compose images;
- `deploy`: для push в `main`/`master` и manual dispatch запускается на self-hosted Linux x64 runner;
- перед сборкой self-hosted runner очищает неиспользуемый Docker build cache командой `docker builder prune --all --force --keep-storage 5GB`; следующие сборки могут заново скачать/собрать удалённые слои;
- deployment выполняет `docker compose --parallel 1 --profile frontend --profile monitoring up --build --detach --remove-orphans`;
- health gate до 60 секунд проверяет `/login` и `/api/health/live`.

Self-hosted runner должен заранее иметь production `.env`, внешний PostgreSQL volume, Docker Compose и доступ к нужным портам. Checkout использует `clean: false`, поэтому host-managed secrets сохраняются.
Очистка затрагивает кэш текущего Docker builder на хосте runner, включая неиспользуемый кэш других проектов, если хост общий; images, контейнеры и volumes этой командой не удаляются. Лимит относится к неиспользуемому кэшу после очистки, а не ко всему диску и не к пиковому размеру текущей сборки. Диагностика заполненного диска — в [monitoring.md](monitoring.md).

## Docker images

- API/worker/bot собираются статическими Go binaries и запускаются от UID/GID `65534` в `scratch`;
- только API image содержит `/migrations`;
- web image — multi-stage `node:22-alpine`, запускает `.next/standalone/server.js` от пользователя `nextjs`;
- контейнеры публикуются только на loopback; внешний TLS/reverse proxy настраивается вне текущего Compose.

## Checklist изменения

1. Измените минимально необходимый слой и его тесты.
2. Выполните проверки затронутого приложения.
3. Обновите документы по матрице из `docs/README.md`.
4. Проверьте `git diff` на secrets, generated output и случайные локальные файлы.
5. Коммиты пишите по-русски в прошедшем времени: краткая информативная первая строка, затем пункты выполненного.

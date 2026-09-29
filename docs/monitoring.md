# Мониторинг и логи

В compose добавлены Prometheus, Grafana, Alertmanager, Loki, Promtail, node-exporter, nginx-exporter, postgres-exporter и redis-exporter.

Запуск: `docker compose --profile frontend --profile monitoring up -d --build`. Профиль `monitoring` можно включить отдельно, но тогда target `movie-tracker-web` в Prometheus будет `down`, пока не запущен `frontend`.

По умолчанию:

- Grafana: `http://localhost:3001`, логин/пароль `admin`/`admin`
- Prometheus: `http://localhost:9091`
- Loki: `http://localhost:3100`
- API metrics: `http://localhost:8080/metrics`
- Web metrics: `http://localhost:3000/api/metrics`

Grafana автоматически подхватывает datasource `Prometheus` и `Loki`, а также дашборды:

- `Movie Tracker API`
- `Movie Tracker Web`
- `Movie Tracker Nginx`
- `Movie Tracker PostgreSQL`
- `Movie Tracker Redis`

PostgreSQL показывает доступность, число соединений, транзакции, долю попаданий в буферный кэш и deadlock по базам. Redis показывает доступность, клиентов, память, число ключей, команды, попадания/промахи и истечения/вытеснения ключей. Для скоростей используется окно 5 минут; сразу после запуска графики могут быть пустыми до накопления выборок.

## Оповещения в Telegram

Grafana загружает правила и contact point из `monitoring/grafana/provisioning/alerting/alerts.yml`. Contact point передаёт события во внутренний Alertmanager, который отправляет их в Telegram. В корневом `.env` на сервере задайте `TELEGRAM_BOT_TOKEN` существующего бота, `TELEGRAM_ALERT_CHAT_ID` нужного чата и при необходимости `TELEGRAM_API_BASE_URL` — тот же базовый адрес API, который использует бот. Значение по умолчанию — `https://api.telegram.org`; адрес должен быть доступен из контейнера Alertmanager. После изменения `.env` пересоздайте Alertmanager и Grafana командой `docker compose --profile monitoring up -d --force-recreate alertmanager grafana`. Затем в Grafana откройте **Alerting → Contact points → Movie Tracker Telegram → Test** и проверьте доставку.

Встроенный Telegram contact point Grafana не поддерживает смену базового адреса API, поэтому уведомления идут через Alertmanager. Если прокси работает по обычному HTTP через внешнюю сеть, токен бота передаётся ему без шифрования; используйте HTTPS или закрытый канал. Не публикуйте порт `9093` наружу.

| Событие | Условие | Период устойчивого условия |
| --- | --- | --- |
| Источник метрик недоступен | `up=0` для API, web, nginx-exporter, postgres-exporter, redis-exporter, Loki или Promtail; ошибка запроса Prometheus тоже считается аварией | 2 минуты |
| nginx не читается exporter | `nginx_up=0`; если exporter недоступен, сработает предыдущее правило | 2 минуты |
| PostgreSQL недоступен | `pg_up=0` или ошибка последнего сбора postgres-exporter | 2 минуты |
| Подозрительный всплеск трафика | nginx выше 100 RPS по среднему за 1 минуту | 3 минуты |
| Ошибки API | более 5% ответов 5xx при минимум 100 запросах за 5 минут | 5 минут |

Порог RPS меняется в `monitoring/grafana/provisioning/alerting/alerts.yml` (`evaluator.params` правила `movie_tracker_nginx_high_rps`). Это стартовый порог: сравните его с обычными пиками, затем подстройте. Уведомление о трафике означает необходимость проверить логи, IP, URL и User-Agent; высокий RPS сам по себе не доказывает DDoS. Grafana отправляет также сообщение о восстановлении по умолчанию. Если запущен только профиль `monitoring`, незапущенный frontend вызовет алерт о web.

Эти правила выполняет сама Grafana. При падении Grafana, Alertmanager или всего сервера сообщение может не дойти; для такого случая нужен внешний uptime-check. После выкладки проверьте статус правил в **Alerting → Alert rules** и доставку тестового сообщения. При ошибке доставки посмотрите `docker compose logs --tail=100 grafana alertmanager`.

Prometheus опрашивает postgres-exporter по `postgres-exporter:9187`; exporter подключается к PostgreSQL с учётными данными `DB_USER`/`DB_PASSWORD` из `.env` и не публикует свой порт на хосте.
Redis-exporter опрашивается по `redis-exporter:9121`, подключается к `redis:6379` с `REDIS_PASSWORD` из `.env` и также не публикует порт на хосте. Проверяйте отдельно `up{job="redis"}` (доступность exporter) и `redis_up{job="redis"}` (доступность самого Redis).
Exporter и Prometheus запускаются независимо от доступности API и PostgreSQL, чтобы продолжать проверку во время их отказа.

Панель Nginx показывает активные/ожидающие/читающие/записывающие соединения и RPS из `stub_status`, доступность exporter (`up{job="nginx"}`), успешность чтения самого nginx (`nginx_up{job="nginx"}`), логи и 50 самых активных User-Agent по среднему RPM за 5 минут. RPM считается запросом LogQL по access-логам стандартного nginx `combined` format. User-Agent извлекается во время запроса и не записывается как постоянный label Loki; при другом формате access-лога запрос панели нужно изменить. Строка User-Agent задаётся клиентом и сама по себе не доказывает, что запрос сделал бот.

API пишет JSON-логи одновременно в stdout и в файл `/var/log/movie-tracker/api.log` внутри контейнера. Файл лежит в volume `api_logs`, Promtail читает его и отправляет в Loki с label `job="movie-tracker-api"`.

API-метрики группируют неизвестные маршруты в `route="<unmatched>"`, чтобы произвольные URL не создавали неограниченное число рядов Prometheus.

Для nginx, установленного на сервере через systemctl, нужно включить `stub_status` на хосте. Подключите файл `deploy/nginx/monitoring.conf` внутри контекста `http`, затем проверьте и перезагрузите nginx:

```bash
sudo nginx -t
sudo systemctl reload nginx
curl -fsS http://127.0.0.1:8081/nginx_status
curl -fsS http://127.0.0.1:9113/metrics | grep '^nginx_up'
```

nginx-exporter запущен в host network и читает `http://127.0.0.1:8081/nginx_status`, а Promtail забирает `/var/log/nginx/access.log` и `/var/log/nginx/error.log` с хоста. В Prometheus `up=0` означает недоступность exporter из контейнера Prometheus, а `up=1` вместе с `nginx_up=0` — недоступность или неверный ответ `stub_status`. Проверяйте оба участка отдельно: `curl` на хосте к порту `8081`, затем `nginx_up` на порту `9113`, затем `up{job="nginx"}` в Prometheus. Если хостовый `curl :9113` работает, а `up=0`, проверьте Docker gateway и firewall; доступ к `9113` нужен только из подсети Compose. Если `nginx_up=1`, но RPS пустой, нужны минимум две выборки и запросы за последние 5 минут. Prometheus автоматически перечитывает изменённый `prometheus.yml`.

Для production задайте `GRAFANA_ADMIN_PASSWORD` в `.env` и ограничьте порт `9113` firewall: exporter в host network слушает все интерфейсы по умолчанию. Prometheus, Loki и Grafana опубликованы только на `127.0.0.1`; не публикуйте их наружу без аутентификации. API больше не пишет refresh token при ошибке обновления, но уже сохранённые логи могут содержать старые токены; проверьте их срок жизни и при необходимости отзовите сессии.

На Windows через Docker Desktop `node-exporter` показывает метрики Linux VM Docker Desktop, а не полноценные метрики Windows-хоста. Для production Linux-сервера этого достаточно для базового контроля CPU, памяти и дисков контейнерного окружения.

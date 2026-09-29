# Мониторинг и логи

В compose добавлены Prometheus, Grafana, Loki, Promtail, node-exporter, nginx-exporter и postgres-exporter.

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

## Оповещения в Telegram

Grafana загружает правила и Telegram contact point из `monitoring/grafana/provisioning/alerting/alerts.yml`. В корневом `.env` должны быть `TELEGRAM_BOT_TOKEN` существующего бота и `TELEGRAM_ALERT_CHAT_ID` чата, в который бот уже добавлен. На сервере эти значения нужно задать отдельно от локального `.env`; после изменения перезапустите Grafana. В Grafana откройте **Alerting → Contact points → Movie Tracker Telegram → Test**, чтобы проверить доставку.

| Событие | Условие | Период устойчивого условия |
| --- | --- | --- |
| Сервис недоступен | `up=0` для API, web, nginx-exporter, postgres-exporter, Loki или Promtail; ошибка запроса Prometheus тоже считается аварией | 2 минуты |
| nginx не читается exporter | `nginx_up=0`; если exporter недоступен, сработает предыдущее правило | 2 минуты |
| PostgreSQL недоступен | `pg_up=0` или ошибка последнего сбора postgres-exporter | 2 минуты |
| Подозрительный всплеск трафика | nginx выше 100 RPS по среднему за 1 минуту | 3 минуты |
| Ошибки API | более 5% ответов 5xx при минимум 100 запросах за 5 минут | 5 минут |

Порог RPS меняется в `monitoring/grafana/provisioning/alerting/alerts.yml` (`evaluator.params` правила `movie_tracker_nginx_high_rps`). Это стартовый порог: сравните его с обычными пиками, затем подстройте. Уведомление о трафике означает необходимость проверить логи, IP, URL и User-Agent; высокий RPS сам по себе не доказывает DDoS. Grafana отправляет также сообщение о восстановлении по умолчанию. Если запущен только профиль `monitoring`, незапущенный frontend вызовет алерт о web.

Эти правила выполняет сама Grafana. При падении Grafana или всего сервера она не сможет отправить сообщение; для такого случая нужен внешний uptime-check. После выкладки проверьте статус правил в **Alerting → Alert rules** и доставку тестового сообщения.

Prometheus опрашивает postgres-exporter по `postgres-exporter:9187`; exporter подключается к PostgreSQL с учётными данными `DB_USER`/`DB_PASSWORD` из `.env` и не публикует свой порт на хосте.
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

nginx-exporter запущен в host network и читает `http://127.0.0.1:8081/nginx_status`, а Promtail забирает `/var/log/nginx/access.log` и `/var/log/nginx/error.log` с хоста. В Prometheus `up=0` означает недоступность exporter, а `up=1` вместе с `nginx_up=0` — недоступность или неверный ответ `stub_status`. Если `nginx_up=1`, но RPS пустой, нужны минимум две выборки и запросы за последние 5 минут.

Для production задайте `GRAFANA_ADMIN_PASSWORD` в `.env` и ограничьте порт `9113` firewall: exporter в host network слушает все интерфейсы по умолчанию. Prometheus, Loki и Grafana опубликованы только на `127.0.0.1`; не публикуйте их наружу без аутентификации. API больше не пишет refresh token при ошибке обновления, но уже сохранённые логи могут содержать старые токены; проверьте их срок жизни и при необходимости отзовите сессии.

На Windows через Docker Desktop `node-exporter` показывает метрики Linux VM Docker Desktop, а не полноценные метрики Windows-хоста. Для production Linux-сервера этого достаточно для базового контроля CPU, памяти и дисков контейнерного окружения.

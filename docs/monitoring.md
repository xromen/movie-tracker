# Мониторинг и логи

В compose добавлены Prometheus, Grafana, Alertmanager, Loki, Promtail, node-exporter, nginx-exporter, postgres-exporter и redis-exporter.

Запуск: `docker compose --profile frontend --profile monitoring up -d --build`. Профиль `monitoring` можно включить отдельно, но тогда target `movie-tracker-web` в Prometheus будет `down`, пока не запущен `frontend`.

По умолчанию:

- Grafana: `http://localhost:3001`, логин/пароль `admin`/`admin`
- Prometheus: `http://localhost:9091`
- Loki: `http://localhost:3100`
- API metrics: `http://localhost:8080/metrics`
- Web metrics: `http://localhost:3000/api/metrics`

## Grafana за nginx

Для отдельного HTTPS-домена задайте в серверном `.env` полный внешний URL со слешем в конце:

```env
GRAFANA_ROOT_URL=https://grafana.movietracker.ru/
```

Compose передаёт его в `GF_SERVER_ROOT_URL`. Без переопределения используется `http://localhost:${GRAFANA_PORT:-3001}/`. Для отдельного домена без подпути `serve_from_sub_path` не нужен. После изменения пересоздайте Grafana:

```bash
docker compose --profile monitoring up -d --force-recreate grafana
```

Системный nginx должен проксировать корень домена в `http://127.0.0.1:${GRAFANA_PORT:-3001}` с исходным Host и поддержкой WebSocket для `/api/live/`. Настройки URL и reverse proxy описаны в [документации Grafana](https://grafana.com/tutorials/run-grafana-behind-a-proxy/).

При странице «Grafana has failed to load its application files» проверьте в браузере Network полную загрузку `/public/build/*.js` и ошибки Console. Ответ `200` на HEAD или работающий `/api/health` ещё не подтверждают загрузку тела файла. Если большие файлы обрываются, сравните скачивание одного и того же JS напрямую с `127.0.0.1:3001` и через HTTPS-домен (путь возьмите из Network):

```bash
curl -fsS http://127.0.0.1:3001/public/build/FILE.js -o /dev/null -w '%{size_download}\n'
curl -fsS https://grafana.movietracker.ru/public/build/FILE.js -o /dev/null -w '%{size_download}\n'
sudo nginx -T 2>&1 | grep -E 'error_log|proxy_temp_path'
df -h
df -i
```

Проверьте активный error log из `nginx -T`, свободное место и права пользователя nginx на временный каталог. Если причина — заполненный диск, сначала восстановите свободное место и повторите скачивание; при успешной передаче изменение буферизации не требуется. `root_url` исправляет внешние ссылки и редиректы; обрыв передачи JS диагностируется отдельно. После восстановления обновите страницу без кэша.

Grafana автоматически подхватывает datasource `Prometheus` и `Loki`, а также дашборды:

- `Movie Tracker API`
- `Movie Tracker Web`
- `Movie Tracker Nginx`
- `Movie Tracker PostgreSQL`
- `Movie Tracker Redis`
- `Movie Tracker System`

PostgreSQL показывает доступность, число соединений, транзакции, долю попаданий в буферный кэш и deadlock по базам. Redis показывает доступность, клиентов, память, число ключей, команды, попадания/промахи и истечения/вытеснения ключей. Для скоростей используется окно 5 минут; сразу после запуска графики могут быть пустыми до накопления выборок.

### Системный дашборд и алерты

`Movie Tracker System` (`/d/movie-tracker-system`) загружается из `monitoring/grafana/dashboards/system.json` и содержит 16 панелей: доступность node-exporter, CPU и I/O wait, доступную RAM, доступное место на `/`, uptime, load на ядро, RAM в байтах, swap, свободное место и inode по разделам, дисковый I/O, сеть хоста и OOM. RAM считается через `MemAvailable`, а не только `MemFree`: освобождаемый кэш не создаёт ложный алерт. Load нормализуется по количеству CPU-ядер. Отсутствующий swap отображается нулевым объёмом. Скорости вычисляются за 5 минут.

Метрики идут от существующего node-exporter с `job="node"`. Для корректной сети хоста он работает в host network/PID namespace, root filesystem монтируется read-only с `rslave`. Listener `host.docker.internal:9100` привязан к адресу Docker gateway через `extra_hosts: host-gateway`; exporter не слушает `0.0.0.0:9100`. Prometheus опрашивает этот же адрес. Gateway должен быть доступен из сети Compose. Схема host namespaces и root mount основана на [рекомендациях node-exporter](https://github.com/prometheus/node_exporter#docker).

При host network пустой столбец `PORTS` в `docker ps` ожидаем: exporter использует сетевой namespace хоста без публикации порта. Если target `node` имеет `health=down` и `context deadline exceeded`, проверьте listener и подсеть Prometheus:

```bash
docker compose logs --tail=30 node-exporter
docker inspect movie-tracker-prometheus --format '{{json .NetworkSettings.Networks}}'
sudo ufw status verbose
```

Проверьте `/metrics` с хоста по адресу из строки `Listening on`. Если он быстро отвечает `200`, а UFW запрещает входящие соединения и не разрешает `9100` из сети Compose, добавьте точечное правило. Например, для подсети Prometheus `172.20.0.0/16` и listener `172.17.0.1:9100`:

```bash
sudo ufw allow proto tcp from 172.20.0.0/16 to 172.17.0.1 port 9100 comment 'Movie Tracker node-exporter'
```

Подставьте фактические подсеть и адрес listener; доступ нужен только из сети Compose. При изменении этой сети актуализируйте правило. Перезапуск контейнеров после изменения UFW не требуется: через два интервала сбора (около 30 секунд) проверьте `up{job="node"}=1`. Графикам скоростей нужны минимум две выборки.

Правила находятся в `monitoring/grafana/provisioning/alerting/system.json`, группе `Movie Tracker system`, вычисляются раз в минуту и отправляют события в существующий `Movie Tracker Telegram`. Каждое правило связано с соответствующей панелью дашборда. Для нехватки ресурсов заданы два уровня:

| Событие | Условие | Устойчивость | Уровень |
| --- | --- | --- | --- |
| Мало доступной RAM | менее 15% / менее 5% | 5 минут / 2 минуты | warning / critical |
| Мало места на разделе | менее 20% / менее 10% доступных байтов | 10 минут / 5 минут | warning / critical |
| Мало inode на разделе | менее 10% / менее 3% свободных inode | 10 минут / 5 минут | warning / critical |
| Высокая загрузка CPU | выше 90%, среднее по ядрам за 5 минут | 10 минут | warning |
| Высокий I/O wait | выше 20%, среднее по ядрам за 5 минут | 10 минут | warning |
| Высокий системный load | load15 / число ядер выше 1.5 | 15 минут | warning |
| OOM killer завершил процесс | `increase(node_vmstat_oom_kill[5m]) > 0.5` | без дополнительного ожидания | critical |
| Node-exporter недоступен | `up{job="node"}=0`, общее правило доступности | 2 минуты | critical |

Дисковые правила и графики по разделам исключают tmpfs, devtmpfs, overlay, squashfs, nsfs, `/run` и внутренние mounts Docker/containerd/kubelet. Read-only разделы и разделы без положительного размера/числа inode не вызывают low-space/low-inode алерты. Пороги меняются в `data[1].model.conditions[0].evaluator.params`, длительность — в `for`. Если critical условие устойчиво достаточно долго, одновременно могут быть активны warning и critical. Отсутствие системной метрики (`NoData`) не порождает отдельный ресурсный алерт: доступность exporter/Prometheus проверяет общее правило; ошибки запросов имеют состояние `Error`.

При обычной CI-выкладке Compose пересоздаёт node-exporter из-за изменённых параметров, а Prometheus автоматически перечитывает конфигурацию. Новые правила алертов Grafana загружаются при запуске: если Grafana не перезапустилась во время выкладки, достаточно перезапустить только её:

```bash
docker compose --profile monitoring restart grafana
curl -fsS 'http://127.0.0.1:9091/api/v1/query?query=up%7Bjob%3D%22node%22%7D'
```

Для вашего `PROMETHEUS_PORT` замените `9091` при необходимости. Ожидается `up=1` с instance `host.docker.internal:9100`. Затем откройте `Movie Tracker System`, проверьте правила группы `Movie Tracker system` в Grafana Alerting и тест доставки существующего contact point. Проверьте формулы до выкладки командой `python monitoring/check_system.py /path/to/promtool`: используется стандартная библиотека Python и официальный promtool; покрыты все запросы панелей, warning/critical, границы порогов, pending, OOM и исключённые mounts.

## Оповещения в Telegram

Grafana загружает правила и contact point из `monitoring/grafana/provisioning/alerting/alerts.yml`. Contact point передаёт события во внутренний Alertmanager, который отправляет их в Telegram. В корневом `.env` на сервере задайте `TELEGRAM_BOT_TOKEN` существующего бота, `TELEGRAM_ALERT_CHAT_ID` нужного чата и при необходимости `TELEGRAM_API_BASE_URL` — тот же базовый адрес API, который использует бот. Значение по умолчанию — `https://api.telegram.org`; адрес должен быть доступен из контейнера Alertmanager. После изменения `.env` пересоздайте Alertmanager и Grafana командой `docker compose --profile monitoring up -d --force-recreate alertmanager grafana`. Затем в Grafana откройте **Alerting → Contact points → Movie Tracker Telegram → Test** и проверьте доставку.

Встроенный Telegram contact point Grafana не поддерживает смену базового адреса API, поэтому уведомления идут через Alertmanager. Если прокси работает по обычному HTTP через внешнюю сеть, токен бота передаётся ему без шифрования; используйте HTTPS или закрытый канал. Не публикуйте порт `9093` наружу.

Alertmanager использует `monitoring/alertmanager/telegram.tmpl` и `parse_mode: ""` (обычный текст). Сообщение содержит количество проблем и восстановлений, summary, job/instance и причину `NoData`/`MissingSeries`, если она передана Grafana. Служебные labels/annotations не выводятся. Показываются максимум пять событий; длина каждого поля ограничена, а для большей группы выводится общее число и предложение открыть Grafana. Это удерживает сообщение ниже лимита Telegram в 4096 символов, включая текст с emoji. После обновления шаблона или Compose пересоздайте Alertmanager командой выше. Проверка шаблона не отправляет сообщения:

```bash
cd monitoring/alertmanager
go test -v telegram_test.go
```

| Событие | Условие | Период устойчивого условия |
| --- | --- | --- |
| Источник метрик недоступен | `up=0` для API, web, nginx-exporter, postgres-exporter, redis-exporter, Loki, Promtail или node-exporter; ошибка запроса Prometheus тоже считается аварией | 2 минуты |
| nginx не читается exporter | `nginx_up=0`; если exporter недоступен, сработает предыдущее правило | 2 минуты |
| PostgreSQL недоступен | `pg_up=0` или ошибка последнего сбора postgres-exporter | 2 минуты |
| Подозрительный всплеск трафика | nginx выше 100 RPS по среднему за 1 минуту | 3 минуты |
| Ошибки API | более 5% ответов 5xx при минимум 100 запросах за 5 минут | 5 минут |

Порог RPS меняется в `monitoring/grafana/provisioning/alerting/alerts.yml` (`evaluator.params` правила `movie_tracker_nginx_high_rps`). Это стартовый порог: сравните его с обычными пиками, затем подстройте. Уведомление о трафике означает необходимость проверить логи, IP, URL и User-Agent; высокий RPS сам по себе не доказывает DDoS. Grafana отправляет также сообщение о восстановлении по умолчанию. Если запущен только профиль `monitoring`, незапущенный frontend вызовет алерт о web.

Эти правила выполняет сама Grafana. При падении Grafana, Alertmanager или всего сервера сообщение может не дойти; для такого случая нужен внешний uptime-check. После выкладки проверьте статус правил в **Alerting → Alert rules** и доставку тестового сообщения. При ошибке доставки посмотрите `docker compose logs --tail=100 grafana alertmanager`.

`grafana_state_reason=NoData` с `__values__={"A":-1,"B":-1}` означает отсутствие данных для вычисления правила, а не измеренное `up=0`. `MissingSeries` означает, что ранее наблюдавшийся ряд исчез. Не отключайте аварийное состояние только ради устранения сообщений: сначала проверьте Prometheus и targets:

```bash
docker compose --profile monitoring ps
curl -fsS 'http://127.0.0.1:9091/api/v1/query?query=up'
curl -fsS 'http://127.0.0.1:9091/api/v1/targets?state=active'
docker compose logs --tail=100 prometheus grafana postgres-exporter
```

Если `PROMETHEUS_PORT` переопределён, используйте его вместо `9091`. Пустой result требует проверки scrape config, targets и логов Prometheus. Если ряды `up` присутствуют, проверьте в Grafana datasource `Prometheus` (`http://prometheus:9090`) и выполнение запроса правила через Preview. Ошибка интерфейса Grafana и `NoData` сами по себе не доказывают общую причину.

### Заполненный диск

Если targets имеют `health=up`, но запрос `up` пустой, проверьте ошибки `Scrape commit failed` и `write to WAL ... no space left on device`. Prometheus может успешно опрашивать сервисы, но не сохранять результаты. Заполненный корневой раздел также может прерывать большие ответы nginx при записи proxy temp files и мешать записи в журналы.

```bash
df -h
df -i
docker system df
sudo du -xhd1 /var/lib/docker /var/log 2>/dev/null | sort -h
```

Очистку выбирайте по измеренному источнику роста: неиспользуемый build cache и устаревшие образы, ротация Docker/системных логов или расширение диска. Не удаляйте `prometheus_data`, WAL или PostgreSQL volume для освобождения места: это уничтожит сохранённые данные. Сначала освободите место, затем проверьте возобновление записи и повторите запрос `up`; перезапустите Prometheus только если запись не возобновилась. После восстановления проверьте правила Grafana и полную загрузку её JS через nginx. Ограничения хранения и сборка мусора должны настраиваться для того компонента, который занял диск.

Если `docker system df` показывает большой reclaimable build cache, освободите неиспользуемые слои, оставив до 5 ГБ кэша:

```bash
docker builder prune --all --keep-storage 5GB
df -h /
```

Команда запросит подтверждение и удалит только неиспользуемый build cache; следующая сборка может стать медленнее. Данные volumes и работающие контейнеры сохраняются. CI выполняет такую очистку перед production-сборкой с `--force`. Это не жёсткая квота и не ограничение размеров логов/volumes; поддерживайте свободное место и после сборки. Семантика параметров — в [документации Docker](https://docs.docker.com/reference/cli/docker/builder/prune/).

Prometheus опрашивает postgres-exporter по `postgres-exporter:9187`; exporter подключается к PostgreSQL с учётными данными `DB_USER`/`DB_PASSWORD` из `.env` и не публикует свой порт на хосте.
Каталог `monitoring/prometheus` монтируется целиком, чтобы Prometheus видел обновления `prometheus.yml` после выкладки и автоматически перечитывал конфигурацию.
Redis-exporter опрашивается по `redis-exporter:9121`, подключается к `redis:6379` с `REDIS_PASSWORD` из `.env` и также не публикует порт на хосте. Проверяйте отдельно `up{job="redis"}` (доступность exporter) и `redis_up{job="redis"}` (доступность самого Redis).
Exporter и Prometheus запускаются независимо от доступности API и PostgreSQL, чтобы продолжать проверку во время их отказа.

Панель Nginx показывает активные/ожидающие/читающие/записывающие соединения и RPS из `stub_status`, доступность exporter (`up{job="nginx"}`), успешность чтения самого nginx (`nginx_up{job="nginx"}`), логи и 50 самых активных User-Agent по среднему RPM за 5 минут. RPM считается запросом LogQL по access-логам стандартного nginx `combined` format. User-Agent извлекается во время запроса и не записывается как постоянный label Loki; при другом формате access-лога запрос панели нужно изменить. Строка User-Agent задаётся клиентом и сама по себе не доказывает, что запрос сделал бот.

API пишет JSON-логи одновременно в stdout и в файл `/var/log/movie-tracker/api.log` внутри контейнера. Файл лежит в volume `api_logs`, Promtail читает его и отправляет в Loki с label `job="movie-tracker-api"`.

API-метрики группируют неизвестные маршруты в `route="<unmatched>"`, чтобы произвольные URL не создавали неограниченное число рядов Prometheus.

Web-метрики `movie_tracker_web_backend_requests_total` и `movie_tracker_web_backend_request_duration_seconds` также используют ограниченный набор `route`: шаблоны API из `apps/web/src/lib/metrics/prometheus.ts`, например `/v1/movie/:id` и `/v1/tv/:id/season/:season_number`. Конкретные ID не попадают в labels; неизвестные пути объединяются в `<unmatched>`. Query string proxy не передаёт в метрики. При добавлении API endpoint обновляйте список шаблонов; статические маршруты должны стоять перед динамическими. Имена метрик, histogram buckets, method/status и запросы существующего web-дашборда сохраняются.

### Высокое потребление памяти Prometheus

Проверьте `docker stats --no-stream movie-tracker-prometheus` и `curl -fsS http://127.0.0.1:9091/api/v1/status/tsdb` (замените порт своим `PROMETHEUS_PORT`). В TSDB смотрите `headStats.numSeries`, `seriesCountByMetricName` и `labelValueCountByLabelName`. Большое число разных `route` у web-метрик означает сбор конкретных URL вместо шаблонов: каждый набор method/route/status создаёт 15 рядов (counter, 12 histogram buckets, sum и count). Web registry сохраняет их до перезапуска процесса и повторно отдаёт при каждом scrape, даже если новых запросов по URL нет.

После исправления пересоберите frontend на сервере из обновлённого checkout:

```bash
docker compose --profile frontend --profile monitoring up -d --no-deps --build frontend
```

Новый web-процесс начинает registry с шаблонов маршрутов. Старые ряды перестают поступать в Prometheus, но память освободится постепенно после compaction текущего блока и очистки неактивных рядов; это занимает несколько часов. Перезапуск только Prometheus не устраняет причину и повторно загружает текущие данные из WAL. Проверяйте TSDB и память после нескольких циклов compaction. Старую историю можно продолжать читать до истечения retention; широкие запросы по ней всё ещё могут потреблять много ресурсов. Не удаляйте volume или WAL для ускорения очистки.

Retention по умолчанию — 15 дней: уменьшение до недели или двух дней прежде всего экономит диск, а не исправляет рост числа активных рядов. Сначала исправьте labels, затем подбирайте интервал сбора и лимит памяти по фактическому потреблению. Уменьшение лимита контейнера ниже рабочего объёма может вызвать OOM и перезапуски. Устройство head/WAL и retention описано в [документации Prometheus](https://prometheus.io/docs/prometheus/latest/storage/).

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

# Мониторинг и логи

В compose добавлены Prometheus, Grafana, Loki, Promtail, node-exporter и nginx-exporter.

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

API пишет JSON-логи одновременно в stdout и в файл `/var/log/movie-tracker/api.log` внутри контейнера. Файл лежит в volume `api_logs`, Promtail читает его и отправляет в Loki с label `job="movie-tracker-api"`.

Для nginx, установленного на сервере через systemctl, нужно включить `stub_status` на хосте. Подключите файл `deploy/nginx/monitoring.conf` в конфиг nginx и перезагрузите сервис:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

nginx-exporter запущен в host network и читает `http://127.0.0.1/nginx_status`, а Promtail забирает `/var/log/nginx/access.log` и `/var/log/nginx/error.log` с хоста.

На Windows через Docker Desktop `node-exporter` показывает метрики Linux VM Docker Desktop, а не полноценные метрики Windows-хоста. Для production Linux-сервера этого достаточно для базового контроля CPU, памяти и дисков контейнерного окружения.

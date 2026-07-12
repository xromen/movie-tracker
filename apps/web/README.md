# Movie Tracker Web

Frontend-приложение на Next.js для поиска фильмов и сериалов, просмотра деталей, трейлеров, рекомендаций, сезонов и ведения личного списка просмотра.

## Возможности

- каталоги фильмов и сериалов с фильтрами и пагинацией;
- поиск из шапки через общий `/search/multi`;
- страницы деталей для `movie` и `tv`;
- постеры, рейтинги, жанры, даты выхода, трейлеры, коллекции и рекомендации;
- сезоны сериалов и отметки просмотра сезонов/эпизодов;
- регистрация, вход, выход и автоматический refresh сессии;
- защищенная страница `/watchlist`;
- статусы просмотра `watched`, `want_to_watch`, `favorite`;
- backend proxy через Next route handler `/api/backend/[...path]`;
- SEO metadata и structured data для публичных страниц;
- локальные Inter-шрифты и standalone production build.

## Стек

- Next.js 16 App Router
- React 19
- TypeScript 6
- CSS Modules
- TanStack React Query
- Zustand
- Embla Carousel
- Lucide React
- React Content Loader
- ESLint

## Требования

- Node.js 22 для Docker-образа и CI; локально лучше использовать Node.js 22 или новее
- npm
- запущенный backend API

По умолчанию серверный proxy обращается к API по адресу:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
```

Публичные URL приложения:

```env
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

## Запуск

1. Установите зависимости:

```bash
npm install
```

2. При необходимости создайте `.env.local`:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_SITE_URL=http://localhost:3000
```

3. Запустите dev-сервер:

```bash
npm run dev
```

4. Откройте адрес, который выведет Next.js, обычно `http://localhost:3000`.

## Скрипты

```bash
npm run dev
```

Запускает приложение в режиме разработки.

```bash
npm run build
```

Собирает production-версию Next.js в standalone-режиме.

```bash
npm run start
```

Запускает production-сборку после `npm run build`.

```bash
npm run lint
```

Запускает ESLint для TypeScript/React-кода.

```bash
npm run typecheck
```

Запускает TypeScript без генерации файлов.

## Маршруты приложения

| Путь | Назначение |
| --- | --- |
| `/` | Каталог фильмов |
| `/movie` | Каталог фильмов с query `filter` и `page` |
| `/tv` | Каталог сериалов с query `filter` и `page` |
| `/details/[mediaType]/[mediaId]` | Детальная страница фильма или сериала |
| `/login` | Вход |
| `/register` | Регистрация |
| `/watchlist` | Личный список просмотра |
| `/health` | Клиентская страница проверки доступности |
| `/api/backend/[...path]` | Proxy к Go API |

## Backend proxy и auth

Весь frontend-клиент ходит в API через `/api/backend`. В браузере используется относительный путь, а при server-side fetch Next строит абсолютный URL из `Host`/`X-Forwarded-*` или `NEXT_PUBLIC_APP_URL`.

Proxy:

- прокидывает `accept`, `content-type` и `cookie`;
- возвращает `cache-control`, `content-type` и `Set-Cookie`;
- при отсутствии access token и наличии refresh token делает single-flight refresh;
- выставляет заголовок `X-Auth-Session-Changed`, чтобы клиент обновил состояние auth-сессии.

## Docker

Dockerfile использует многостадийную сборку на `node:22-alpine`, выполняет `npm ci`, `npm run build` и запускает standalone-сервер:

```bash
docker build -t movie-tracker-web .
docker run --rm -p 3000:3000 \
  -e MOVIE_TRACKER_API_URL=http://host.docker.internal:8080/api \
  -e NEXT_PUBLIC_APP_URL=http://localhost:3000 \
  -e NEXT_PUBLIC_SITE_URL=http://localhost:3000 \
  movie-tracker-web
```

В корневом `compose.yaml` контейнер получает `MOVIE_TRACKER_API_URL=http://api:${PORT:-8080}/api`.

## Структура

```text
src/
  app/                    маршруты Next App Router, страницы и route handlers
  components/             переиспользуемые UI- и layout-компоненты
  lib/api/                API-клиент, DTO-мапперы и backend proxy helpers
  lib/auth/               cookie/session/JWT helpers и refresh single-flight
  lib/seo/                structured data
  lib/utils/              общие утилиты
public/                   favicon, robots.txt, icons.svg, локальные шрифты
```

## Проверки перед изменениями

```bash
npm run typecheck
npm run lint
npm run build
```

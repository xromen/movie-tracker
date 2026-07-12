# Movie Tracker

Movie Tracker - фронтенд-приложение на Next.js для поиска фильмов и сериалов, просмотра подробной информации, трейлеров, рекомендаций и ведения личного списка просмотра.

## Возможности

- каталоги фильмов и сериалов с фильтрами и пагинацией;
- быстрый поиск по фильмам и сериалам из шапки приложения;
- страницы деталей с постером, рейтингом, жанрами, датами выхода, трейлерами, коллекциями, сезонами и рекомендациями;
- регистрация и вход пользователя;
- защищенный раздел личного списка;
- статусы просмотра: просмотрено, хочу посмотреть, избранное;
- проксирование backend API через Next route handler с cookie-based auth и refresh-flow;
- SEO metadata для публичных страниц.

## Стек

- Next.js 16 App Router
- React 19
- TypeScript
- CSS Modules
- Embla Carousel
- Lucide React
- ESLint

## Требования

- Node.js 20 или новее
- npm
- запущенный backend API

По умолчанию приложение обращается к backend API по адресу:

```env
http://localhost:8080/api
```

Адрес можно переопределить переменной окружения:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
```

Для публичного абсолютного URL сайта можно задать:

```env
NEXT_PUBLIC_SITE_URL=https://movietracker.ru
```

## Запуск

1. Установите зависимости:

```bash
npm install
```

2. Создайте локальный `.env.local` при необходимости:

```env
MOVIE_TRACKER_API_URL=http://localhost:8080/api
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

Собирает production-версию и проверяет Next/TypeScript pipeline.

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

## Структура проекта

```text
src/
  app/            маршруты Next App Router, страницы и route handlers
  components/     переиспользуемые UI- и layout-компоненты
  lib/api/        API-клиент, DTO-мапперы и backend proxy helpers
  lib/auth/       cookie/session/JWT helpers
  lib/seo/        структурированные данные
  lib/utils/      общие утилиты
public/           favicon, robots.txt, локальные шрифты и статические ассеты
```

## Проверка перед изменениями

Минимальный набор локальных проверок:

```bash
npm run typecheck
npm run lint
npm run build
```
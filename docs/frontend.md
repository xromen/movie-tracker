# Frontend

## Стек и структура

Next.js 16 App Router, React 19, TypeScript 6, CSS Modules, TanStack React Query, Zustand и Embla. Production build использует `output: standalone`.

```text
src/app/          routes, pages, route handlers
src/components/   UI и layout components
src/lib/api/      API client, DTO и feature functions
src/lib/auth/     cookies, JWT claims, session, single-flight refresh
src/lib/metrics/  web Prometheus registry
src/lib/seo/      structured data
public/           favicon, icons, robots.txt, Inter fonts
```

## Routes

| Route | Назначение |
| --- | --- |
| `/`, `/movie` | каталог фильмов; `/movie` принимает `filter`, `page` |
| `/tv` | каталог сериалов; `filter`, `page` |
| `/details/[mediaType]/[mediaId]` | детали movie/TV, видео, рекомендации, сезоны/коллекции |
| `/company/[companyId]/[mediaType]` | компания и её movie/TV; `page` |
| `/login`, `/register` | auth формы |
| `/watchlist` | защищённый личный список |
| `/health` | защищённая admin-only страница readiness, refresh каждые 30 секунд |
| `/api/backend/[...path]` | BFF proxy к Go API |
| `/api/metrics` | Prometheus metrics frontend proxy |

`/` отображает movie catalog. Допустимые фильтры определяются в `MovieCatalog.tsx` и `TvCatalog.tsx` и преобразуются в API paths через `lib/api/media.ts`.

## API client и caching

`fetchApi` всегда строит URL через `/api/backend`. В браузере используется относительный URL. На сервере origin берётся из `X-Forwarded-Host`/`Host` и `X-Forwarded-Proto`, fallback — `NEXT_PUBLIC_APP_URL`.

- server request с cookie: `cache: no-store`;
- auth paths: всегда `no-store`;
- публичный server GET: Next revalidate 300 секунд по умолчанию;
- company detail: 3600 секунд;
- movie/TV detail: 900 секунд.

Не обращайтесь из UI напрямую к Go API: иначе потеряются единый cookie flow, refresh и web metrics.

Web registry (`src/lib/metrics/prometheus.ts`) записывает backend-вызовы по шаблонам маршрутов, например `/v1/movie/:id`, а неизвестные пути объединяет в `<unmatched>`. Это ограничивает число рядов и память web/Prometheus при обходе большого каталога. При добавлении API endpoint обновите список шаблонов; ID и query string не должны становиться labels. Диагностика и порядок выкладки исправлений метрик описаны в [monitoring.md](monitoring.md).

## Auth flow

1. Go API устанавливает `access_token` и `refresh_token`.
2. `getSession()` декодирует access JWT на сервере для username/roles; подпись frontend не проверяет, окончательная проверка остаётся в API.
3. `src/proxy.ts` обновляет токены для page requests, если access cookie отсутствует, но refresh cookie есть.
4. BFF route делает то же перед API request.
5. `X-Auth-Session-Changed: 1` вызывает browser event; `AuthSessionListener` делает `router.refresh()`.
6. `/watchlist` перенаправляет гостя на `/login?from=/watchlist`; `/health` дополнительно требует role `admin`.

Имена cookies можно переопределить серверными env `MOVIE_TRACKER_ACCESS_TOKEN_COOKIE` и `MOVIE_TRACKER_REFRESH_TOKEN_COOKIE`; legacy-имена читаются для совместимости.

## Где менять функциональность

- route metadata/redirect/data loading — `src/app/<route>/page.tsx`;
- интерактивное состояние — client component рядом со страницей или в `components`;
- HTTP contract — сначала `src/lib/api/types.ts`, затем feature module;
- auth/proxy — `src/proxy.ts`, `src/app/api/backend`, `src/lib/auth`;
- глобальный layout/header/footer — `src/components/layout`;
- SEO JSON-LD — `src/lib/seo/structuredData.ts`.

## Проверки

```bash
npm ci
npm run typecheck
npm run lint
npm run build
```

Node.js 22 используется в Docker и CI. `apps/web/nginx.conf` не подключён к текущей standalone-сборке.

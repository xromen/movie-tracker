export type JsonLd = Record<string, unknown>;

const SITE_NAME = "Movie Tracker";

export const getSiteUrl = () => {
  return (process.env.NEXT_PUBLIC_SITE_URL ?? "https://movietracker.ru").replace(/\/$/, "");
};

export const getAbsoluteUrl = (path: string) => {
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${getSiteUrl()}${normalizedPath}`;
};

const formatDate = (date?: Date) => date?.toISOString().slice(0, 10);

const omitEmpty = (data: JsonLd): JsonLd =>
  Object.fromEntries(
    Object.entries(data).filter(([, value]) => {
      if (value === undefined || value === null || value === "") return false;
      if (Array.isArray(value) && value.length === 0) return false;

      return true;
    }),
  );

export const createWebsiteJsonLd = (): JsonLd => {
  const siteUrl = getSiteUrl();

  return {
    "@context": "https://schema.org",
    "@type": "WebSite",
    "@id": `${siteUrl}/#website`,
    name: SITE_NAME,
    url: siteUrl,
    inLanguage: "ru",
    potentialAction: {
      "@type": "SearchAction",
      target: `${siteUrl}/movie?q={search_term_string}`,
      "query-input": "required name=search_term_string",
    },
  };
};

export const createBreadcrumbJsonLd = (items: { name: string; path: string }[]): JsonLd => ({
  "@context": "https://schema.org",
  "@type": "BreadcrumbList",
  itemListElement: items.map((item, index) => ({
    "@type": "ListItem",
    position: index + 1,
    name: item.name,
    item: getAbsoluteUrl(item.path),
  })),
});

export const createCollectionPageJsonLd = ({
  name,
  description,
  path,
}: {
  name: string;
  description: string;
  path: string;
}): JsonLd => {
  const url = getAbsoluteUrl(path);

  return {
    "@context": "https://schema.org",
    "@type": "CollectionPage",
    "@id": `${url}#webpage`,
    url,
    name,
    description,
    inLanguage: "ru",
    isPartOf: {
      "@id": `${getSiteUrl()}/#website`,
    },
  };
};

export const createItemListJsonLd = ({
  name,
  items,
  type,
  pathPrefix,
}: {
  name: string;
  items: {
    id: number;
    title: string;
    overview?: string;
    releaseDate?: Date;
    posterPath?: string;
  }[];
  type: "Movie" | "TVSeries";
  pathPrefix: "/movie" | "/tv";
}): JsonLd => ({
  "@context": "https://schema.org",
  "@type": "ItemList",
  name,
  itemListElement: items.map((item, index) => {
    const url = getAbsoluteUrl(`${pathPrefix}/${item.id}`);

    return {
      "@type": "ListItem",
      position: index + 1,
      item: omitEmpty({
        "@type": type,
        "@id": `${url}#${type.toLowerCase()}`,
        url,
        name: item.title,
        description: item.overview,
        image: item.posterPath,
        datePublished: formatDate(item.releaseDate),
      }),
    };
  }),
});

export const createMovieJsonLd = ({
  id,
  title,
  description,
  releaseDate,
  posterPath,
  genres,
  originalLanguage,
}: {
  id: number;
  title: string;
  description?: string;
  releaseDate?: Date;
  posterPath?: string;
  genres: string[];
  originalLanguage?: string;
}): JsonLd => {
  const url = getAbsoluteUrl(`/movie/${id}`);

  return omitEmpty({
    "@context": "https://schema.org",
    "@type": "Movie",
    "@id": `${url}#movie`,
    url,
    name: title,
    description,
    image: posterPath,
    datePublished: formatDate(releaseDate),
    genre: genres,
    inLanguage: originalLanguage,
    isPartOf: {
      "@id": `${getSiteUrl()}/#website`,
    },
  });
};

export const createTvSeriesJsonLd = ({
  id,
  title,
  description,
  releaseDate,
  posterPath,
  genres,
  originalLanguage,
  numberOfSeasons,
  numberOfEpisodes,
}: {
  id: number;
  title: string;
  description?: string;
  releaseDate?: Date;
  posterPath?: string;
  genres: string[];
  originalLanguage?: string;
  numberOfSeasons?: number;
  numberOfEpisodes?: number;
}): JsonLd => {
  const url = getAbsoluteUrl(`/tv/${id}`);

  return omitEmpty({
    "@context": "https://schema.org",
    "@type": "TVSeries",
    "@id": `${url}#tvseries`,
    url,
    name: title,
    description,
    image: posterPath,
    datePublished: formatDate(releaseDate),
    genre: genres,
    inLanguage: originalLanguage,
    numberOfSeasons,
    numberOfEpisodes,
    isPartOf: {
      "@id": `${getSiteUrl()}/#website`,
    },
  });
};

export const stringifyJsonLd = (data: JsonLd | JsonLd[]) => JSON.stringify(data).replace(/</g, "\\u003c");

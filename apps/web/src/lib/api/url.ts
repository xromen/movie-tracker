const API_URL = process.env.MOVIE_TRACKER_API_URL ?? "http://localhost:8080/api";

export const getApiUrl = (path: string) => {
  const normalizedBaseUrl = API_URL.replace(/\/$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;

  return `${normalizedBaseUrl}${normalizedPath}`;
};

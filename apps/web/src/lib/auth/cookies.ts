export const ACCESS_TOKEN_COOKIE =
  process.env.MOVIE_TRACKER_ACCESS_TOKEN_COOKIE ??
  process.env.NEXT_PUBLIC_ACCESS_TOKEN_COOKIE ??
  "access_token";

export const REFRESH_TOKEN_COOKIE =
  process.env.MOVIE_TRACKER_REFRESH_TOKEN_COOKIE ??
  process.env.NEXT_PUBLIC_REFRESH_TOKEN_COOKIE ??
  "refresh_token";

export const LEGACY_ACCESS_TOKEN_COOKIE = "movie_tracker_access_token";
export const LEGACY_REFRESH_TOKEN_COOKIE = "movie_tracker_refresh_token";

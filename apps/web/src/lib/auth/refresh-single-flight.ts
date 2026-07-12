const REFRESH_RESULT_CACHE_MS = 5_000;

interface RefreshFlight {
  expiresAt?: number;
  promise: Promise<string[]>;
}

const refreshFlights = new Map<string, RefreshFlight>();

export const getSingleFlightRefreshCookies = (
  refreshToken: string,
  refresh: () => Promise<string[]>,
) => {
  const now = Date.now();
  const currentFlight = refreshFlights.get(refreshToken);

  if (currentFlight) {
    if (currentFlight.expiresAt === undefined || currentFlight.expiresAt > now) {
      return currentFlight.promise;
    }

    refreshFlights.delete(refreshToken);
  }

  const promise = refresh();

  refreshFlights.set(refreshToken, {
    promise,
  });

  promise
    .then((cookies) => {
      const flight = refreshFlights.get(refreshToken);

      if (flight?.promise !== promise) {
        return;
      }

      if (cookies.length === 0) {
        refreshFlights.delete(refreshToken);
        return;
      }

      flight.expiresAt = Date.now() + REFRESH_RESULT_CACHE_MS;

      setTimeout(() => {
        if (refreshFlights.get(refreshToken)?.promise === promise) {
          refreshFlights.delete(refreshToken);
        }
      }, REFRESH_RESULT_CACHE_MS);
    })
    .catch(() => {
      if (refreshFlights.get(refreshToken)?.promise === promise) {
        refreshFlights.delete(refreshToken);
      }
    });

  return promise;
};

import { cookies } from "next/headers.js";
import { ACCESS_TOKEN_COOKIE, LEGACY_ACCESS_TOKEN_COOKIE } from "./cookies";
import { getRolesFromToken, getUsernameFromToken } from "./jwt";

export const getSession = async () => {
  const cookieStore = await cookies();
  const accessToken =
    cookieStore.get(ACCESS_TOKEN_COOKIE)?.value ??
    cookieStore.get(LEGACY_ACCESS_TOKEN_COOKIE)?.value;

  if (!accessToken) {
    return {
      isAuthenticated: false,
      username: "",
      roles: [] as string[],
    };
  }

  return {
    isAuthenticated: true,
    username: getUsernameFromToken(accessToken),
    roles: getRolesFromToken(accessToken),
  };
};

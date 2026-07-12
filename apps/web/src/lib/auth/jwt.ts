export interface JwtPayload {
  exp?: number;
  username?: string;
  roles?: string[];
}

const decodeBase64Url = (value: string) => {
  const base64 = value.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(base64.length + ((4 - (base64.length % 4)) % 4), "=");

  return Buffer.from(padded, "base64").toString("utf8");
};

export const parseJwt = (token: string): JwtPayload | null => {
  try {
    const payload = token.split(".")[1];

    if (!payload) {
      return null;
    }

    return JSON.parse(decodeBase64Url(payload)) as JwtPayload;
  } catch {
    return null;
  }
};

export const getUsernameFromToken = (token: string) => parseJwt(token)?.username ?? "";

export const getRolesFromToken = (token: string) => parseJwt(token)?.roles ?? [];

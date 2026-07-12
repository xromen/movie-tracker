import { fetchApi } from "./client";
import type { LoginRequest, RegisterRequest } from "./types";

const submitAuth = <TBody>(path: string, body?: TBody) =>
  fetchApi(path, {
    method: "POST",
    body,
  });

export const login = (data: LoginRequest) => submitAuth("/v1/auth/login", data);

export const register = (data: RegisterRequest) => submitAuth("/v1/auth/register", data);

export const refresh = () => submitAuth("/v1/auth/refresh");

export const logout = () => submitAuth("/v1/auth/logout");

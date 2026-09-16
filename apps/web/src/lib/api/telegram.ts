import { fetchApi } from "./client";

export interface BindingUrl {
    Url: string,
    ExpiresAt: Date,
    CreatedAt: Date
}

export interface BindingUrlReponse {
    url: string,
    expires_at: Date,
    created_at: Date
}

export const getBindingUrl = async (): Promise<BindingUrl> => {
  const data = await fetchApi<BindingUrlReponse>("/v1/telegram/binding-url")

  return {
    Url: data.url,
    ExpiresAt: data.expires_at,
    CreatedAt: data.created_at
  }
};

import { clearToken, getToken } from "./auth";

const DEFAULT_API_URL = "http://localhost:8080";

export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_URL ?? DEFAULT_API_URL;
}

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public body?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export type ApiFetchOptions = {
  method?: string;
  body?: unknown;
  token?: string | null;
  idempotencyKey?: string;
  headers?: Record<string, string>;
};

export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const {
    method = "GET",
    body,
    token = getToken(),
    idempotencyKey,
    headers: extraHeaders = {},
  } = options;

  const headers: Record<string, string> = {
    Accept: "application/json",
    ...extraHeaders,
  };

  if (body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  if (idempotencyKey) {
    headers["Idempotency-Key"] = idempotencyKey;
  }

  const res = await fetch(`${getApiBaseUrl()}${path}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    clearToken();
  }

  const text = await res.text();
  let parsed: unknown;
  try {
    parsed = text ? JSON.parse(text) : undefined;
  } catch {
    parsed = text;
  }

  if (!res.ok) {
    let message = `Error HTTP ${res.status}`;
    if (typeof parsed === "object" && parsed !== null) {
      const obj = parsed as Record<string, unknown>;
      if (typeof obj.error === "string") {
        message = obj.error;
      } else if (typeof obj.message === "string") {
        message = obj.message;
      }
    }
    throw new ApiError(message, res.status, parsed);
  }

  return parsed as T;
}

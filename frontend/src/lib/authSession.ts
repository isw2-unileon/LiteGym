import { API_BASE_URL } from "./api";

const AUTH_TOKEN_STORAGE_KEY = "litegym_auth_token";

let isFetchInterceptorInstalled = false;
let memoryAuthToken: string | null = null;

function getRequestUrl(input: RequestInfo | URL) {
  if (typeof input === "string") {
    return input;
  }

  if (input instanceof URL) {
    return input.toString();
  }

  return input.url;
}

function shouldAttachAuthHeader(requestUrl: string) {
  return requestUrl.startsWith("/api/") || requestUrl.startsWith("/health") || (
    API_BASE_URL !== "" && requestUrl.startsWith(API_BASE_URL)
  );
}

function buildAuthorizedRequest(input: RequestInfo | URL, init?: RequestInit) {
  const headers = new Headers(init?.headers);
  if (!headers.has("Authorization")) {
    const token = getAuthToken();
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  const requestInit: RequestInit = {
    ...init,
    headers,
  };

  if (input instanceof Request) {
    return new Request(input, requestInit);
  }

  return { input, requestInit };
}

export function getAuthToken() {
  try {
    return window.localStorage.getItem(AUTH_TOKEN_STORAGE_KEY);
  } catch {
    return memoryAuthToken;
  }
}

export function setAuthToken(token: string) {
  memoryAuthToken = token;

  try {
    window.localStorage.setItem(AUTH_TOKEN_STORAGE_KEY, token);
  } catch {
    // Ignore storage failures and keep the token in memory for the current tab.
  }
}

export function clearAuthToken() {
  memoryAuthToken = null;

  try {
    window.localStorage.removeItem(AUTH_TOKEN_STORAGE_KEY);
  } catch {
    // Ignore storage failures; the in-memory token is already cleared.
  }
}

export function installAuthFetchInterceptor() {
  if (isFetchInterceptorInstalled) {
    return;
  }

  isFetchInterceptorInstalled = true;

  const originalFetch = window.fetch.bind(window);

  window.fetch = ((input: RequestInfo | URL, init?: RequestInit) => {
    const requestUrl = getRequestUrl(input);
    if (!shouldAttachAuthHeader(requestUrl)) {
      return originalFetch(input, init);
    }

    const authorizedRequest = buildAuthorizedRequest(input, init);
    if (authorizedRequest instanceof Request) {
      return originalFetch(authorizedRequest);
    }

    return originalFetch(authorizedRequest.input, authorizedRequest.requestInit);
  }) as typeof window.fetch;
}

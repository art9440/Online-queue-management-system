import axios from "axios";

const createApi = (baseURL) =>
  axios.create({
    baseURL,
    headers: { "Content-Type": "application/json" },
    withCredentials: true,
  });

// Runtime config (public/config.js, generated from env in the container) wins over
// build-time VITE_* variables, which win over local defaults.
const runtimeConfig = globalThis.__APP_CONFIG__ ?? {};

const apiUrl = (key, fallback) =>
  runtimeConfig[key] || import.meta.env[`VITE_${key}`] || fallback;

export const registrationApi = createApi(apiUrl("REGISTRATION_URL", "http://localhost:8081"));

export const authorizationApi = createApi(apiUrl("AUTH_URL", "http://localhost:8082"));

export const branchesApi = createApi(apiUrl("BRANCHES_URL", "http://localhost:8083"));

export const bookingApi = createApi(apiUrl("BOOKING_URL", "http://localhost:8084"));

let refreshRequest = null;

const refreshAccessToken = () => {
  if (!refreshRequest) {
    refreshRequest = authorizationApi
      .post("/auth/refresh", null, { skipAuthRefresh: true })
      .finally(() => {
        refreshRequest = null;
      });
  }

  return refreshRequest;
};

const notifyUnauthorized = () => {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event("auth:unauthorized"));
  }
};

const shouldRefreshToken = (error) => {
  const { config, response } = error;
  const requestUrl = config?.url || "";
  const isAuthAction =
    requestUrl.includes("/auth/login") ||
    requestUrl.includes("/auth/refresh") ||
    requestUrl.includes("/auth/logout");

  return (
    response?.status === 401 &&
    config &&
    !config._retry &&
    !config.skipAuthRefresh &&
    !isAuthAction
  );
};

const addRefreshInterceptor = (api) => {
  api.interceptors.response.use(
    (response) => response,
    async (error) => {
      if (!shouldRefreshToken(error)) {
        return Promise.reject(error);
      }

      const originalRequest = error.config;
      originalRequest._retry = true;

      try {
        await refreshAccessToken();
        return api(originalRequest);
      } catch (refreshError) {
        notifyUnauthorized();
        return Promise.reject(refreshError);
      }
    }
  );
};

addRefreshInterceptor(authorizationApi);
addRefreshInterceptor(branchesApi);
addRefreshInterceptor(bookingApi);

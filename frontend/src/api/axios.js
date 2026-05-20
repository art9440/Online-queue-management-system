import axios from "axios";

const createApi = (baseURL) =>
  axios.create({
    baseURL,
    headers: { "Content-Type": "application/json" },
    withCredentials: true,
  });

export const registrationApi = createApi(
  import.meta.env.VITE_REGISTRATION_URL || "http://localhost:8081"
);

export const authorizationApi = createApi(
  import.meta.env.VITE_AUTH_URL || "http://localhost:8082"
);

export const branchesApi = createApi(
  import.meta.env.VITE_BRANCHES_URL || "http://localhost:8083"
);

export const bookingApi = createApi(
  import.meta.env.VITE_BOOKING_URL || "http://localhost:8084"
);

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

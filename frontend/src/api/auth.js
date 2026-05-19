import { authorizationApi } from "./axios";

export const authApi = {
    login: (login, password) =>
        authorizationApi.post('/auth/login', { login, password }),
    me: () =>
        authorizationApi.get('/auth/me'),
    refresh: () =>
        authorizationApi.post('/auth/refresh'),
    logout: () =>
        authorizationApi.post('/auth/logout'),
};
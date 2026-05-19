import axios from "axios";

export const registrationApi = axios.create({
    baseURL: 'http://localhost:8081',
    headers: { 'Content-Type':'application/json' },
    withCredentials: true
});

export const authorizationApi = axios.create({
    baseURL: 'http://localhost:8082',
    headers: { 'Content-Type':'application/json' },
    withCredentials: true
});

export const branchesApi = axios.create({
    baseURL: 'http://localhost:8083',
    headers: { 'Content-Type':'application/json' },
});

export const bookingApi = axios.create({
    baseURL: 'http://localhost:8084',
    headers: { 'Content-Type':'application/json' },
});

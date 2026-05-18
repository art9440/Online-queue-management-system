<<<<<<< HEAD
/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext } from "react";

const AuthContext = createContext({
    user: null,
    loading: false,
    login: () => {},
    logout: () => {},
    getRedirectPath: () => "/register",
});

export const AuthProvider = ({ children }) => {
    return (
        <AuthContext.Provider
            value={{
                user: null,
                loading: false,
                login: () => {},
                logout: () => {},
                getRedirectPath: () => "/register",
            }}
        >
=======
import { createContext, useContext, useEffect, useState } from "react";
import { authApi } from "../api/auth";
import { getRoleName } from "../constants/roles";

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const getRedirectPath = (userData) => {
        if (!userData) return '/login';

        const roleName = getRoleName(userData.role_id);
        if (roleName === 'super_admin' || roleName === 'business_admin') {
            return '/admin';
        }
        if (roleName === 'manager' && userData.branch_id) {
            return `/admin/branch/${userData.branch_id}`;
        }
        if (roleName === 'employee') {
            return '/admin/schedule';
        }
        return '/admin';
    };
    
    useEffect(() => {
        checkAuth();
    }, []);

    const checkAuth = async () => {
        try {
            const response = await authApi.me();
            setUser(response.data);
            setError(null);
        } catch (err) {
            if (err.response?.status !== 401) {
                setError('Ошибка проверки авторизации');
            }
            setUser(null);
        } finally {
            setLoading(false);
        }
    };

    const login = async(login, password) => {
        setError(null);
        try {
            await authApi.login(login, password);
            const response = await authApi.me();
            setUser(response.data);

            return response.data;
        } catch (err) {
            const msg = err?.response?.data?.message
            || err?.response?.data?.error
            || 'Ошибка входа';
            setError(msg);
            throw new Error(msg);
        }
    };

    const logout = async () => {
        try {
            await authApi.logout();
        } catch (err) {
            console.error('Ошибка выхода: ', err);
        }
        setUser(null);
        setError(null);
    };

    const value = {
        user,
        loading,
        error,
        login,
        logout,
        isAuthenticated: Boolean(user),
        getRedirectPath,
    };

    return (
        <AuthContext.Provider value={value}>
>>>>>>> 2299878 (add AuthContext)
            {children}
        </AuthContext.Provider>
    );
};

<<<<<<< HEAD
export const useAuth = () => useContext(AuthContext);
=======
export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) throw new Error('useAuth должен использоваться внутри AuthProvider');
    return context;
};
>>>>>>> 2299878 (add AuthContext)

/* eslint-disable react-refresh/only-export-components */
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { authApi } from "../api/auth";
import { getRoleName } from "../constants/roles";

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    const getRedirectPath = useCallback((userData) => {
        if (!userData) return '/login';

        const roleName = getRoleName(userData.role_id);
        if (roleName === 'super_admin' || roleName === 'business_admin') {
            return '/admin';
        }
        if (roleName === 'manager' && userData.branch_id) {
            return '/manager';
        }
        if (roleName === 'employee') {
            return '/admin/schedule';
        }
        return '/admin';
    }, []);
    
    const checkAuth = useCallback(async () => {
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
    }, []);

    useEffect(() => {
        checkAuth();
    }, [checkAuth]);

    useEffect(() => {
        const handleUnauthorized = () => {
            setUser(null);
            setError('Сессия истекла. Войдите снова.');
        };

        window.addEventListener('auth:unauthorized', handleUnauthorized);

        return () => {
            window.removeEventListener('auth:unauthorized', handleUnauthorized);
        };
    }, []);

    const login = useCallback(async(login, password) => {
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
    }, []);

    const logout = useCallback(async () => {
        try {
            await authApi.logout();
        } catch (err) {
            console.error('Ошибка выхода: ', err);
        }
        setUser(null);
        setError(null);
    }, []);

    const value = useMemo(() => ({
        user,
        loading,
        error,
        login,
        logout,
        isAuthenticated: Boolean(user),
        getRedirectPath,
    }), [error, getRedirectPath, loading, login, logout, user]);

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) throw new Error('useAuth должен использоваться внутри AuthProvider');
    return context;
};

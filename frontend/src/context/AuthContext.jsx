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
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);

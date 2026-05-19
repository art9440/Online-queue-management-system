import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { hasRole, getRoleName } from "../constants/roles"

export const ProtectedRoute = ({ children, allowedRoles = []}) => {
    const { user, loading, getRedirectPath } = useAuth();
    const location = useLocation();

    if (loading) {
        return (
            <div className="min-h-screen flex items-center justify-center bg-gray-50">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600" />
            </div>
        )
    }

    if (!user){
        return <Navigate to="/login" state={{from: location}} replace/>
    }

    if (allowedRoles.length > 0 && user.role_id && !hasRole(user.role_id, allowedRoles)){
        return <Navigate to={getRedirectPath(user)} replace />;
    }

    if (location.pathname.match(/\/admin\/branch\/(\d+)/)) {
        const branchId = parseInt(location.pathname.match(/\/admin\/branch\/(\d+)/)[1]);
        const roleName = getRoleName(user.role_id);
        
        if (roleName === 'manager' && user.branch_id && branchId !== user.branch_id) {
            return <Navigate to={`/admin/branch/${user.branch_id}`} replace />;
        }
    }

    return children;
};

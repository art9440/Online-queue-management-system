import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./context/AuthContext";
import { ProtectedRoute } from "./components/ProtectedRoute";
import { AdminPage } from "./pages/AdminPage";
import { ClientsPage } from "./pages/ClientsPage";
import { EmployeesPage } from "./pages/EmployeesPage";
import { LoginPage } from "./pages/LoginPage";
import { ManagerPage } from "./pages/ManagerPage";
import { PublicBookingPage } from "./pages/PublicBookingPage";
import { RegistrationPage } from "./pages/RegistrationPage";
import { ServicesPage } from "./pages/ServicesPage";
import { SettingsPage } from "./pages/SettingsPage";
import { VerifyPage } from "./pages/VerifyPage";

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/register" element={<RegistrationPage />} />
          <Route path="/verify" element={<VerifyPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/public/:registrationSlug" element={<PublicBookingPage />} />

          <Route
            path="/admin"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <AdminPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/branch/:branchId"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <ManagerPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/clients"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <ClientsPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/services"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <ServicesPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/employees"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <EmployeesPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/admin/settings"
            element={
              <ProtectedRoute allowedRoles={["super_admin", "business_admin"]}>
                <SettingsPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/manager"
            element={
              <ProtectedRoute allowedRoles={["manager"]}>
                <ManagerPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/manager/clients"
            element={
              <ProtectedRoute allowedRoles={["manager"]}>
                <ClientsPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/manager/services"
            element={
              <ProtectedRoute allowedRoles={["manager"]}>
                <ServicesPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/manager/employees"
            element={
              <ProtectedRoute allowedRoles={["manager"]}>
                <EmployeesPage />
              </ProtectedRoute>
            }
          />

          <Route
            path="/manager/settings"
            element={
              <ProtectedRoute allowedRoles={["manager"]}>
                <SettingsPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;

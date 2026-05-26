import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Scissors, Search } from "lucide-react";
import { getServices } from "../api/branches";
import { ServiceCard } from "../components/dashboard/ServiceCard";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { Sidebar } from "../components/layouts/Sidebar";
import { useAuth } from "../context/AuthContext";
import { groupServicesForDisplay } from "../mocks/serviceGroups";

const getIsManagerPath = () => globalThis.window.location.pathname.startsWith("/manager");

const normalizeService = (service) => ({
  id: service.id,
  name: service.name,
  duration_minutes: service.duration_minutes,
  price: service.price,
});

export const ServicesPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [search, setSearch] = useState("");
  const { user, logout } = useAuth();
  const isManagerPath = getIsManagerPath();

  const { data: apiServices = [], isLoading, isError } = useQuery({
    queryKey: ["services", isManagerPath ? "manager" : "admin"],
    queryFn: getServices,
    retry: 1,
  });

  const services = useMemo(() => {
    const source = Array.isArray(apiServices) ? apiServices : [];
    return groupServicesForDisplay(source.map(normalizeService));
  }, [apiServices]);

  const visibleServices = services.filter((service) => {
    const text = `${service.name}`;
    return text.toLowerCase().includes(search.toLowerCase());
  });

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(99,102,241,0.14),transparent_32rem),linear-gradient(135deg,#f8fbff_0%,#eef6ff_48%,#f7fbf4_100%)]">
      <Sidebar
        role={isManagerPath ? "manager" : "admin"}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onLogout={logout}
      />

      <DashboardTopBar
        user={user}
        title="Услуги"
        onMenuClick={() => setSidebarOpen(true)}
        onLogout={logout}
      />

      <main className="p-4 lg:p-6">
        <div className="mx-auto max-w-7xl space-y-5">
          {isError && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Не удалось загрузить услуги из API. Проверь branches service и
              авторизацию через cookie access_token.
            </div>
          )}

          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Услуги</h2>
              <p className="mt-1 text-sm text-gray-500">
                {isManagerPath
                  ? "Услуги вашего филиала"
                  : "Услуги по всем доступным филиалам"}
              </p>
            </div>

            <label className="flex w-full items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 sm:max-w-xs">
              <Search size={16} className="text-gray-400" />
              <input
                type="search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Поиск услуги"
                className="w-full bg-transparent text-sm outline-none"
              />
            </label>
          </div>

          {isLoading && (
            <div className="rounded-xl bg-white p-5 text-sm text-gray-500 shadow-sm">
              Загружаем услуги...
            </div>
          )}

          {!isLoading && visibleServices.length === 0 && (
            <div className="rounded-xl bg-white p-8 text-center shadow-sm">
              <Scissors size={36} className="mx-auto text-gray-300" />
              <p className="mt-3 font-medium text-gray-900">
                Услуги не найдены
              </p>
              <p className="mt-1 text-sm text-gray-500">
                API вернул пустой список услуг.
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {visibleServices.map((service) => (
              <ServiceCard key={service.id} service={service} />
            ))}
          </div>
        </div>
      </main>
    </div>
  );
};

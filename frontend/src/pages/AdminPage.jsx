import { useMemo, useState } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import {
  CalendarDays,
  ExternalLink,
  Link as LinkIcon,
  MapPin,
  Store,
  Users,
} from "lucide-react";
import {
  getBranchBookings,
  getBranchEmployees,
  getBranches,
  getServices,
} from "../api/branches";
import { DashboardStats } from "../components/dashboard/DashboardStats";
import { TodayInfoCard } from "../components/dashboard/TodayInfoCard";
import { Sidebar } from "../components/layouts/Sidebar";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { useAuth } from "../context/AuthContext";

const normalizeBranch = (branch) => ({
  id: branch.id,
  name: branch.name,
  address: branch.address,
});

const getTodayKey = () => new Date().toISOString().slice(0, 10);

const normalizeEmployee = (employee, branch) => ({
  id: employee.id,
  branch_id: branch.id,
  name: employee.name,
  surname: employee.surname,
  position: employee.position,
});

const getServicePriceMap = (services) =>
  new Map(
    services.map((service) => [
      Number(service.id),
      Number(service.price || 0),
    ])
  );

const getBookingPrice = (booking, servicePrices) => {
  return Number(
    booking.price || servicePrices.get(Number(booking.service_id)) || 0
  );
};

const BranchOverviewCard = ({
  branch,
  employees = [],
  bookings = [],
  servicePrices,
  onOpen,
}) => {
  const todayBookings = bookings;
  const registrationLink =
    typeof window === "undefined"
      ? `/register?branch_id=${branch.id}`
      : `${window.location.origin}/register?branch_id=${branch.id}`;
  const revenue = todayBookings.reduce(
    (sum, booking) => sum + getBookingPrice(booking, servicePrices),
    0
  );

  return (
    <article
      onClick={onOpen}
      className="group cursor-pointer overflow-hidden rounded-2xl border border-indigo-100 bg-white p-5 shadow-lg shadow-indigo-100/40 transition-all hover:-translate-y-0.5 hover:border-indigo-300 hover:shadow-xl hover:shadow-indigo-200/70"
    >
      <div className="-mx-5 -mt-5 mb-5 h-2 bg-indigo-600" />
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="flex h-9 w-9 items-center justify-center rounded-xl bg-indigo-50 text-indigo-600 ring-1 ring-indigo-100">
              <Store size={18} />
            </span>
            <h3 className="font-semibold text-slate-950 truncate group-hover:text-indigo-700">
              {branch.name}
            </h3>
          </div>
          <p className="mt-2 flex items-center gap-1.5 text-sm text-gray-500">
            <MapPin size={14} />
            <span className="truncate">{branch.address}</span>
          </p>
        </div>
      </div>

      <div className="mt-5 grid grid-cols-3 gap-3">
        <div className="rounded-xl bg-blue-50 p-3 text-center ring-1 ring-blue-100">
          <CalendarDays size={16} className="mx-auto text-gray-400" />
          <p className="mt-1 text-lg font-semibold text-blue-950">
            {todayBookings.length}
          </p>
          <p className="text-xs text-blue-600">записей</p>
        </div>
        <div className="rounded-xl bg-emerald-50 p-3 text-center ring-1 ring-emerald-100">
          <Users size={16} className="mx-auto text-gray-400" />
          <p className="mt-1 text-lg font-semibold text-emerald-950">
            {employees.length}
          </p>
          <p className="text-xs text-emerald-700">сотрудников</p>
        </div>
        <div className="rounded-xl bg-amber-50 p-3 text-center ring-1 ring-amber-100">
          <p className="text-lg font-semibold text-amber-950">
            {revenue ? `${Math.round(revenue / 1000)}K` : "0"}
          </p>
          <p className="text-xs text-amber-700">выручка</p>
        </div>
      </div>

      <div className="mt-4 border-t border-gray-100 pt-4">
        <p className="mb-2 flex items-center gap-1.5 text-xs font-medium text-gray-500">
          <LinkIcon size={13} />
          Ссылка для регистрации сотрудников
        </p>
        <a
          href={registrationLink}
          onClick={(event) => event.stopPropagation()}
          className="flex items-center justify-between gap-3 rounded-xl bg-indigo-50 px-3 py-2 text-sm font-medium text-indigo-700 ring-1 ring-indigo-100 hover:bg-indigo-100"
        >
          <span className="truncate">{registrationLink}</span>
          <ExternalLink size={14} className="shrink-0" />
        </a>
      </div>
    </article>
  );
};

export const AdminPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const todayKey = getTodayKey();

  const {
    data: apiBranches,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["branches"],
    queryFn: getBranches,
    retry: 1,
  });

  const branches = useMemo(() => {
    const source = Array.isArray(apiBranches) ? apiBranches : [];
    return source.map(normalizeBranch);
  }, [apiBranches]);

  const { data: apiServices = [] } = useQuery({
    queryKey: ["services", "admin"],
    queryFn: getServices,
    retry: 1,
  });

  const services = useMemo(
    () => (Array.isArray(apiServices) ? apiServices : []),
    [apiServices]
  );
  const servicePrices = useMemo(() => getServicePriceMap(services), [services]);

  const employeeQueries = useQueries({
    queries: branches.map((branch) => ({
      queryKey: ["branches", branch.id, "employees", "admin-dashboard"],
      queryFn: () => getBranchEmployees(branch.id),
      enabled: Boolean(branch.id),
      retry: 1,
    })),
  });

  const bookingQueries = useQueries({
    queries: branches.map((branch) => ({
      queryKey: ["branches", branch.id, "bookings", todayKey, "admin-dashboard"],
      queryFn: () => getBranchBookings(branch.id, todayKey),
      enabled: Boolean(branch.id),
      retry: 1,
    })),
  });

  const branchData = useMemo(() => {
    const result = new Map();

    branches.forEach((branch, index) => {
      const employees = (employeeQueries[index]?.data || []).map((employee) =>
        normalizeEmployee(employee, branch)
      );
      const bookings = bookingQueries[index]?.data || [];

      result.set(branch.id, { employees, bookings });
    });

    return result;
  }, [branches, employeeQueries, bookingQueries]);

  const employees = Array.from(branchData.values()).flatMap(
    (data) => data.employees
  );
  const todayBookings = Array.from(branchData.values()).flatMap(
    (data) => data.bookings
  );
  const todayRevenue = todayBookings.reduce(
    (sum, booking) => sum + getBookingPrice(booking, servicePrices),
    0
  );
  const isDetailsError =
    employeeQueries.some((query) => query.isError) ||
    bookingQueries.some((query) => query.isError);

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(99,102,241,0.14),transparent_32rem),linear-gradient(135deg,#f8fbff_0%,#eef6ff_48%,#f7fbf4_100%)]">
      <Sidebar
        role="admin"
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onLogout={logout}
      />

      <div className="flex min-w-0 flex-1 flex-col">
        <DashboardTopBar
          user={user}
          title="Филиалы"
          onMenuClick={() => setSidebarOpen(true)}
          onLogout={logout}
        />

        <main className="flex-1 overflow-auto p-4 lg:p-6">
          <div className="mx-auto max-w-7xl space-y-5">
            {isError && (
              <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                Не удалось загрузить филиалы из API. Проверь, что branches
                service запущен на порту 8083 и пользователь авторизован через
                cookie access_token.
              </div>
            )}

            {!isError && isDetailsError && (
              <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
                Филиалы загрузились, но часть дополнительных данных
                (сотрудники или записи за сегодня) пока недоступна.
              </div>
            )}

            <div className="grid gap-4 xl:grid-cols-[1fr_1.35fr]">
              <TodayInfoCard
                branchName="Все филиалы"
                bookingsCount={todayBookings.length}
                revenue={todayRevenue}
                newClients={0}
                utilization={null}
              />

              <DashboardStats
                bookings={todayBookings}
                employees={employees}
                services={services}
                servicePrices={servicePrices}
              />
            </div>

            <div className="flex items-end justify-between gap-4">
              <div>
                <h2 className="text-xl font-semibold text-gray-900">Филиалы</h2>
                <p className="mt-1 text-sm text-gray-500">
                  Нажмите на филиал, чтобы открыть расписание сотрудников.
                </p>
              </div>

              {isLoading && (
                <span className="text-sm text-gray-400">Загрузка филиалов...</span>
              )}
            </div>

            {!isLoading && !isError && branches.length === 0 && (
              <div className="rounded-xl bg-white p-8 text-center shadow-sm">
                <Store size={36} className="mx-auto text-gray-300" />
                <p className="mt-3 font-medium text-gray-900">
                  Филиалы не найдены
                </p>
                <p className="mt-1 text-sm text-gray-500">
                  API вернул пустой список. Моки здесь специально не
                  подставляются, чтобы не скрывать проблему с данными.
                </p>
              </div>
            )}

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
              {branches.map((branch) => (
                <BranchOverviewCard
                  key={branch.id}
                  branch={branch}
                  employees={branchData.get(branch.id)?.employees || []}
                  bookings={branchData.get(branch.id)?.bookings || []}
                  servicePrices={servicePrices}
                  onOpen={() => navigate(`/admin/branch/${branch.id}`)}
                />
              ))}
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};

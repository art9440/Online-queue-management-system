import { useMemo, useState } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import { Search, Users } from "lucide-react";
import {
  getBranchBookings,
  getBranchClients,
  getBranches,
} from "../api/branches";
import { ClientCard } from "../components/dashboard/ClientCard";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { Sidebar } from "../components/layouts/Sidebar";
import { useAuth } from "../context/AuthContext";
import { clientMatchesSearch } from "../mocks/clientSearch";

const normalizeBranch = (branch) => ({
  id: branch.id,
  name: branch.name,
  address: branch.address,
});

const getIsManagerPath = () => globalThis.window.location.pathname.startsWith("/manager");
const getTodayKey = () => new Date().toISOString().slice(0, 10);

const normalizeClient = (client, branch) => ({
  id: client.id,
  branch_id: branch.id,
  branchName: branch.name,
  email: client.email,
  phone: client.phone,
  name: client.name,
  surname: client.surname,
  tg_username: client.tg_username,
  created_at: client.created_at,
});

const getClientName = (client) =>
  [client?.surname, client?.name].filter(Boolean).join(" ") || client?.name || "";

const normalizeBooking = (booking, branch) => ({
  id: booking.id,
  branch_id: booking.branch_id,
  branchName: branch.name,
  client_id: booking.client?.id,
  client_name: getClientName(booking.client),
  employee_id: booking.employee_id,
  employeeName: [booking.employee_surname, booking.employee_name]
    .filter(Boolean)
    .join(" "),
  service_id: booking.service_id,
  service_name: booking.service_name,
  start_time: booking.start_time,
  end_time: booking.end_time,
  status:
    booking.status === "cancelled"
      ? "canceled"
      : booking.status || "confirmed",
  comment: booking.comment,
});

export const ClientsPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [selectedDate, setSelectedDate] = useState(getTodayKey());
  const { user, logout } = useAuth();
  const isManagerPath = getIsManagerPath();

  const { data: apiBranches = [], isLoading, isError } = useQuery({
    queryKey: ["branches", isManagerPath ? "manager" : "admin", "clients"],
    queryFn: getBranches,
    retry: 1,
  });

  const branches = useMemo(() => {
    const source = Array.isArray(apiBranches) ? apiBranches : [];
    return source.map(normalizeBranch);
  }, [apiBranches]);

  const clientQueries = useQueries({
    queries: branches.map((branch) => ({
      queryKey: ["branches", branch.id, "clients"],
      queryFn: () => getBranchClients(branch.id),
      enabled: Boolean(branch.id),
      retry: 1,
    })),
  });

  const bookingQueries = useQueries({
    queries: branches.map((branch) => ({
      queryKey: ["branches", branch.id, "client-bookings", selectedDate],
      queryFn: () => getBranchBookings(branch.id, selectedDate),
      enabled: Boolean(branch.id && selectedDate),
      retry: 1,
    })),
  });

  const clients = useMemo(() => {
    const clientMap = new Map();

    branches.forEach((branch, index) => {
      const source = clientQueries[index]?.data || [];
      source.forEach((client) => {
        const normalizedClient = normalizeClient(client, branch);
        clientMap.set(normalizedClient.id, {
          ...clientMap.get(normalizedClient.id),
          ...normalizedClient,
        });
      });
    });

    return Array.from(clientMap.values());
  }, [branches, clientQueries]);

  const bookings = useMemo(() => {
    return branches.flatMap((branch, index) => {
      const source = bookingQueries[index]?.data || [];
      return source.map((booking) => normalizeBooking(booking, branch));
    });
  }, [branches, bookingQueries]);

  const bookingsByClient = useMemo(() => {
    const result = new Map();

    bookings.forEach((booking) => {
      if (!booking.client_id) return;

      const clientBookings = result.get(booking.client_id) || [];
      result.set(booking.client_id, [...clientBookings, booking]);
    });

    return result;
  }, [bookings]);

  const isClientsLoading =
    isLoading ||
    clientQueries.some((query) => query.isLoading) ||
    bookingQueries.some((query) => query.isLoading);
  const isClientsError =
    clientQueries.some((query) => query.isError) ||
    bookingQueries.some((query) => query.isError);

  const visibleClients = clients.filter((client) =>
    clientMatchesSearch(client, search)
  );

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
        title="Клиенты"
        onMenuClick={() => setSidebarOpen(true)}
        onLogout={logout}
      />

      <main className="p-4 lg:p-6">
        <div className="mx-auto max-w-7xl space-y-5">
          {isError && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Не удалось загрузить филиалы. Проверь branches service и
              авторизацию через cookie access_token.
            </div>
          )}

          {!isError && isClientsError && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Филиалы загрузились, но не удалось получить клиентов по одному
              или нескольким филиалам.
            </div>
          )}

          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Клиенты</h2>
              <p className="mt-1 text-sm text-gray-500">
                {isManagerPath
                  ? "Клиенты вашего филиала"
                  : "Клиенты по всем доступным филиалам"}
              </p>
            </div>

            <label className="flex w-full items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 sm:max-w-xs">
              <span className="sr-only">Поиск клиента</span>
              <Search size={16} className="text-gray-400" />
              <input
                type="search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Поиск клиента"
                className="w-full bg-transparent text-sm outline-none"
              />
            </label>

            <label className="flex items-center rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-600">
              <span className="sr-only">Дата записей</span>
              <input
                type="date"
                value={selectedDate}
                onChange={(event) => setSelectedDate(event.target.value)}
                className="bg-transparent text-gray-900 outline-none"
              />
            </label>
          </div>

          {isClientsLoading && (
            <div className="rounded-xl bg-white p-5 text-sm text-gray-500 shadow-sm">
              Загружаем клиентов...
            </div>
          )}

          {!isClientsLoading && visibleClients.length === 0 && (
            <div className="rounded-xl bg-white p-8 text-center shadow-sm">
              <Users size={36} className="mx-auto text-gray-300" />
              <p className="mt-3 font-medium text-gray-900">
                Клиенты не найдены
              </p>
              <p className="mt-1 text-sm text-gray-500">
                В БД пока нет клиентов по доступным филиалам.
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {visibleClients.map((client) => (
              <ClientCard
                key={client.id}
                client={client}
                bookings={bookingsByClient.get(client.id) || []}
              />
            ))}
          </div>
        </div>
      </main>
    </div>
  );
};

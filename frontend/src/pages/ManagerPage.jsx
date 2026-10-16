import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import {
  getBranchBookings,
  getBranchEmployees,
  getBranches,
} from "../api/branches";
import { BranchDaySummary } from "../components/dashboard/BranchDaySummary";
import { EmployeeSchedule } from "../components/dashboard/EmployeeSchedule";
import { Sidebar } from "../components/layouts/Sidebar";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { useAuth } from "../context/AuthContext";

const normalizeBranch = (branch) => ({
  id: branch.id,
  name: branch.name,
  address: branch.address,
});

const normalizeEmployee = (employee, branch) => ({
  id: employee.id,
  branch_id: branch?.id,
  name: employee.name,
  surname: employee.surname,
  position: employee.position,
});

const getClientName = (client) =>
  [client?.surname, client?.name].filter(Boolean).join(" ") ||
  client?.name ||
  "";

const normalizeBooking = (booking) => ({
  id: booking.id,
  branch_id: booking.branch_id,
  employee_id: booking.employee_id,
  client_name: getClientName(booking.client),
  service_id: booking.service_id,
  service_name: booking.service_name,
  employeeName: [booking.employee_surname, booking.employee_name]
    .filter(Boolean)
    .join(" "),
  start_time: booking.start_time,
  end_time: booking.end_time,
  status:
    booking.status === "cancelled"
      ? "canceled"
      : booking.status || "confirmed",
  comment: booking.comment,
});

export const ManagerPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [selectedDate, setSelectedDate] = useState(
    new Date().toISOString().slice(0, 10)
  );
  const { user, logout } = useAuth();
  const { branchId } = useParams();

  const {
    data: apiBranches,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["branches", "manager"],
    queryFn: getBranches,
    retry: 1,
  });

  const branches = useMemo(() => {
    const source = Array.isArray(apiBranches) ? apiBranches : [];
    return source.map(normalizeBranch);
  }, [apiBranches]);

  const currentBranch =
    branches.find((branch) => Number(branch.id) === Number(branchId)) ||
    branches[0];

  const {
    data: apiEmployees = [],
    isLoading: isEmployeesLoading,
    isError: isEmployeesError,
  } = useQuery({
    queryKey: ["branches", currentBranch?.id, "employees", "schedule"],
    queryFn: () => getBranchEmployees(currentBranch.id),
    enabled: Boolean(currentBranch?.id),
    retry: 1,
  });

  const {
    data: apiBookings = [],
    isLoading: isBookingsLoading,
    isError: isBookingsError,
  } = useQuery({
    queryKey: ["branches", currentBranch?.id, "bookings", selectedDate],
    queryFn: () => getBranchBookings(currentBranch.id, selectedDate),
    enabled: Boolean(currentBranch?.id && selectedDate),
    retry: 1,
  });

  const employees = useMemo(() => {
    const source = Array.isArray(apiEmployees) ? apiEmployees : [];
    return source.map((employee) => normalizeEmployee(employee, currentBranch));
  }, [apiEmployees, currentBranch]);

  const selectedBookings = useMemo(() => {
    const source = Array.isArray(apiBookings) ? apiBookings : [];
    return source.map(normalizeBooking);
  }, [apiBookings]);

  const isScheduleLoading = isLoading || isEmployeesLoading || isBookingsLoading;

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(99,102,241,0.14),transparent_32rem),linear-gradient(135deg,#f8fbff_0%,#eef6ff_48%,#f7fbf4_100%)]">
      <Sidebar
        role={branchId ? "admin" : "manager"}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onLogout={logout}
      />

      <div className="flex min-w-0 flex-1 flex-col">
        <DashboardTopBar
          user={user}
          title="Расписание филиала"
          currentDate={selectedDate}
          branchName={currentBranch?.name || "Филиал менеджера"}
          onMenuClick={() => setSidebarOpen(true)}
          onLogout={logout}
          rightSlot={
            <label className="flex items-center rounded-lg border border-gray-200 bg-white px-3 py-2 text-sm text-gray-600">
              <span className="sr-only">Дата расписания</span>
              <input
                type="date"
                value={selectedDate}
                onChange={(event) => setSelectedDate(event.target.value)}
                className="bg-transparent text-gray-900 outline-none"
              />
            </label>
          }
        />

        <main className="flex-1 overflow-auto p-4 lg:p-6">
          <div className="mx-auto max-w-400 space-y-4">
            {isError && (
              <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
                Не удалось загрузить филиал из API. Проверь branches service на
                порту 8083 и cookie access_token.
              </div>
            )}

            {!isError && (isEmployeesError || isBookingsError) && (
              <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800">
                Филиал загрузился, но не удалось получить сотрудников или записи
                за выбранный день.
              </div>
            )}

            <BranchDaySummary
              branchName={currentBranch?.name || "Филиал менеджера"}
              bookings={selectedBookings}
              employees={employees}
            />

            {isScheduleLoading && (
              <div className="rounded-xl border border-gray-100 bg-white p-5 text-sm text-gray-500 shadow-sm">
                Загружаем расписание...
              </div>
            )}

            <EmployeeSchedule
              employees={employees}
              bookings={selectedBookings}
              selectedDate={selectedDate}
            />
          </div>
        </main>
      </div>
    </div>
  );
};

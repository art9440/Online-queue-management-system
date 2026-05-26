import { useMemo, useState } from "react";
import { useQueries, useQuery } from "@tanstack/react-query";
import { BriefcaseBusiness, Clock, Search, Scissors, X } from "lucide-react";
import {
  getBranchEmployees,
  getBranches,
  getServiceBranchEmployees,
  getServices,
} from "../api/branches";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { Sidebar } from "../components/layouts/Sidebar";
import { EMPLOYEE_AVATAR_COLORS } from "../constants/schedule";
import { useAuth } from "../context/AuthContext";
import { employeeMatchesSearch } from "../mocks/employeeSearch";
import PropTypes from "prop-types";

const normalizeBranch = (branch) => ({
  id: branch.id,
  name: branch.name,
  address: branch.address,
});

const getIsManagerPath = () => globalThis.window.location.pathname.startsWith("/manager");

const normalizeEmployee = (employee, branch) => ({
  id: employee.id,
  branch_id: branch.id,
  branchName: branch.name,
  name: employee.name,
  surname: employee.surname,
  position: employee.position,
});

const normalizeService = (service) => ({
  id: service.id,
  name: service.name,
  duration_minutes: service.duration_minutes,
  price: service.price,
});

const getEmployeeFullName = (employee) =>
  [employee.surname, employee.name].filter(Boolean).join(" ") || employee.name;

const getEmployeeInitials = (employee) =>
  (getEmployeeFullName(employee) || "?")
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();

const getAvatarColor = (employee) =>
  EMPLOYEE_AVATAR_COLORS[
    Number(employee.id || 0) % EMPLOYEE_AVATAR_COLORS.length
  ];

const formatCurrency = (value) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    maximumFractionDigits: 0,
  }).format(value || 0);

const EmployeeServicesModal = ({
  employee,
  services,
  isLoading,
  isError,
  onClose,
}) => {
  if (!employee) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4">
      <section className="w-full max-w-2xl overflow-hidden rounded-2xl bg-white shadow-2xl shadow-slate-950/20">
        <div className="flex items-start justify-between gap-4 border-b border-gray-100 p-5">
          <div>
            <p className="text-sm font-medium text-indigo-700">
              {employee.branchName}
            </p>
            <h2 className="mt-1 text-xl font-semibold text-gray-950">
              {getEmployeeFullName(employee)}
            </h2>
            <p className="mt-1 text-sm text-gray-500">
              {employee.position || "Сотрудник"}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl p-2 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700"
            aria-label="Закрыть"
          >
            <X size={20} />
          </button>
        </div>

        <div className="max-h-[60vh] overflow-auto p-5">
          {isLoading && (
            <p className="rounded-xl bg-gray-50 p-4 text-sm text-gray-500">
              Загружаем услуги сотрудника...
            </p>
          )}

          {isError && (
            <p className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Не удалось загрузить услуги сотрудника.
            </p>
          )}

          {!isLoading && !isError && services.length === 0 && (
            <p className="rounded-xl bg-gray-50 p-4 text-sm text-gray-500">
              Для сотрудника пока не найдены услуги.
            </p>
          )}

          {!isLoading && !isError && services.length > 0 && (
            <div className="grid gap-3 sm:grid-cols-2">
              {services.map((service) => (
                <article
                  key={service.id}
                  className="rounded-xl border border-indigo-100 bg-indigo-50/50 p-4"
                >
                  <div className="flex items-start gap-3">
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-white text-indigo-600 ring-1 ring-indigo-100">
                      <Scissors size={18} />
                    </span>
                    <div className="min-w-0">
                      <h3 className="truncate font-semibold text-gray-950">
                        {service.name}
                      </h3>
                      <p className="mt-2 flex items-center gap-1.5 text-sm text-gray-600">
                        <Clock size={14} />
                        {service.duration_minutes || 0} мин
                      </p>
                      <p className="mt-1 text-sm font-semibold text-gray-900">
                        {formatCurrency(service.price)}
                      </p>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  );
};

EmployeeServicesModal.propTypes = {
  employee: PropTypes.shape({
    id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    branchName: PropTypes.string,
    name: PropTypes.string,
    surname: PropTypes.string,
    position: PropTypes.string,
  }),
  services: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      name: PropTypes.string,
      duration_minutes: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      price: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    })
  ).isRequired,
  isLoading: PropTypes.bool.isRequired,
  isError: PropTypes.bool.isRequired,
  onClose: PropTypes.func.isRequired,
};

export const EmployeesPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [selectedEmployee, setSelectedEmployee] = useState(null);
  const { user, logout } = useAuth();
  const isManagerPath = getIsManagerPath();

  const { data: apiBranches = [], isLoading, isError } = useQuery({
    queryKey: ["branches", isManagerPath ? "manager" : "admin", "employees"],
    queryFn: getBranches,
    retry: 1,
  });

  const branches = useMemo(() => {
    const source = Array.isArray(apiBranches) ? apiBranches : [];
    return source.map(normalizeBranch);
  }, [apiBranches]);

  const employeeQueries = useQueries({
    queries: branches.map((branch) => ({
      queryKey: ["branches", branch.id, "employees"],
      queryFn: () => getBranchEmployees(branch.id),
      enabled: Boolean(branch.id),
      retry: 1,
    })),
  });

  const {
    data: apiServices = [],
    isLoading: isServicesLoading,
    isError: isServicesError,
  } = useQuery({
    queryKey: ["services", "employee-details"],
    queryFn: getServices,
    enabled: Boolean(selectedEmployee),
    retry: 1,
  });

  const services = useMemo(() => {
    const source = Array.isArray(apiServices) ? apiServices : [];
    return source.map(normalizeService);
  }, [apiServices]);

  const serviceEmployeeQueries = useQueries({
    queries: services.map((service) => ({
      queryKey: [
        "services",
        service.id,
        "branches",
        selectedEmployee?.branch_id,
        "employees",
      ],
      queryFn: () =>
        getServiceBranchEmployees(service.id, selectedEmployee.branch_id),
      enabled: Boolean(selectedEmployee?.branch_id && service.id),
      retry: 1,
    })),
  });

  const employees = useMemo(() => {
    return branches.flatMap((branch, index) => {
      const source = employeeQueries[index]?.data || [];
      return source.map((employee) => normalizeEmployee(employee, branch));
    });
  }, [branches, employeeQueries]);

  const isEmployeesLoading =
    isLoading || employeeQueries.some((query) => query.isLoading);
  const isEmployeesError = employeeQueries.some((query) => query.isError);

  const visibleEmployees = employees.filter((employee) =>
    employeeMatchesSearch(employee, search)
  );
  const selectedEmployeeServices = services.filter((service, index) => {
    const serviceEmployees = serviceEmployeeQueries[index]?.data || [];
    return serviceEmployees.some(
      (employee) => Number(employee.id) === Number(selectedEmployee?.id)
    );
  });
  const isEmployeeServicesLoading =
    isServicesLoading ||
    serviceEmployeeQueries.some((query) => query.isLoading);
  const isEmployeeServicesError =
    isServicesError ||
    serviceEmployeeQueries.some((query) => query.isError);

  const handleEmployeeRowKeyDown = (event, employee) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      setSelectedEmployee(employee);
    }
  };

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
        title="Сотрудники"
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

          {!isError && isEmployeesError && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Филиалы загрузились, но не удалось получить сотрудников по одному
              или нескольким филиалам.
            </div>
          )}

          <div className="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">
                Сотрудники
              </h2>
              <p className="mt-1 text-sm text-gray-500">
                {isManagerPath
                  ? "Команда вашего филиала"
                  : "Сотрудники по всем доступным филиалам"}
              </p>
            </div>

            <label className="flex w-full items-center gap-2 rounded-lg border border-gray-200 bg-white px-3 py-2 sm:max-w-xs">
              <span className="sr-only">Поиск сотрудника</span>
              <Search size={16} className="text-gray-400" />
              <input
                type="search"
                value={search}
                onChange={(event) => setSearch(event.target.value)}
                placeholder="Поиск сотрудника"
                className="w-full bg-transparent text-sm outline-none"
              />
            </label>
          </div>

          {isEmployeesLoading && (
            <div className="rounded-xl bg-white p-5 text-sm text-gray-500 shadow-sm">
              Загружаем сотрудников...
            </div>
          )}

          {!isEmployeesLoading && visibleEmployees.length === 0 && (
            <div className="rounded-xl bg-white p-8 text-center shadow-sm">
              <BriefcaseBusiness size={36} className="mx-auto text-gray-300" />
              <p className="mt-3 font-medium text-gray-900">
                Сотрудники не найдены
              </p>
              <p className="mt-1 text-sm text-gray-500">
                API вернул пустой список по доступным филиалам.
              </p>
            </div>
          )}

          {visibleEmployees.length > 0 && (
            <div className="overflow-x-auto rounded-2xl border border-indigo-100 bg-white shadow-sm">
              <table className="min-w-180 w-full border-collapse text-left text-sm">
                <thead className="bg-indigo-50 text-xs uppercase tracking-wide text-indigo-700">
                  <tr>
                    <th className="px-4 py-3 font-semibold">Сотрудник</th>
                    <th className="px-4 py-3 font-semibold">Филиал</th>
                    <th className="px-4 py-3 font-semibold">Должность</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  {visibleEmployees.map((employee) => {
                    const fullName = getEmployeeFullName(employee);

                    return (
                      <tr
                        key={employee.id}
                        role="button"
                        tabIndex={0}
                        onClick={() => setSelectedEmployee(employee)}
                        onKeyDown={(event) => handleEmployeeRowKeyDown(event, employee)}
                        className="cursor-pointer hover:bg-indigo-50/40"
                      >
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <span
                              className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${getAvatarColor(employee)}`}
                            >
                              {getEmployeeInitials(employee)}
                            </span>
                            <div className="min-w-0">
                              <p className="truncate font-medium text-gray-900">
                                {fullName}
                              </p>
                              {employee.phone && (
                                <p className="mt-0.5 text-xs text-gray-500">
                                  {employee.phone}
                                </p>
                              )}
                            </div>
                          </div>
                        </td>
                        <td className="px-4 py-3 text-gray-600">
                          {employee.branchName}
                        </td>
                        <td className="px-4 py-3 text-gray-600">
                          {employee.position || "Сотрудник"}
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </main>

      <EmployeeServicesModal
        employee={selectedEmployee}
        services={selectedEmployeeServices}
        isLoading={isEmployeeServicesLoading}
        isError={isEmployeeServicesError}
        onClose={() => setSelectedEmployee(null)}
      />
    </div>
  );
};

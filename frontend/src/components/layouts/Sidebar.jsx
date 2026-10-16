import { NavLink } from "react-router-dom";
import {
  BriefcaseBusiness,
  CalendarDays,
  LogOut,
  Scissors,
  Settings,
  Store,
  Users,
} from "lucide-react";
import PropTypes from "prop-types";

const adminItems = [
  { to: "/admin", label: "Филиалы", icon: Store },
  { to: "/admin/clients", label: "Клиенты", icon: Users },
  { to: "/admin/services", label: "Услуги", icon: Scissors },
  { to: "/admin/employees", label: "Сотрудники", icon: BriefcaseBusiness },
  { to: "/admin/settings", label: "Настройки", icon: Settings },
];

const managerItems = [
  { to: "/manager", label: "Расписание", icon: CalendarDays },
  { to: "/manager/clients", label: "Клиенты", icon: Users },
  { to: "/manager/services", label: "Услуги", icon: Scissors },
  { to: "/manager/employees", label: "Сотрудники", icon: BriefcaseBusiness },
  { to: "/manager/settings", label: "Настройки", icon: Settings },
];

export const Sidebar = ({
  role = "admin",
  isOpen = true,
  onClose,
  onLogout,
  brand = "Queue Manager",
}) => {
  const items = role === "manager" ? managerItems : adminItems;

  return (
    <>
      {isOpen && (
        <button
          type="button"
          aria-label="Закрыть меню"
          className="fixed inset-0 z-40 bg-black/35"
          onClick={onClose}
        />
      )}

      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-64 flex-col bg-linear-to-b from-slate-950 via-indigo-950 to-slate-950 text-white shadow-2xl transition-transform duration-200 ${
          isOpen ? "translate-x-0" : "-translate-x-full"
        }`}
      >
        <div className="flex h-18 items-center gap-3 border-b border-white/10 px-5 py-4">
          <div className="flex h-10 w-10 items-center justify-center rounded-2xl bg-linear-to-br from-violet-500 to-fuchsia-400 text-lg font-black shadow-lg shadow-indigo-950/40">
            Q
          </div>
          <div>
            <span className="block text-base font-bold tracking-tight">{brand}</span>
            <span className="text-xs text-slate-400">Панель управления</span>
          </div>
        </div>

        <nav className="flex-1 overflow-y-auto px-3 py-4">
          <ul className="space-y-1">
            {items.map((item) => (
              <li key={item.to}>
                <NavLink
                  to={item.to}
                  end={item.to === "/admin" || item.to === "/manager"}
                  onClick={() => {
                    if (globalThis.window.innerWidth < 1024) onClose?.();
                  }}
                  className={({ isActive }) =>
                    `flex items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium transition-all ${
                      isActive
                        ? "bg-white text-slate-950 shadow-lg shadow-black/20"
                        : "text-slate-300 hover:bg-white/10 hover:text-white"
                    }`
                  }
                >
                  <item.icon size={19} />
                  <span>{item.label}</span>
                </NavLink>
              </li>
            ))}
          </ul>
        </nav>

        <div className="border-t border-white/10 p-4">
          <button
            type="button"
            onClick={onLogout}
            className="flex w-full items-center gap-3 rounded-xl px-4 py-3 text-sm font-medium text-slate-300 transition-colors hover:bg-rose-500/15 hover:text-white"
          >
            <LogOut size={19} />
            <span>Выйти</span>
          </button>
        </div>
      </aside>
    </>
  );
};

Sidebar.propTypes = {
  role: PropTypes.string,
  isOpen: PropTypes.bool,
  onClose: PropTypes.func,
  onLogout: PropTypes.func,
  brand: PropTypes.string,
};
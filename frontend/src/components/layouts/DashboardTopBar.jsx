import {
  Bell,
  CalendarDays,
  LogOut,
  Menu,
  Store,
} from "lucide-react";
import { useState } from "react";

const formatDateParts = (date = new Date()) => {
  const source = new Date(date);
  return {
    weekday: new Intl.DateTimeFormat("ru-RU", { weekday: "long" }).format(source),
    date: new Intl.DateTimeFormat("ru-RU", {
      day: "numeric",
      month: "long",
      year: "numeric",
    }).format(source),
  };
};

export const DashboardTopBar = ({
  user,
  title,
  currentDate = new Date(),
  businessName,
  branchName,
  onMenuClick,
  onLogout,
  rightSlot,
}) => {
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const dateParts = formatDateParts(currentDate);
  const businessLabel =
    businessName ||
    user?.business_name ||
    user?.businessName ||
    user?.business ||
    (user?.business_id ? `Бизнес #${user.business_id}` : "Мой бизнес");

  return (
    <header className="sticky top-0 z-30 bg-indigo-700 px-4 py-3 text-white shadow-lg shadow-indigo-900/20 lg:px-6">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
        <div className="flex min-w-0 items-center gap-3">
          <button
            type="button"
            onClick={onMenuClick}
            className="rounded-xl bg-white/10 p-2 text-white transition hover:bg-white/20"
            aria-label="Открыть меню"
          >
            <Menu size={20} />
          </button>

          <div className="min-w-0">
            <p className="text-xs font-medium uppercase tracking-wide text-indigo-100">
              {title}
            </p>
            <h1 className="truncate text-2xl font-bold text-white">
              {businessLabel}
            </h1>
            <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-sm text-indigo-100">
              {branchName && (
                <span className="flex items-center gap-1.5">
                  <Store size={14} />
                  {branchName}
                </span>
              )}
              <span className="flex items-center gap-1.5 capitalize">
                <CalendarDays size={14} />
                {dateParts.weekday}, {dateParts.date}
              </span>
            </div>
          </div>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
          <div className="rounded-2xl border border-white/15 bg-white/10 px-4 py-3 backdrop-blur">
            <p className="text-xs font-medium text-indigo-100">Пользователь</p>
            <p className="mt-0.5 truncate text-sm font-semibold text-white">
              {user?.login || "Пользователь"}
            </p>
          </div>

          <div className="flex items-center gap-2">
            {rightSlot && <div>{rightSlot}</div>}

            <div className="relative">
              <button
                type="button"
                onClick={() => setNotificationsOpen((prev) => !prev)}
                className="relative rounded-xl border border-white/15 bg-white/10 p-2.5 text-white shadow-sm transition hover:bg-white/20"
                aria-label="Уведомления"
              >
                <Bell size={18} />
                <span className="absolute right-2 top-2 h-2 w-2 rounded-full bg-amber-300 ring-2 ring-indigo-700" />
              </button>

              {notificationsOpen && (
                <div className="absolute right-0 top-12 w-72 rounded-2xl border border-indigo-100 bg-white p-4 text-slate-900 shadow-2xl shadow-indigo-950/20">
                  <p className="font-semibold">Уведомления</p>
                  <p className="mt-2 text-sm text-slate-500">
                    Пока новых уведомлений нет. Здесь позже появятся события по
                    записям, клиентам и филиалам.
                  </p>
                </div>
              )}
            </div>

            <button
              type="button"
              onClick={onLogout}
              className="inline-flex items-center gap-2 rounded-xl bg-white px-3 py-2.5 text-sm font-semibold text-indigo-700 shadow-sm transition hover:bg-indigo-50"
            >
              <LogOut size={17} />
              <span className="hidden sm:inline">Выйти</span>
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};

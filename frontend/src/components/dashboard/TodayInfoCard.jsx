import { CalendarDays, CircleDollarSign, Sparkles, TrendingUp } from "lucide-react";

const formatCurrency = (value) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    maximumFractionDigits: 0,
  }).format(value || 0);

export const TodayInfoCard = ({
  bookingsCount = 0,
  revenue = 0,
  newClients = 0,
  utilization = null,
}) => {
  return (
    <section className="rounded-2xl border border-white/70 bg-white/90 p-4 shadow-sm shadow-slate-200/80">
      <div className="flex flex-col gap-1 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-sm font-medium text-indigo-700">Сегодня</p>
          <h2 className="text-lg font-bold text-gray-900">
            Краткая сводка
          </h2>
        </div>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-2 lg:grid-cols-4">
        <div className="rounded-lg bg-gray-50 p-2.5">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <CalendarDays size={14} />
            <span>Записей</span>
          </div>
          <p className="mt-1 text-base font-semibold text-gray-900">
            {bookingsCount}
          </p>
        </div>
        <div className="rounded-lg bg-gray-50 p-2.5">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <CircleDollarSign size={14} />
            <span>Выручка</span>
          </div>
          <p className="mt-1 text-base font-semibold text-gray-900">
            {formatCurrency(revenue)}
          </p>
        </div>
        <div className="rounded-lg bg-gray-50 p-2.5">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <Sparkles size={14} />
            <span>Новые клиенты</span>
          </div>
          <p className="mt-1 text-base font-semibold text-gray-900">
            {newClients}
          </p>
        </div>
        <div className="rounded-lg bg-gray-50 p-2.5">
          <div className="flex items-center gap-2 text-xs text-gray-500">
            <TrendingUp size={14} />
            <span>Загрузка</span>
          </div>
          <p className="mt-1 text-base font-semibold text-gray-900">
            {utilization === null || utilization === undefined
              ? "н/д"
              : `${Math.round(utilization)}%`}
          </p>
        </div>
      </div>
    </section>
  );
};

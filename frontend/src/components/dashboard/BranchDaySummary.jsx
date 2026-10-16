import { CircleDollarSign, Clock, UserRoundCheck, Users } from "lucide-react";
import PropTypes from "prop-types";

const formatCurrency = (value) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    maximumFractionDigits: 0,
  }).format(value || 0);

const getBookingDuration = (booking) => {
  const start = new Date(booking.start_time);
  const end = new Date(booking.end_time);

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) return 0;
  return Math.max(0, Math.round((end.getTime() - start.getTime()) / 60000));
};

const SummaryItem = ({ icon, label, value }) => {
  const IconComponent = icon;

  return (
  <div className="rounded-xl bg-white/75 px-3 py-2 shadow-sm ring-1 ring-white/70">
    <div className="flex items-center gap-2 text-xs font-medium text-slate-500">
      <IconComponent size={14} />
      <span>{label}</span>
    </div>
    <p className="mt-1 text-base font-semibold text-slate-950">{value}</p>
  </div>
  );
};

export const BranchDaySummary = ({
  branchName,
  bookings = [],
  employees = [],
}) => {
  const revenue = bookings.reduce(
    (sum, booking) => sum + Number(booking.price || 0),
    0
  );
  const bookedEmployeeIds = new Set(
    bookings.map((booking) => booking.employee_id)
  );
  const totalMinutes = bookings.reduce(
    (sum, booking) => sum + getBookingDuration(booking),
    0
  );

  return (
    <section className="rounded-2xl bg-linear-to-br from-indigo-50 via-sky-50 to-emerald-50 p-4 shadow-sm ring-1 ring-indigo-100">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="min-w-0">
          <p className="text-sm font-medium text-indigo-700">
            {branchName || "Филиал"}
          </p>
          <h2 className="mt-1 text-2xl font-bold text-slate-950">
            Сводка по выбранному дню
          </h2>
        </div>

        <div className="grid grid-cols-2 gap-2 md:grid-cols-4 lg:min-w-170">
          <SummaryItem
            icon={Clock}
            label="Записей"
            value={bookings.length}
          />
          <SummaryItem
            icon={CircleDollarSign}
            label="Выручка"
            value={formatCurrency(revenue)}
          />
          <SummaryItem
            icon={Users}
            label="Сотрудников"
            value={`${bookedEmployeeIds.size}/${employees.length}`}
          />
          <SummaryItem
            icon={UserRoundCheck}
            label="Занятость"
            value={`${Math.round(totalMinutes / 60)} ч`}
          />
        </div>
      </div>
    </section>
  );
};

SummaryItem.propTypes = {
  icon: PropTypes.elementType.isRequired,
  label: PropTypes.string.isRequired,
  value: PropTypes.oneOfType([PropTypes.number, PropTypes.string]).isRequired,
};

BranchDaySummary.propTypes = {
  branchName: PropTypes.string,
  bookings: PropTypes.arrayOf(
    PropTypes.shape({
      start_time: PropTypes.string,
      end_time: PropTypes.string,
      price: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      employee_id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    })
  ),
  employees: PropTypes.arrayOf(PropTypes.object),
};
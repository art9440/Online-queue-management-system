import { BookingCard } from "./BookingCard";
import {
  DEFAULT_TIME_SLOTS,
  EMPLOYEE_AVATAR_COLORS,
} from "../../constants/schedule";
import PropTypes from "prop-types";

const getBookingTime = (booking) => {
  const start = booking.start_time;
  if (!start) return "";
  const date = new Date(start);
  if (Number.isNaN(date.getTime())) {
    return String(start).split("T")[1]?.slice(0, 5) || String(start).slice(0, 5);
  }
  return new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(date);
};

const getEmployeeName = (employee) =>
  [employee.surname, employee.name].filter(Boolean).join(" ") || employee.name;

const getEmployeeInitials = (employee) => {
  const name = getEmployeeName(employee);
  return name
    .split(" ")
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0])
    .join("")
    .toUpperCase();
};

export const EmployeeSchedule = ({
  employees = [],
  bookings = [],
  selectedDate,
  timeSlots = DEFAULT_TIME_SLOTS,
  emptyText = "Свободно",
}) => {
  const dateKey =
    selectedDate || new Date().toISOString().slice(0, 10);

  const getBooking = (employeeId, time) =>
    bookings.find((booking) => {
      const bookingDate = String(booking.start_time || "").slice(0, 10);
      return (
        Number(booking.employee_id) === Number(employeeId) &&
        bookingDate === dateKey &&
        getBookingTime(booking) === time
      );
    });

  if (!employees.length) {
    return (
      <section className="rounded-xl border border-gray-100 bg-white p-8 text-center shadow-sm">
        <p className="text-sm text-gray-500">Нет сотрудников для расписания</p>
      </section>
    );
  }

  return (
    <section className="overflow-hidden rounded-xl border border-gray-100 bg-white shadow-sm">
      <div className="border-b border-gray-100 p-5">
        <h2 className="font-semibold text-gray-900">Расписание сотрудников</h2>
        <p className="mt-1 text-sm text-gray-500">
          Записи и свободные слоты на выбранный день
        </p>
      </div>

      <div className="overflow-x-auto">
        <div
          className="min-w-190 grid border-b border-gray-100 bg-gray-50"
          style={{
            gridTemplateColumns: `96px repeat(${employees.length}, minmax(180px, 1fr))`,
          }}
        >
          <div className="border-r border-gray-100 p-4 text-sm font-medium text-gray-500">
            Время
          </div>
          {employees.map((employee) => (
            <div
              key={employee.id}
              className="flex items-center gap-3 border-r border-gray-100 p-4 last:border-r-0"
            >
              <span
                className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full text-xs font-bold text-white ${
                  EMPLOYEE_AVATAR_COLORS[
                    Number(employee.id || 0) % EMPLOYEE_AVATAR_COLORS.length
                  ]
                }`}
              >
                {getEmployeeInitials(employee)}
              </span>
              <div className="min-w-0">
                <p className="truncate text-sm font-semibold text-gray-900">
                  {getEmployeeName(employee)}
                </p>
                <p className="mt-1 truncate text-xs text-gray-500">
                  {employee.position || "Сотрудник"}
                </p>
              </div>
            </div>
          ))}
        </div>

        {timeSlots.map((time) => (
          <div
            key={time}
            className="min-w-190 grid border-b border-gray-100 last:border-b-0"
            style={{
              gridTemplateColumns: `96px repeat(${employees.length}, minmax(180px, 1fr))`,
            }}
          >
            <div className="border-r border-gray-100 bg-indigo-50 px-3 py-4 text-center">
              <span className="inline-flex rounded-full bg-white px-2.5 py-1 text-sm font-semibold text-indigo-700 shadow-sm ring-1 ring-indigo-100">
                {time}
              </span>
            </div>
            {employees.map((employee) => {
              const booking = getBooking(employee.id, time);
              return (
                <div
                  key={employee.id}
                  className={`border-r border-gray-100 p-2 last:border-r-0 ${
                    booking ? "min-h-32" : "min-h-18"
                  }`}
                >
                  {booking ? (
                    <BookingCard booking={booking} />
                  ) : (
                    <div className="flex min-h-12 items-center justify-center rounded-lg border border-dashed border-gray-200 bg-gray-50/70 text-xs text-gray-400">
                      {emptyText}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        ))}
      </div>
    </section>
  );
};

EmployeeSchedule.propTypes = {
  employees: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      name: PropTypes.string,
      surname: PropTypes.string,
      position: PropTypes.string,
    })
  ),
  bookings: PropTypes.arrayOf(
    PropTypes.shape({
      id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      employee_id: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      start_time: PropTypes.string,
    })
  ),
  selectedDate: PropTypes.string,
  timeSlots: PropTypes.arrayOf(PropTypes.string),
  emptyText: PropTypes.string,
};
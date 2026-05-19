import { Clock, UserRound } from "lucide-react";

const bookingTones = [
  "border-teal-200 bg-teal-50 text-teal-950",
  "border-orange-200 bg-orange-50 text-orange-950",
  "border-sky-200 bg-sky-50 text-sky-950",
  "border-fuchsia-200 bg-fuchsia-50 text-fuchsia-950",
  "border-lime-200 bg-lime-50 text-lime-950",
  "border-violet-200 bg-violet-50 text-violet-950",
];

const statusClasses = {
  confirmed: "bg-green-50 text-green-700",
  pending: "bg-amber-50 text-amber-700",
  canceled: "bg-red-50 text-red-700",
  completed: "bg-gray-100 text-gray-700",
};

const statusLabels = {
  confirmed: "Подтверждена",
  pending: "Ожидает",
  canceled: "Отменена",
  completed: "Завершена",
};

const formatTime = (value) => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat("ru-RU", {
    hour: "2-digit",
    minute: "2-digit",
  }).format(date);
};

export const BookingCard = ({ booking }) => {
  if (!booking) return null;

  const status = booking.status || "confirmed";
  const tone = bookingTones[Number(booking.id || 0) % bookingTones.length];

  return (
    <article className={`rounded-lg border p-3 shadow-sm ${tone}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="flex items-center gap-1.5 text-sm font-semibold truncate">
            <UserRound size={14} />
            <span>{booking.client_name}</span>
          </p>
          <p className="mt-1 text-sm opacity-75 truncate">
            {booking.service_name}
          </p>
        </div>
        <span
          className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-medium ${
            statusClasses[status] || statusClasses.confirmed
          }`}
        >
          {statusLabels[status] || status}
        </span>
      </div>

      <div className="mt-3 flex items-center justify-between text-sm">
        <span className="flex items-center gap-1.5 font-medium">
          <Clock size={14} />
          {formatTime(booking.start_time)}
          {booking.end_time
            ? ` - ${formatTime(booking.end_time)}`
            : ""}
        </span>
        <span className="opacity-75">
          {booking.price ? `${booking.price} ₽` : ""}
        </span>
      </div>

      {booking.employeeName && (
        <p className="mt-2 text-xs opacity-75">Мастер: {booking.employeeName}</p>
      )}

      {booking.comment && (
        <p className="mt-3 border-t border-current/10 pt-3 text-sm opacity-75">
          {booking.comment}
        </p>
      )}
    </article>
  );
};

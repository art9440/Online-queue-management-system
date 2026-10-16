const getTime = (value) => {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date.getTime();
};

const getMinutesBetween = (start, end) => {
  const startTime = getTime(start);
  const endTime = getTime(end);

  if (!startTime || !endTime || endTime <= startTime) return 0;

  return Math.round((endTime - startTime) / 60000);
};

export const getBookingEmployeeId = (booking) =>
  Number(booking.employee_id || booking.employeeId);

export const getBookingDateKey = (booking) =>
  String(booking.start_time || booking.startTime || "").slice(0, 10);

export const getEmployeeDayStats = ({
  employee,
  bookings = [],
  schedules = [],
  dateKey,
}) => {
  const employeeId = Number(employee.id);
  const employeeBookings = bookings.filter(
    (booking) =>
      getBookingEmployeeId(booking) === employeeId &&
      (!dateKey || getBookingDateKey(booking) === dateKey)
  );
  const employeeSchedules = schedules.filter(
    (schedule) =>
      Number(schedule.employee_id || schedule.employeeId) === employeeId
  );

  const bookedMinutes = employeeBookings.reduce(
    (sum, booking) =>
      sum +
      getMinutesBetween(
        booking.start_time || booking.startTime,
        booking.end_time || booking.endTime
      ),
    0
  );
  const workingMinutes = employeeSchedules.reduce(
    (sum, schedule) =>
      sum +
      getMinutesBetween(
        schedule.starts_at || schedule.startsAt,
        schedule.ends_at || schedule.endsAt
      ),
    0
  );

  return {
    bookingsToday: employeeBookings.length,
    bookedMinutes,
    workingMinutes,
    utilization:
      workingMinutes > 0
        ? Math.round((bookedMinutes / workingMinutes) * 100)
        : null,
  };
};

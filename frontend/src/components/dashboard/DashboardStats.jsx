import { BriefcaseBusiness, CalendarDays, CircleDollarSign, Users } from "lucide-react";
import { StatsCard } from "./StatsCard";

const toNumber = (value) => Number(value || 0);

const sum = (items, selector) =>
  items.reduce((total, item) => total + toNumber(selector(item)), 0);

export const DashboardStats = ({
  bookings = [],
  employees = [],
  services = [],
  servicePrices = new Map(),
}) => {
  const todayRevenue = sum(bookings, (booking) => {
    return booking.price || servicePrices.get(Number(booking.service_id));
  });

  const cards = [
    {
      title: "Записей сегодня",
      value: bookings.length,
      subtitle: "по всем доступным филиалам",
      icon: CalendarDays,
      tone: "blue",
    },
    {
      title: "Выручка сегодня",
      value: new Intl.NumberFormat("ru-RU", {
        style: "currency",
        currency: "RUB",
        maximumFractionDigits: 0,
      }).format(todayRevenue),
      subtitle: "по завершенным и активным записям",
      icon: CircleDollarSign,
      tone: "green",
    },
    {
      title: "Сотрудников",
      value: employees.length,
      subtitle: "в доступных филиалах",
      icon: Users,
      tone: "amber",
    },
    {
      title: "Услуг",
      value: services.length,
      subtitle: "доступно для записи",
      icon: BriefcaseBusiness,
      tone: "indigo",
    },
  ];

  return (
    <section className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      {cards.map((card) => (
        <StatsCard key={card.title} {...card} />
      ))}
    </section>
  );
};

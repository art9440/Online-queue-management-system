const today = new Date().toISOString().slice(0, 10);
const tomorrow = new Date(Date.now() + 86400000).toISOString().slice(0, 10);
const createdAt = `${today}T08:00:00Z`;

export const mockBranches = [
  {
    id: 1,
    business_id: 1,
    name: "Центральный филиал",
    address: "ул. Ленина, 10",
    timezone: "Asia/Novosibirsk",
  },
  {
    id: 2,
    business_id: 1,
    name: "Левый берег",
    address: "пр. Карла Маркса, 25",
    timezone: "Asia/Novosibirsk",
  },
  {
    id: 3,
    business_id: 1,
    name: "Академгородок",
    address: "ул. Николаева, 8",
    timezone: "Asia/Novosibirsk",
  },
];

export const mockEmployees = [
  {
    id: 1,
    branch_id: 1,
    name: "Анна",
    surname: "Петрова",
    position: "Стилист",
    created_at: createdAt,
  },
  {
    id: 2,
    branch_id: 1,
    name: "Дмитрий",
    surname: "Соколов",
    position: "Барбер",
    created_at: createdAt,
  },
  {
    id: 3,
    branch_id: 2,
    name: "Елена",
    surname: "Смирнова",
    position: "Мастер маникюра",
    created_at: createdAt,
  },
  {
    id: 4,
    branch_id: 3,
    name: "Мария",
    surname: "Волкова",
    position: "Косметолог",
    created_at: createdAt,
  },
];

export const mockServices = [
  {
    id: 1,
    branch_id: 1,
    name: "Консультация",
    duration_minutes: 30,
    price: 1500,
    created_at: createdAt,
  },
  {
    id: 2,
    branch_id: 1,
    name: "Диагностика",
    duration_minutes: 45,
    price: 2200,
    created_at: createdAt,
  },
  {
    id: 3,
    branch_id: 2,
    name: "Консультация",
    duration_minutes: 30,
    price: 1400,
    created_at: createdAt,
  },
  {
    id: 4,
    branch_id: 2,
    name: "Маникюр",
    duration_minutes: 90,
    price: 2600,
    created_at: createdAt,
  },
  {
    id: 5,
    branch_id: 3,
    name: "Консультация",
    duration_minutes: 30,
    price: 1600,
    created_at: createdAt,
  },
];

export const mockClients = [
  {
    id: 1,
    email: "elena.petrovna@example.com",
    phone: "+7 900 123-45-67",
    name: "Елена",
    surname: "Петрова",
    tg_username: "elena_queue",
    created_at: createdAt,
  },
  {
    id: 2,
    email: "d.volkov@example.com",
    phone: "+7 900 765-43-21",
    name: "Дмитрий",
    surname: "Волков",
    tg_username: null,
    created_at: createdAt,
  },
  {
    id: 3,
    email: null,
    phone: "+7 900 456-78-90",
    name: "Марина",
    surname: "Кузнецова",
    tg_username: "marina_nsk",
    created_at: createdAt,
  },
  {
    id: 4,
    email: "olga.novikova@example.com",
    phone: "+7 900 333-12-10",
    name: "Ольга",
    surname: "Новикова",
    tg_username: null,
    created_at: createdAt,
  },
];

export const mockBookings = [
  {
    id: 1,
    client_id: 1,
    branch_id: 1,
    employee_id: 1,
    service_id: 1,
    start_time: `${today}T09:00:00Z`,
    end_time: `${today}T09:30:00Z`,
    status: "confirmed",
    comment: "Первичный визит",
    created_at: createdAt,
  },
  {
    id: 2,
    client_id: 2,
    branch_id: 1,
    employee_id: 2,
    service_id: 2,
    start_time: `${today}T10:00:00Z`,
    end_time: `${today}T10:45:00Z`,
    status: "pending",
    comment: null,
    created_at: createdAt,
  },
  {
    id: 3,
    client_id: 3,
    branch_id: 2,
    employee_id: 3,
    service_id: 4,
    start_time: `${today}T11:00:00Z`,
    end_time: `${today}T12:30:00Z`,
    status: "confirmed",
    comment: "Просит тихое место",
    created_at: createdAt,
  },
  {
    id: 4,
    client_id: 4,
    branch_id: 3,
    employee_id: 4,
    service_id: 5,
    start_time: `${tomorrow}T12:00:00Z`,
    end_time: `${tomorrow}T12:30:00Z`,
    status: "confirmed",
    comment: null,
    created_at: createdAt,
  },
];

export const getBranchDashboardData = (branchIds) => {
  const ids = new Set(branchIds.map(Number));
  const branchBookings = mockBookings.filter((booking) =>
    ids.has(Number(booking.branch_id))
  );
  const branchClientIds = new Set(
    branchBookings.map((booking) => Number(booking.client_id))
  );

  return {
    employees: mockEmployees.filter((employee) =>
      ids.has(Number(employee.branch_id))
    ),
    clients: mockClients.filter((client) => branchClientIds.has(Number(client.id))),
    services: mockServices.filter((service) =>
      ids.has(Number(service.branch_id))
    ),
    bookings: branchBookings,
  };
};

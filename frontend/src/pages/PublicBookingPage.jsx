import { useMemo, useState } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useParams } from "react-router-dom";
import {
    createPublicAppointment,
    getGoogleOAuthUrl,
    getPublicBranches,
    getPublicEmployees,
    getPublicServices,
    getPublicSlots,
} from "../api/publicBooking";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

const steps = ["service", "branch", "employee", "time", "client"];
const stepTitles = {
    service: "Выберите услугу",
    branch: "Выберите филиал",
    employee: "Выберите мастера",
    time: "Выберите дату и время",
    client: "Ваши данные",
};

const today = () => new Date().toISOString().slice(0, 10);

const formatMoney = (value) => {
    const number = Number(value);
    if (Number.isNaN(number)) return value;

    return new Intl.NumberFormat("ru-RU", {
        style: "currency",
        currency: "RUB",
        maximumFractionDigits: 0,
    }).format(number);
};

const formatSlot = (slot, fallbackTimezone = "Asia/Novosibirsk") => {
    const formatter = new Intl.DateTimeFormat("ru-RU", {
        hour: "2-digit",
        minute: "2-digit",
        timeZone: slot.timezone || fallbackTimezone,
    });

    return `${formatter.format(new Date(slot.start_time))} - ${formatter.format(new Date(slot.end_time))}`;
};

const OptionButton = ({ active, title, subtitle, onClick, dataCy }) => {
    return (
        <button
            type="button"
            onClick={onClick}
            data-cy={dataCy}
            className={`w-full text-left p-4 rounded-lg border transition ${
                active
                    ? "border-blue-500 bg-blue-50"
                    : "border-gray-200 hover:border-blue-300 hover:bg-blue-50"
            }`}
        >
            <p className="font-medium text-gray-800">{title}</p>
            {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
        </button>
    );
};

const BackButton = ({ onClick }) => {
    return (
        <button
            data-cy="back-button"
            type="button"
            onClick={onClick}
            className="text-sm text-blue-600 hover:text-blue-700 hover:underline"
        >
            Назад
        </button>
    );
};

export const PublicBookingPage = () => {
    const { registrationSlug = "" } = useParams();

    const [step, setStep] = useState("service");
    const [serviceId, setServiceId] = useState("");
    const [branchId, setBranchId] = useState("");
    const [employeeId, setEmployeeId] = useState("");
    const [date, setDate] = useState(today());
    const [selectedSlot, setSelectedSlot] = useState(null);
    const [fieldError, setFieldError] = useState({});
    const [client, setClient] = useState({
        name: "",
        surname: "",
        phone: "",
        email: "",
        tg_username: "",
        comment: "",
    });

    const servicesQuery = useQuery({
        queryKey: ["public-services", registrationSlug],
        queryFn: () => getPublicServices(registrationSlug),
        enabled: Boolean(registrationSlug),
    });

    const branchesQuery = useQuery({
        queryKey: ["public-branches", registrationSlug, serviceId],
        queryFn: () => getPublicBranches({ registrationSlug, serviceId }),
        enabled: Boolean(registrationSlug && serviceId),
    });

    const employeesQuery = useQuery({
        queryKey: ["public-employees", registrationSlug, serviceId, branchId],
        queryFn: () => getPublicEmployees({ registrationSlug, serviceId, branchId }),
        enabled: Boolean(registrationSlug && serviceId && branchId),
    });

    const slotsQuery = useQuery({
        queryKey: ["public-slots", registrationSlug, serviceId, branchId, employeeId, date],
        queryFn: () => getPublicSlots({ registrationSlug, serviceId, branchId, employeeId, date }),
        enabled: Boolean(registrationSlug && serviceId && branchId && employeeId && date),
    });

    const selectedService = useMemo(() => {
        return servicesQuery.data?.find((service) => String(service.id) === String(serviceId));
    }, [serviceId, servicesQuery.data]);

    const selectedBranch = useMemo(() => {
        return branchesQuery.data?.find((branch) => String(branch.id) === String(branchId));
    }, [branchId, branchesQuery.data]);

    const selectedEmployee = useMemo(() => {
        return employeesQuery.data?.find((employee) => String(employee.id) === String(employeeId));
    }, [employeeId, employeesQuery.data]);

    const currentIndex = steps.indexOf(step);
    const pageError = servicesQuery.error?.message
        || branchesQuery.error?.message
        || employeesQuery.error?.message
        || slotsQuery.error?.message;

    const goBack = () => {
        const previousStep = steps[Math.max(currentIndex - 1, 0)];
        setStep(previousStep);
        setFieldError({});
    };

    const resetAfterService = () => {
        setBranchId("");
        setEmployeeId("");
        setSelectedSlot(null);
    };

    const resetAfterBranch = () => {
        setEmployeeId("");
        setSelectedSlot(null);
    };

    const bookingMutation = useMutation({
        mutationFn: createPublicAppointment,
        onSuccess: () => {
            setFieldError({});
            void slotsQuery.refetch();
        },
        onError: (error) => {
            setFieldError((prev) => ({ ...prev, submit: error.message }));
        },
    });

    const handleClientChange = (event) => {
        const { name, value } = event.target;
        setClient((prev) => ({ ...prev, [name]: value }));

        if (fieldError[name] || fieldError.submit) {
            setFieldError((prev) => ({ ...prev, [name]: null, submit: null }));
        }
    };

    const isValidPhone = (phone) => {
        if (!phone) return false;
        const digitsOnly = phone.replace(/\D/g, '');
        return digitsOnly.length === 11 && (['7', '8', '9'].includes(digitsOnly[0]));
    };

    const isValidEmail = (email) => {
        if (!email || email.length > 254) return false;
        const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
        return emailRegex.test(email);
};

    const validateClient = () => {
        const errors = {};

        if (!client.name.trim()) errors.name = "Введите имя";
        if (!client.surname.trim()) errors.surname = "Введите фамилию";

        if (!client.phone.trim()) { 
            errors.phone = "Введите телефон";
        } else if (!isValidPhone(client.phone)) {
            errors.phone = "Неверный формат телефона";
        }

        if (client.email && client.email.trim() && !isValidEmail(client.email)) {
            errors.email = "Неверный формат email";
        }

        setFieldError(errors);
        return Object.keys(errors).length === 0;
    };

    const handleSubmit = (event) => {
        event.preventDefault();
        if (!validateClient() || bookingMutation.isPending || !selectedSlot) return;

        bookingMutation.mutate({
            registrationSlug,
            payload: {
                client: {
                    name: client.name.trim(),
                    surname: client.surname.trim(),
                    phone: client.phone.trim(),
                    email: client.email.trim() || undefined,
                    tg_username: client.tg_username.trim() || undefined,
                },
                branch_id: Number(branchId),
                employee_id: Number(employeeId),
                service_id: Number(serviceId),
                start_time: selectedSlot.start_time,
                comment: client.comment.trim() || undefined,
            },
        });
    };

    const googleCalendarExportMutation = useMutation({
        mutationFn: async () => {
            const appointment = bookingMutation.data;
            
            return getGoogleOAuthUrl(appointment.google_calendar_export_url);
        },

        onSuccess: (data) => {
            if (!data?.url) {
                setFieldError((prev) => ({
                    ...prev,
                    submit: 'Google OAuth URL no exist',
                }));

                return;
            }

            globalThis.location.href = data.url;
        },

        onError: (error) => {
            setFieldError((prev) => ({
                ...prev,
                submit:
                    error instanceof Error
                        ? error.message
                        : 'Google Calendar export error',
            }));
        },
    });

    const handleGoogleCalendarExport = () => {
        googleCalendarExportMutation.mutate();
    };

    if (bookingMutation.isSuccess) {
        const appointment = bookingMutation.data;
        const displaySlot = selectedSlot || {
            start_time: appointment.start_time,
            end_time: appointment.end_time,
        };

        return (
            <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
                <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-7">
                    <div className="text-center mb-7">
                        <h1 className="text-gray-800 text-2xl font-bold">Запись создана</h1>
                        <p className="text-gray-500 text-sm mt-2">Заявка сохранена со статусом pending</p>
                    </div>

                    <div className="space-y-3 text-sm text-gray-700">
                        <p><span className="font-medium">Номер:</span> {appointment.appointment_id}</p>
                        <p><span className="font-medium">Услуга:</span> {appointment.service.name}</p>
                        <p><span className="font-medium">Филиал:</span> {appointment.branch.name}</p>
                        <p><span className="font-medium">Мастер:</span> {appointment.employee.name} {appointment.employee.surname}</p>
                        <p><span className="font-medium">Время:</span> {formatSlot(displaySlot)}</p>
                    </div>

                    <Button
                        className="mt-7"
                        onClick={() => {
                            bookingMutation.reset();
                            setStep("service");
                            setServiceId("");
                            resetAfterService();
                        }}
                    >
                        Создать еще одну запись
                    </Button>

                    <Button
                        data-cy="google-calendar-button"
                        className="mt-4"
                        onClick={handleGoogleCalendarExport}
                    >
                        Выгрузить в Google Календарь
                    </Button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
            <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-7">
                <div className="mb-6">
                    <p className="text-sm text-gray-500 mb-2">Шаг {currentIndex + 1} из {steps.length}</p>
                    <h1 className="text-gray-800 text-2xl font-bold">Онлайн-запись</h1>
                    <p className="text-gray-500 text-sm mt-2">{stepTitles[step]}</p>
                </div>

                {(pageError || fieldError.submit) && (
                    <div className="mb-5 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
                        {fieldError.submit || pageError}
                    </div>
                )}

                {step !== "service" && (
                    <div className="mb-5 p-4 bg-slate-50 border border-gray-200 rounded-xl text-sm text-gray-600">
                        {selectedService && <p>Услуга: {selectedService.name}</p>}
                        {selectedBranch && <p>Филиал: {selectedBranch.name}</p>}
                        {selectedEmployee && <p>Мастер: {selectedEmployee.name} {selectedEmployee.surname}</p>}
                        {selectedSlot && <p>Время: {formatSlot(selectedSlot)}</p>}
                    </div>
                )}

                {step === "service" && (
                    <div className="space-y-3">
                        {servicesQuery.isLoading && (
                            <div className="p-4 border border-gray-200 rounded-xl text-sm text-gray-500 bg-gray-50">
                                Загружаем услуги...
                            </div>
                        )}

                        {servicesQuery.data?.map((service) => (
                            <OptionButton
                                dataCy="service-option-button"
                                key={service.id}
                                active={String(service.id) === String(serviceId)}
                                title={service.name}
                                subtitle={`${service.duration_minutes} мин · ${formatMoney(service.price)}`}
                                onClick={() => {
                                    setServiceId(String(service.id));
                                    resetAfterService();
                                }}
                            />
                        ))}

                        <Button 
                            data-cy="go-to-branch-button"
                            className="mt-5"
                            loading={false}
                            onClick={() => serviceId && setStep("branch")}
                        >
                            Выбрать филиал
                        </Button>
                    </div>
                )}

                {step === "branch" && (
                    <div className="space-y-3">
                        {branchesQuery.isLoading && (
                            <div className="p-4 border border-gray-200 rounded-xl text-sm text-gray-500 bg-gray-50">
                                Загружаем филиалы...
                            </div>
                        )}

                        {branchesQuery.data?.map((branch) => (
                            <OptionButton
                                dataCy="branch-option-button"
                                key={branch.id}
                                active={String(branch.id) === String(branchId)}
                                title={branch.name}
                                subtitle={branch.address}
                                onClick={() => {
                                    setBranchId(String(branch.id));
                                    resetAfterBranch();
                                }}
                            />
                        ))}

                        <div className="flex items-center gap-3 mt-5">
                            <div className="w-28"><BackButton onClick={goBack} /></div>
                            <Button 
                                data-cy="go-to-master-button"
                                onClick={() => branchId && setStep("employee")}>
                                Выбрать мастера
                            </Button>
                        </div>
                    </div>
                )}

                {step === "employee" && (
                    <div className="space-y-3">
                        {employeesQuery.isLoading && (
                            <div className="p-4 border border-gray-200 rounded-xl text-sm text-gray-500 bg-gray-50">
                                Загружаем мастеров...
                            </div>
                        )}

                        {employeesQuery.data?.map((employee) => (
                            <OptionButton
                                dataCy="master-option-button"
                                key={employee.id}
                                active={String(employee.id) === String(employeeId)}
                                title={`${employee.name} ${employee.surname}`}
                                subtitle={employee.position}
                                onClick={() => {
                                    setEmployeeId(String(employee.id));
                                    setSelectedSlot(null);
                                }}
                            />
                        ))}

                        <div className="flex items-center gap-3 mt-5">
                            <div className="w-28"><BackButton onClick={goBack} /></div>
                            <Button 
                                data-cy="go-to-time-button"
                                onClick={() => employeeId && setStep("time")}>
                                Выбрать дату
                            </Button>
                        </div>
                    </div>
                )}

                {step === "time" && (
                    <div>
                        <Input
                            data-cy="date-input"
                            label="Дата"
                            type="date"
                            name="date"
                            value={date}
                            min={today()}
                            onChange={(event) => {
                                setDate(event.target.value);
                                setSelectedSlot(null);
                            }}
                            required
                        />

                        <div className="mb-5">
                            <p className="block text-gray-600 text-sm font-medium mb-2">Время</p>

                            {slotsQuery.isLoading && (
                                <div className="p-4 border border-gray-200 rounded-xl text-sm text-gray-500 bg-gray-50">
                                    Загружаем свободное время...
                                </div>
                            )}

                            {!slotsQuery.isLoading && slotsQuery.data?.length === 0 && (
                                <div className="p-4 border border-gray-200 rounded-xl text-sm text-gray-500 bg-gray-50">
                                    На выбранную дату свободных слотов нет
                                </div>
                            )}

                            {slotsQuery.data?.length > 0 && (
                                <div className="grid grid-cols-2 gap-3">
                                    {slotsQuery.data.map((slot) => (
                                        <button
                                            data-cy="time-select-button"
                                            key={slot.start_time}
                                            type="button"
                                            onClick={() => setSelectedSlot(slot)}
                                            className={`p-3 rounded-lg border text-sm font-medium transition ${
                                                selectedSlot?.start_time === slot.start_time
                                                    ? "border-blue-500 bg-blue-50 text-blue-700"
                                                    : "border-gray-300 hover:border-blue-300 hover:bg-blue-50 text-gray-700"
                                            }`}
                                        >
                                            {formatSlot(slot)}
                                        </button>
                                    ))}
                                </div>
                            )}
                        </div>

                        <div className="flex items-center gap-3 mt-5">
                            <div className="w-28"><BackButton onClick={goBack} /></div>
                            <Button 
                                data-cy="go-to-data-button"
                                onClick={() => selectedSlot && setStep("client")}>
                                Ввести данные
                            </Button>
                        </div>
                    </div>
                )}

                {step === "client" && (
                    <form onSubmit={handleSubmit} noValidate>
                        <Input
                            data-cy="data-name-input"
                            label="Имя"
                            name="name"
                            value={client.name}
                            onChange={handleClientChange}
                            placeholder="Иван"
                            error={fieldError.name}
                        />

                        <Input
                            data-cy="data-surname-input"
                            label="Фамилия"
                            name="surname"
                            value={client.surname}
                            onChange={handleClientChange}
                            placeholder="Петров"
                            error={fieldError.surname}
                        />

                        <Input
                            data-cy="data-phone-input"
                            label="Телефон"
                            name="phone"
                            value={client.phone}
                            onChange={handleClientChange}
                            placeholder="+79990000000"
                            error={fieldError.phone}
                        />

                        <Input
                            data-cy="data-email-input"
                            label="Email"
                            type="email"
                            name="email"
                            value={client.email}
                            onChange={handleClientChange}
                            placeholder="email@example.com"
                            required={false}
                            error={fieldError.email}
                        />

                        <Input
                            data-cy="data-telegram-input"
                            label="Telegram"
                            name="tg_username"
                            value={client.tg_username}
                            onChange={handleClientChange}
                            placeholder="username"
                            required={false}
                        />

                        <Input
                            data-cy="data-comment-input"
                            label="Комментарий"
                            name="comment"
                            value={client.comment}
                            onChange={handleClientChange}
                            placeholder="Дополнительная информация"
                            required={false}
                        />

                        <div className="flex items-center gap-3 mt-5">
                            <div className="w-28"><BackButton onClick={goBack} /></div>
                            <Button
                                data-cy="confirm-booking-button" 
                                type="submit" loading={bookingMutation.isPending}>
                                Записаться
                            </Button>
                        </div>
                    </form>
                )}
            </div>
        </div>
    );
};

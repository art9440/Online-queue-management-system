import { bookingApi, branchesApi } from "./axios";

const getErrorMessage = (error) => {
    if (error.response) {
        const status = error.response.status;
        const data = error.response.data;
        return data?.error || data?.message || data || `Ошибка ${status}: ${error.response.statusText}`;
    }

    if (error.request) {
        return "Сервер не отвечает. Проверьте, что сервисы запущены.";
    }

    return error.message || "Произошла неизвестная ошибка";
};

const request = async (call) => {
    try {
        const response = await call();
        return response.data;
    } catch (error) {
        throw new Error(getErrorMessage(error));
    }
};

export const getPublicServices = (registrationSlug) => {
    return request(() => branchesApi.get(`/public/${registrationSlug}/services`));
};

export const getPublicBranches = ({ registrationSlug, serviceId }) => {
    return request(() => branchesApi.get(`/public/${registrationSlug}/services/${serviceId}/branches`));
};

export const getPublicEmployees = ({ registrationSlug, serviceId, branchId }) => {
    return request(() => branchesApi.get(`/public/${registrationSlug}/services/${serviceId}/branches/${branchId}/employees`));
};

export const getPublicSlots = ({ registrationSlug, serviceId, branchId, employeeId, date }) => {
    return request(() => bookingApi.get(
        `/public/${registrationSlug}/services/${serviceId}/branches/${branchId}/employees/${employeeId}/slots`,
        { params: { date } }
    ));
};

export const createPublicAppointment = ({ registrationSlug, payload }) => {
    return request(() => bookingApi.post(`/public/${registrationSlug}/appointments`, payload));
};

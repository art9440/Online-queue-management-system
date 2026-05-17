import { registrationApi } from "./axios";

export const registerBusiness = async(data) => {
    try {
        const response = await registrationApi.post('/register', {
            email : data.email,
            password : data.password,
            business_name : data.businessName,
            business_type : data.businessType,
        });
        return response.data;
    } catch (error) {
        if (error.response) {
            const status = error.response.status;
            const data = error.response.data;
            const msg = data?.error || data?.message 
                || `Ошибка ${status}: ${error.response.statusText}`;
            throw new Error(msg);
        }

        if (error.request) {
            throw new Error('Сервер не отвечает. Проверьте соединение.');
        }
        throw new Error(error.message || 'Произошла неизвестная ошибка');
    }
}

export const verifyCode = async({ registration_id, code }) => {
    const response = await registrationApi.post('/verify', {
        registration_id, code});
    return response.data;
};

export const resendCode = async({ registration_id }) => {
    const response = await registrationApi.post('/resend', {registration_id});
    return response.data;
};

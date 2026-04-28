import { useMutation } from "@tanstack/react-query";
import axios from "axios";

export const api = axios.create({
    baseURL: '',
    withCredentials: true
});

export const registerBusiness = async(data) => {
    try {
        const response = await api.post('/api/register', {
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

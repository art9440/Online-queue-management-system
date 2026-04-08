import { React, useState } from "react";
import {useNavigate} from "react-router-dom";
import {registerBusiness} from "../api/registration";
import { Input } from "../components/Input";
import { Button } from "../components/Button";
import { Select } from "../components/Select";

export const RegistrationPage = () => {
    const navigate = useNavigate();

    const [fieldError, setFieldError] = useState({});
    const [error, setError] = useState('');
    const [loading, setLoading] = useState(false);
    
    const [formData, setFormData] = useState({
        businessName: "",
        email: "",
        password: "",
        businessType:"salon",
    });
    
    const businessTypes = [
        {value: 'salon', label: 'Салон красоты'},
        {value: 'barbershop', label: 'Барбершоп'},
        {value: 'grooming', label: 'Груминг студия'},
        {value: 'other', label:'Другое'}
    ];

    const validateField = (name, value) => {
        switch(name){
            case 'email':
                if (!value) return 'Введите email';
                const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                if (!emailRegex.test(value)) return 'Введите корректный email';
                return null;

            case 'password':
                if (!value) return 'Введите пароль';
                if (value.length < 6) return 'Минимум 6 символов';
                return null;
            
            case 'businessName':
                if (!value) return 'Введите название бизнеса';
                if (value.length < 2) return 'Минимум 2 символа';
                return null;
            
            default: return null;
        }
    }

    const handleBlur = (e) => {
        const {name, value} = e.target;
        const error = validateField(name, value);
        if (error) {
            setFieldError(prev => ({...prev, [name]: error}));
        }
    };

    const validateForm = () => {
        const errors = {};
        let isValid = true;

        Object.keys(formData).forEach(key => {
            const error = validateField(key, formData[key]);
            if (error) {
                errors[key] = error;
                isValid = false;
            }  
        });

        setFieldError(errors);
        return isValid;
    }

    const handleChange = (e) => {
        const {name, value} = e.target;
        setFormData(prev => ({...prev, [name]: value}));

        if (fieldError[name]) {
            setFieldError(prev=>({...prev, [name]: null}));
        }

        if (error) setError('');
    };

    const handleSubmit = async (e) => {
        e.preventDefault();         
        if (!validateForm()){
            return;
        }
        setError('');
        setLoading(true);

        try {
            const result = await registerBusiness(formData);
            localStorage.setItem('registration_id', result.registration_id);
            navigate('/verify');
        } catch (err) {
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <div className="flex h-screen w-full">

                <div className="flex-1 flex items-center justify-center">
                    <form className="w-full max-w-md px-8"
                    onSubmit={handleSubmit}
                    noValidate
                    >
                        {error && (
                            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
                                <div className="flex items-start gap-2">
                                    <span>{error}</span>
                                </div>
                            </div>
                        )}

                        <h2 className="text-gray-800 text-3xl font-bold text-center mb-8">Регистрация</h2>

                        <Input 
                        label="Название бизнеса"
                        type="text"
                        name="businessName"
                        value={formData.businessName}
                        autoComplete="off"
                        onChange={handleChange}
                        onBlur={handleBlur}
                        placeholder="Введите название бизнеса"
                        required={true}
                        error={fieldError.businessName}
                        />

                        <Input 
                        label="Почта"
                        type="email"
                        name="email"
                        value={formData.email}
                        autoComplete="off"
                        onChange={handleChange}
                        onBlur={handleBlur}
                        placeholder="Введите вашу почту"
                        required={true}
                        error={fieldError.email}
                        />

                        <Input 
                        label="Пароль"
                        type="password"
                        name="password"
                        value={formData.password}
                         autoComplete="new-password"
                        onChange={handleChange}
                        onBlur={handleBlur}
                        placeholder="Введите пароль"
                        required={true}
                        error={fieldError.password}
                        />  

                        <div className="mb-8">
                            <label className="block text-gray-600 text-sm
                            font-medium mb-1">Тип бизнеса</label>
                            <Select
                            name="businessType"
                            value={formData.businessType}
                            onChange={handleChange}
                            required={true}
                            businessTypes={businessTypes}
                            />
                        </div>

                        <Button
                        type="submit"
                        loading={loading}> 
                        Зарегистрироваться
                        </Button>

                    </form>
                </div>
            </div>
        </div>
    )
}
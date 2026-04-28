import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { registerBusiness } from "../api/registration";
import { Input } from "../components/ui/Input";
import { Button } from "../components/ui/Button";
import { Select } from "../components/ui/Select";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export const RegistrationPage = () => {
  const navigate = useNavigate();

  const [fieldError, setFieldError] = useState({});
  const [formData, setFormData] = useState({
    businessName: "",
    email: "",
    password: "",
    businessType: "salon",
  });

  const businessTypes = [
    { value: "salon", label: "Салон красоты" },
    { value: "barbershop", label: "Барбершоп" },
    { value: "grooming", label: "Груминг студия" },
    { value: "other", label: "Другое" },
  ];

  const validateField = (name, value) => {
    switch (name) {
      case "email":
        if (!value) return "Введите email";
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(value)) return "Введите корректный email";
        return null;

      case "password":
        if (!value) return "Введите пароль";
        if (value.length < 6) return "Минимум 6 символов";
        return null;

      case "businessName":
        if (!value) return "Введите название бизнеса";
        if (value.length < 2) return "Минимум 2 символа";
        return null;
      default:
        return null;
    }
  };

  const handleBlur = (e) => {
    const { name, value } = e.target;
    const error = validateField(name, value);
    if (error) {
      setFieldError((prev) => ({ ...prev, [name]: error }));
    }
  };

  const validateForm = () => {
    const errors = {};
    let isValid = true;

    Object.keys(formData).forEach((key) => {
      const error = validateField(key, formData[key]);
      if (error) {
        errors[key] = error;
        isValid = false;
      }
    });

    setFieldError(errors);
    return isValid;
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));

    if (fieldError[name]) {
      setFieldError((prev) => ({ ...prev, [name]: null }));
    }

    if (fieldError.submit) {
      setFieldError((prev) => ({...prev, submit: null}));
    }
  };

  const mutation = useMutation({
    mutationFn: registerBusiness,
    onSuccess: (response, data) => {
      localStorage.setItem("REGISTRATION_ID", response.registration_id);
      navigate("/verify", {state: {email: data.email}});
      setFieldError((prev) => ({...prev, submit: null}));
    },
    onError: (error) => {
      setFieldError((prev) => ({ ...prev, submit: error.message}));
    }

  })

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) return;
    if (isPending) return;
    mutation.mutate(formData);
  };

  const { isPending, isError, error } = mutation;
  const submitError = fieldError.submit || error?.message;

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
        
        <div className="text-center mb-8">
          <h2 className="text-gray-800 text-2xl font-bold">Регистрация</h2>
          <p className="text-gray-500 text-sm mt-2">Создайте аккаунт для вашего бизнеса</p>
        </div>

        {submitError && (
          <div data-cy="error-message"
            className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
              <div className="flex items-start gap-2">
                <span>{submitError}</span>
              </div>
          </div>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <Input data-cy="business-name"
            label="Название бизнеса"
            type="text"
            name="businessName"
            value={formData.businessName}
            autoComplete="off"
            onChange={handleChange}
            onBlur={handleBlur}
            placeholder="Например: My Business"
            required={true}
            error={fieldError.businessName}
          />

          <Input data-cy="email"
            label="Почта"
            type="email"
            name="email"
            value={formData.email}
            autoComplete="off"
            onChange={handleChange}
            onBlur={handleBlur}
            placeholder="email@business.com"
            required={true}
            error={fieldError.email}
          />

          <Input data-cy="password"
            label="Пароль"
            type="password"
            name="password"
            value={formData.password}
            onChange={handleChange}
            onBlur={handleBlur}
            placeholder="Минимум 6 символов"
            required={true}
            error={fieldError.password}
          />

          <div className="mb-8">
            <label className="block text-gray-700 text-sm font-medium mb-2">
              Тип бизнеса
            </label>
            <div className="relative">
              <select data-cy="business-type"
                name="businessType"
                value={formData.businessType}
                onChange={handleChange}
                required
                className={`w-full px-4 py-3 border rounded-lg focus:outline-none focus:ring-2 transition appearance-none bg-white ${
                  fieldError.businessType
                    ? "border-red-300 focus:ring-red-200 focus:border-red-400 bg-red-50"
                    : "border-gray-300 focus:ring-indigo-200 focus:border-indigo-400"
                }`}
              >
                {businessTypes.map((type) => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </select>
              <div className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none text-gray-500">
                <svg
                  className="w-5 h-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 9l-7 7-7-7"
                  />
                </svg>
              </div>
            </div>
            {fieldError.businessType && (
              <p className="mt-1 text-sm text-red-600">{fieldError.businessType}</p>
            )}
          </div>

            <Button data-cy="submit"
            type="submit" 
            loading={isPending}>
                Зарегистрироваться
            </Button>

          <p className="mt-6 text-center text-sm text-gray-600">
            Уже есть аккаунт?{" "}
            <button
              type="button"
              onClick={() => navigate("/login")}
              className="text-indigo-600 hover:text-indigo-700 font-medium hover:underline"
            >
              Войти
            </button>
          </p>
        </form>
      </div>
    </div>
  );
};
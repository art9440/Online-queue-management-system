import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { Input } from "../components/ui/Input";
import { Button } from "../components/ui/Button";
import { useFormValidation } from "../hooks/useFormValidation";

export const LoginPage = () => {
  const navigate = useNavigate();
  const { user, login, loading, error, getRedirectPath } = useAuth();

  const validateField = (name, value) => {
    switch (name) {
      case "login":
        if (!value) return "Введите логин или email";
        return null;
      case "password":
        if (!value) return "Введите пароль";
        if (value.length < 6) return "Минимум 6 символов";
        return null;
      default:
        return null;
    }
  };

  const {
    formData,
    fieldError,
    setFieldError,
    handleBlur,
    handleChange,
    validateForm,
  } = useFormValidation(
    {
      login: "",
      password: "",
    },
    validateField
  );

  useEffect(() => {
    if (user && !loading) {
      navigate(getRedirectPath(user), { replace: true });
    }
  }, [user, loading, navigate, getRedirectPath]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) return;
    if (loading) return;
    try {
      const user = await login(formData.login, formData.password);
      navigate(getRedirectPath(user), { replace: true });
    } catch (error) {
      const msg = (error.message || "").toLowerCase();
      if (msg === "bad credentials"){
        setFieldError({
          login: "",
          password: "Некорректная почта или пароль",
          auth: true
        });
      } else {
        setFieldError((prev) => ({ ...prev, auth: error.message }));
      }
    }
  };

  const submitError = fieldError.auth ? error : null;

  return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
        
        <div className="text-center mb-8">
          <h2 className="text-gray-800 text-2xl font-bold">Вход</h2>
          <p className="text-gray-500 text-sm mt-2">Войдите в аккаунт</p>
        </div>

        {submitError && !fieldError.password &&(
          <div data-cy="error-message"
            className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
            <span>{submitError}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <Input data-cy="login"
            label="Почта"
            type="text"
            name="login"
            value={formData.login}
            autoComplete="username"
            onChange={handleChange}
            onBlur={handleBlur}
            required={true}
            error={fieldError.login === "" ? null : fieldError.login}
          />

          <Input data-cy="password"
            label="Пароль"
            type="password"
            name="password"
            value={formData.password}
            autoComplete="current-password"
            onChange={handleChange}
            onBlur={handleBlur}
            required={true}
            error={fieldError.password || (fieldError.login === null ? "has-error" : null)}
          />

{/*           <div className="flex items-center justify-between mb-6">
            <label className="flex items-center gap-2">
              <input type="checkbox" className="rounded border-gray-300" />
              <span className="text-sm text-gray-600">Запомнить меня</span>
            </label>
            <button 
              type="button"
              className="text-sm text-indigo-600 hover:text-indigo-700 hover:underline"
              onClick={() => navigate('/forgot-password')}
            >
              Забыли пароль?
            </button>
          </div> */}

          <Button data-cy="submit"
            type="submit"
            loading={loading}
            className="w-full"
          >
            Войти
          </Button>
        </form>

        <div className="mt-6 pt-6 border-t border-gray-100 text-center">
          <p className="text-sm text-gray-600">
            Нет аккаунта?{" "}
            <button
              type="button"
              onClick={() => {
                setFieldError({});
                navigate("/register");
              }}
              className="text-indigo-600 hover:text-indigo-700 font-medium hover:underline"
            >
              Зарегистрировать
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

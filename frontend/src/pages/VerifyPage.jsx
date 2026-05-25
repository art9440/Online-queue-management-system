import { useLocation, useNavigate } from "react-router-dom"
import { useState, useEffect, useRef } from "react";
import { resendCode, verifyCode } from "../api/registration";
import { useMutation } from "@tanstack/react-query";
import { Button } from "../components/ui/Button";
import { useAuth } from "../context/AuthContext";
const CODE_LENGTH = 6;

const CODE_INPUTS = [
  { key: "first-digit", index: 0 },
  { key: "second-digit", index: 1 },
  { key: "third-digit", index: 2 },
  { key: "fourth-digit", index: 3 },
  { key: "fifth-digit", index: 4 },
  { key: "sixth-digit", index: 5 },
];

export const VerifyPage = () => {
    const navigate = useNavigate();
    const location = useLocation();
    
    const { user, loading, getRedirectPath } = useAuth();
    useEffect(() => {
      if (user && !loading) {
        navigate(getRedirectPath(user), { replace: true });
      }
    }, [user, loading, navigate, getRedirectPath]);
    const email = location.state?.email || "ваш email";

    const [fieldError, setFieldError] = useState({});
    const [codeDigits, setCodeDigits] = useState(new Array(CODE_LENGTH).fill(""));
    const [resendTimer, setResendTimer] = useState(0);

    const inputRefs = useRef([]);
    const registrationId = localStorage.getItem("REGISTRATION_ID");

    const validateCode = (digits) => {
        const code = digits.join("");
        if (!code) return "Введите код подтверждения";
        if (code.length < CODE_LENGTH) return `Код должен содержать ${CODE_LENGTH} цифр`;
        return null;
    }

    const handleChange = (index, value) => {
        if (!/^\d*$/.test(value)) return;
        const newDigits = [...codeDigits];
        newDigits[index] = value;
        setCodeDigits(newDigits);

        if (fieldError.code) setFieldError((prev) => ({...prev, code: null }));
        if (fieldError.submit) setFieldError((prev) => ({...prev, submit: null}));

        if (value && index < CODE_LENGTH - 1){
            inputRefs.current[index + 1].focus();
        }
    };

    const handleKeyDown = (index, e) => {
        if (e.key === "Backspace" && !codeDigits[index] && index > 0){
            inputRefs.current[index - 1].focus();
        }
        if (e.key === "ArrowLeft" && index > 0){
            inputRefs.current[index - 1].focus();
        }
        if (e.key === "ArrowRight" && index < CODE_LENGTH - 1){
            inputRefs.current[index + 1].focus();
        }
    };

    const handlePaste = (e) => {
        e.preventDefault();
        const pastedData = e.clipboardData.getData('text');
        if (/^\d+$/.test(pastedData) && pastedData.length >= CODE_LENGTH) {
            const newDigits = pastedData.slice(0, CODE_LENGTH).split("");
            setCodeDigits(newDigits);
            inputRefs.current[CODE_LENGTH - 1].focus();

            if (fieldError.code) setFieldError((prev) => ({ ...prev, code: null }));
            if (fieldError.submit) setFieldError((prev) => ({ ...prev, submit: null }));
        }
    };

    const verifyMutation = useMutation({
        mutationFn: verifyCode,
        onSuccess: () => {
            localStorage.removeItem("REGISTRATION_ID");
            navigate("/login", { 
                state: { message: "Аккаунт подтверждён! Теперь вы можете войти." } 
            });
        },
        onError: (error) => {
        let msg = error.response?.data?.error 
                || error.response?.data?.message 
                || error.message 
                || "Ошибка проверки кода";
        if (msg === "redis: nil") {
          msg = "Время подтверждения истекло. Пожалуйста, зарегистрируйтесь заново.";
          localStorage.removeItem("REGISTRATION_ID");
        }     
        if (msg === "invalid code")
          msg = "Некорректный код"
        if (msg === "user with this email already exists") {
          msg = "Аккаунт с такой почтой уже существует";
        }
        setFieldError((prev) => ({ ...prev, submit: msg }));
        }
    });

    const resendMutation = useMutation({
        mutationFn: resendCode,
        onSuccess: () => {
            setResendTimer(60);
            setFieldError((prev) => ({ ...prev, submit: null, resend: "Код отправлен повторно" }));
        },
        onError: (error) => {
            const msg = error.response?.data?.error 
                    || error.response?.data?.message 
                    || "Не удалось отправить код";
            setFieldError((prev) => ({ ...prev, submit: msg }));
        }
    });

    useEffect(() => {
        if (resendTimer > 0) {
        const interval = setInterval(() => setResendTimer((prev) => prev - 1), 1000);
        return () => clearInterval(interval);
        }
    }, [resendTimer]);

    const handleSubmit = (e) => {
        e.preventDefault();

        const error = validateCode(codeDigits);
        if (error) {
            setFieldError((prev) => ({...prev, code: error }));
            return;
        }

        if (!registrationId) {
            setFieldError((prev) => ({ ...prev, submit: "Сессия истекла. Пожалуйста, зарегистрируйтесь снова." }));
            return;
        }

        verifyMutation.mutate({
            registration_id: registrationId,
            code: codeDigits.join("")
        });
    }

    const handleResend = () => {
        if (resendTimer > 0 || !registrationId || verifyMutation.isPending) return;
        resendMutation.mutate({ registration_id: registrationId });
    };

    const handleNavigateToLogin = () => {
      localStorage.removeItem("REGISTRATION_ID");
      navigate("/login", { replace: true });
    };

    const isPending = verifyMutation.isPending || resendMutation.isPending;
    const submitError = fieldError.submit;

    let resendText = "Отправить код еще раз";
    if (resendTimer > 0){
        resendText = `Отправить повторно ${resendTimer} сек`;
    } else if (resendMutation.isPending){
        resendText = "Отправка...";
    }

    if (!registrationId && !email) {
        return (
        <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
            <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-8 text-center">
                <h2 className="text-gray-800 text-2xl font-bold mb-4">Сессия истекла</h2>
                <p className="text-gray-600 mb-6">Мы не нашли данных о регистрации.</p>
                <Button onClick={() => navigate("/register", { replace: true })}>Вернуться к регистрации</Button>
            </div>
        </div>
        );
    }

    return (
    <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
        
        <div className="text-center mb-8">
          <p className="text-gray-700 text-lg font-medium">
            Мы отправили код на{' '}
            <span className="font-semibold text-indigo-600">{email || 'ваш email'}</span>
          </p>
        </div>

        {submitError && (
          <div data-cy="error-message"
            className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
            <span>{submitError}</span>
          </div>
        )}
        
        {fieldError.resend && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-xl text-green-700 text-sm text-center">
            <span>{fieldError.resend}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} noValidate>
          <p className="block text-gray-700 text-sm font-medium mb-3 text-center">
            Введите код из письма
          </p>
          
          <div 
            className="flex justify-center gap-2 mb-6"
            onPaste={handlePaste}
          >
            {CODE_INPUTS.map(({key, index}) => (
                <input
                key={`verify-digit-${index}`}
                ref={(el) => (inputRefs.current[index] = el)}
                type="text"
                inputMode="numeric"
                pattern="[0-9]*"
                maxLength={1}
                value={codeDigits[index]}
                onChange={(e) => handleChange(index, e.target.value)}
                onKeyDown={(e) => handleKeyDown(index, e)}
                disabled={isPending}
                data-cy={`verify-digit-${index}`}
                className={`w-12 h-14 text-center text-xl font-bold border rounded-lg 
                  focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500 
                  transition-all uppercase
                  ${fieldError.code 
                    ? "border-red-300 bg-red-50 text-red-900" 
                    : "border-gray-300 bg-white text-gray-900 hover:border-gray-400"
                  }
                  ${isPending ? "opacity-50 cursor-not-allowed" : ""}
                `}
              />
            )
            )}
              
          </div>

          {fieldError.code && (
            <p className="text-center text-sm text-red-600 mb-4">{fieldError.code}</p>
          )}

          <Button
            data-cy="verify-submit"
            type="submit"
            className="w-full"
            loading={verifyMutation.isPending}
            disabled={isPending}
          >
            {verifyMutation.isPending ? "Проверка..." : "Подтвердить"}
          </Button>
        </form>

        <div className="mt-4 text-center">
          <button
            data-cy="resend-code"
            type="button"
            onClick={handleResend}
            disabled={resendTimer > 0 || resendMutation.isPending || verifyMutation.isPending}
            className={`text-sm font-medium transition-colors ${
              resendTimer > 0 || resendMutation.isPending || verifyMutation.isPending
                ? "text-gray-400 cursor-not-allowed"
                : "text-indigo-600 hover:text-indigo-700 hover:underline"
            }`}
          >
            {resendText}
          </button>
        </div>

        <div className="mt-6 pt-4 border-t border-gray-100 text-center">
          <p className="text-sm text-gray-600">
            Уже есть аккаунт?{" "}
            <button
              type="button"
              onClick={ handleNavigateToLogin }
              className="text-indigo-600 hover:text-indigo-700 font-medium hover:underline"
            >
              Войти
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};
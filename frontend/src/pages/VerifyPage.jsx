import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { useMutation } from "@tanstack/react-query";
import { resendCode, verifyCode } from "../api/registration";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/Input";

export const VerifyPage = () => {
    const navigate = useNavigate();
    const location = useLocation();
    const registrationId = localStorage.getItem("REGISTRATION_ID") || "";
    const email = location.state?.email || "";
    const [code, setCode] = useState("");
    const [error, setError] = useState("");

    const verifyMutation = useMutation({
        mutationFn: verifyCode,
        onSuccess: () => {
            localStorage.removeItem("REGISTRATION_ID");
            navigate("/register", { replace: true });
        },
        onError: (err) => setError(err.message || "Не удалось подтвердить код"),
    });

    const resendMutation = useMutation({
        mutationFn: resendCode,
        onSuccess: () => setError("Код отправлен повторно"),
        onError: (err) => setError(err.message || "Не удалось отправить код повторно"),
    });

    const handleSubmit = (event) => {
        event.preventDefault();

        if (!registrationId) {
            setError("Не найден идентификатор регистрации");
            return;
        }

        if (!code.trim()) {
            setError("Введите код подтверждения");
            return;
        }

        setError("");
        verifyMutation.mutate({ registration_id: registrationId, code: code.trim() });
    };

    return (
        <div className="min-h-screen bg-linear-to-br from-slate-50 to-slate-100 flex items-center justify-center p-4">
            <div className="w-full max-w-md bg-white rounded-2xl shadow-lg border border-gray-100 p-8">
                <div className="text-center mb-8">
                    <h2 className="text-gray-800 text-2xl font-bold">Подтверждение</h2>
                    <p className="text-gray-500 text-sm mt-2">
                        {email ? `Введите код, отправленный на ${email}` : "Введите код из письма"}
                    </p>
                </div>

                {error && (
                    <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-xl text-red-700 text-sm">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} noValidate>
                    <Input
                        label="Код"
                        name="code"
                        value={code}
                        onChange={(event) => {
                            setCode(event.target.value);
                            if (error) setError("");
                        }}
                        placeholder="Введите код"
                    />

                    <Button type="submit" loading={verifyMutation.isPending}>
                        Подтвердить
                    </Button>

                    <button
                        type="button"
                        className="mt-5 w-full text-sm text-blue-600 hover:text-blue-700 hover:underline"
                        disabled={resendMutation.isPending}
                        onClick={() => {
                            if (!registrationId) {
                                setError("Не найден идентификатор регистрации");
                                return;
                            }

                            resendMutation.mutate({ registration_id: registrationId });
                        }}
                    >
                        Отправить код повторно
                    </button>
                </form>
            </div>
        </div>
    );
};

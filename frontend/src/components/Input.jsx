import { FieldError } from "./FieldError"

export const Input = ({
    label, type="text", name, value, autoComplete, onChange, onBlur, placeholder, required=true, error
}) => {
    return (
        <div className="mb-5">
            <label className="block text-gray-600 text-sm
                font-medium mb-1">
                {label}
            </label>
            <input className={`w-full p-3 border border-gray-300 rounded-md focus:ring-2 focus:ring-indigo-200 outline-none ${
                error ? 'border-red-500 bg-red-50 focus:border-red-400' : 'border-gray-300 focus:border-indigo-400'}`}
                type={type}
                name={name}
                value={value}
                autoComplete={autoComplete}
                onChange={onChange}
                onBlur={onBlur}
                placeholder={placeholder}
                required={required}
                />

            <FieldError error={error}/>
        </div>
    )
}
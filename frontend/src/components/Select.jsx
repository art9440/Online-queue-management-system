export const Select = ({name, value, onChange, required, businessTypes}) => {
    return (
        <select className="w-full px-4 py-3 border rounded-lg
        focus:outline-none focus:ring-2 focus:ring-indigo-500 bg-white"
        name={name}
        value={value}
        onChange={onChange}
        required={required}>
            {businessTypes.map(type => (
                <option key={type.value} value={type.value}>
                    {type.label}
                </option>
            ))}
        </select>
    )
}
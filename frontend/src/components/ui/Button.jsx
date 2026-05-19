export const Button = ({ 
    type = "button", 
    loading = false,    
    variant = "primary", 
    children, 
    onClick,
    className = "",
    ...props 
}) => {
    const baseStyles = "w-full p-3 rounded-md transition font-medium";
    
    const variants = {
        primary: loading 
            ? "bg-gray-300 cursor-not-allowed text-gray-500" 
            : "bg-blue-500 hover:bg-blue-600 cursor-pointer text-white",
        outline: "border border-gray-300 hover:bg-gray-50 text-gray-700 cursor-pointer bg-white",
        link: "bg-transparent hover:underline text-blue-600 cursor-pointer p-0 w-auto"
    };

    return (
        <button 
            className={`${baseStyles} ${variants[variant]} ${className}`}
            type={type}
            disabled={loading}
            onClick={onClick}
            {...props}
        >
            {loading ? "Загрузка..." : children}
        </button>
    );
};
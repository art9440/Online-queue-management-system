export const Button = ({type, loading, children}) => {
    return (
        <button className={`w-full p-3 rounded-md transition ${
            loading 
            ? 'bg-gray-300 cursor-not-allowed'
            : 'bg-blue-400 hover:bg-blue-500 cursor-pointer'
        }`}
        type={type}
        disabled={loading}>
        
            {loading ? 'Загрузка...' : children}
        </button>
    )
}

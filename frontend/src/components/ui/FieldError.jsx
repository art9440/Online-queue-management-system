export const FieldError = ({error}) => {
    if (!error) return null;
    return (
        <p data-cy="error-message" className="mt-1 text-sm text-red-600">{error}</p>
    );
}

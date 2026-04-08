export const registerBusiness = async(data) => {
    const response = await fetch('/api/register', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            email : data.email,
            password : data.password,
            business_name : data.businessName,
            business_type : data.businessType,
        }),
    });

    if (!response.ok){
        let errorMsg = '';
        try {
            const errorData = await response.json();
            errorMsg = errorData.error || errorData.message || errorMsg;
        } catch {
            errorMsg = `Ошибка ${response.status}: ${response.statusText}`;
        }
        throw new Error(errorMsg);
    }

    return await response.json();
}
import { useState } from "react";

export const useFormValidation = (initialFormData, validateField) => {
  const [fieldError, setFieldError] = useState({});
  const [formData, setFormData] = useState(initialFormData);

  const handleBlur = (event) => {
    const { name, value } = event.target;
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

  const handleChange = (event) => {
    const { name, value } = event.target;

    setFormData((prev) => ({ ...prev, [name]: value }));

    if (fieldError[name]) {
      setFieldError((prev) => ({ ...prev, [name]: null }));
    }
    
    if (fieldError.submit) {
      setFieldError((prev) => ({ ...prev, submit: null }));
    }

    if (fieldError.auth) {
      setFieldError((prev) => ({ ...prev, auth: null }));
    }
  };

  return {
    formData,
    fieldError,
    setFieldError,
    handleBlur,
    handleChange,
    validateForm,
  };
};
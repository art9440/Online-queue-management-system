import { branchesApi } from "./axios";

export const getBranches = async () => {
  const response = await branchesApi.get("/branches");
  return response.data;
};

export const getBranchEmployees = async (branchId) => {
  const response = await branchesApi.get(`/branches/${branchId}/employees`);
  return response.data;
};

export const getBranchBookings = async (branchId, date) => {
  const response = await branchesApi.get(`/branches/${branchId}/bookings`, {
    params: { date },
  });
  return response.data;
};

export const getBranchClients = async (branchId) => {
  const response = await branchesApi.get(`/branches/${branchId}/clients`);
  return response.data;
};

export const getServices = async () => {
  const response = await branchesApi.get("/services");
  return response.data;
};

export const getServiceBranchEmployees = async (serviceId, branchId) => {
  const response = await branchesApi.get(
    `/services/${serviceId}/branches/${branchId}/employees`
  );
  return response.data;
};

export const checkBranchesHealth = async () => {
  const response = await branchesApi.get("/health");
  return response.data;
};

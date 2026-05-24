export const ROLE_ID_TO_NAME = {
  1: 'super_admin',
  2: 'business_admin',
  3: 'manager',
  4: 'employee',
};

export const ROLE_NAME_TO_ID = {
  'super_admin': 1,
  'business_admin': 2,
  'manager': 3,
  'employee': 4,
};

export const getRoleName = (roleId) => ROLE_ID_TO_NAME[roleId] || null;

export const hasRole = (roleId, allowedRoles) => {
  const roleName = getRoleName(roleId);
  return roleName && allowedRoles.includes(roleName);
};
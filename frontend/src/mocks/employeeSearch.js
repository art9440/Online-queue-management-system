const normalizeSearchValue = (value) =>
  String(value || "").trim().toLowerCase();

const getSearchTokens = (value) =>
  normalizeSearchValue(value).split(/\s+/).filter(Boolean);

const getEmployeeSearchWords = (employee) =>
  [
    employee.surname,
    employee.name,
    employee.position,
    employee.branchName,
  ]
    .flatMap((value) => normalizeSearchValue(value).split(/\s+/))
    .filter(Boolean);

export const employeeMatchesSearch = (employee, search) => {
  const searchTokens = getSearchTokens(search);
  if (!searchTokens.length) return true;

  const employeeWords = getEmployeeSearchWords(employee);

  return searchTokens.every((token) =>
    employeeWords.some((word) => word.startsWith(token))
  );
};

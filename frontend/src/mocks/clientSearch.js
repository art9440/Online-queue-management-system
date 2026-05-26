const normalizeSearchValue = (value) =>
  String(value || "").trim().toLowerCase();

const getSearchTokens = (value) =>
  normalizeSearchValue(value).split(/\s+/).filter(Boolean);

const getClientSearchWords = (client) =>
  [
    client.surname,
    client.name,
    client.phone,
    client.email,
    client.branchName,
  ]
    .flatMap((value) => normalizeSearchValue(value).split(/\s+/))
    .filter(Boolean);

export const clientMatchesSearch = (client, search) => {
  const searchTokens = getSearchTokens(search);
  if (!searchTokens.length) return true;

  const clientWords = getClientSearchWords(client);

  return searchTokens.every((token) =>
    clientWords.some((word) => word.startsWith(token))
  );
};

const normalizeServiceName = (service) =>
  String(service.name || "").trim().toLowerCase();

const getServiceDuration = (service) =>
  Number(service.duration_minutes || 0);

const getServicePrice = (service) => Number(service.price || 0);

const getPriceRange = (services) => {
  const prices = services.map(getServicePrice);

  return {
    min: Math.min(...prices),
    max: Math.max(...prices),
  };
};

export const groupServicesForDisplay = (services = []) => {
  const groups = new Map();

  services.forEach((service) => {
    const groupKey = `${normalizeServiceName(service)}-${getServiceDuration(service)}`;
    const currentGroup = groups.get(groupKey);

    if (currentGroup) {
      currentGroup.variants.push(service);
      return;
    }

    groups.set(groupKey, {
      ...service,
      variants: [service],
      variantsCount: 1,
      priceRange: {
        min: getServicePrice(service),
        max: getServicePrice(service),
      },
    });
  });

  return Array.from(groups.values()).map((group) => {
    const priceRange = getPriceRange(group.variants);

    return {
      ...group,
      variantsCount: group.variants.length,
      priceRange,
    };
  });
};

import { Clock, Scissors } from "lucide-react";
import PropTypes from "prop-types";

const formatCurrency = (value) =>
  new Intl.NumberFormat("ru-RU", {
    style: "currency",
    currency: "RUB",
    maximumFractionDigits: 0,
  }).format(value || 0);

const getServicePriceLabel = (service) => {
  if (!service.priceRange || service.priceRange.min === service.priceRange.max) {
    return formatCurrency(service.price);
  }

  return `${formatCurrency(service.priceRange.min)} - ${formatCurrency(
    service.priceRange.max
  )}`;
};

export const ServiceCard = ({ service }) => {
  if (!service) return null;

  return (
    <article className="rounded-2xl border border-white/70 bg-white/90 p-4 shadow-sm shadow-slate-200/80 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100/70">
      <div className="flex items-start justify-between gap-4">
        <div className="flex min-w-0 items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-50 text-indigo-600">
            <Scissors size={18} />
          </div>
          <div className="min-w-0">
            <h3 className="font-semibold text-gray-900 truncate">
              {service.name}
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              {service.variantsCount > 1
                ? `${service.variantsCount} варианта по филиалам`
                : service.category || "Услуга"}
            </p>
          </div>
        </div>
        <span
          className={`shrink-0 rounded-full px-2.5 py-1 text-xs font-medium ${
            service.isActive === false
              ? "bg-gray-100 text-gray-600"
              : "bg-green-50 text-green-700"
          }`}
        >
          {service.isActive === false ? "Скрыта" : "Активна"}
        </span>
      </div>

      <div className="mt-4 flex items-center justify-between border-t border-gray-100 pt-3">
        <div>
          <p className="text-xs text-gray-500">Длительность</p>
          <p className="mt-1 flex items-center gap-1.5 text-sm font-semibold text-gray-900">
            <Clock size={14} />
            {service.duration_minutes || 0} мин
          </p>
        </div>
        <div className="text-right">
          <p className="text-xs text-gray-500">Стоимость</p>
          <p className="text-sm font-semibold text-gray-900">
            {getServicePriceLabel(service)}
          </p>
        </div>
      </div>
    </article>
  );
};

ServiceCard.propTypes = {
  service: PropTypes.shape({
    name: PropTypes.string,
    variantsCount: PropTypes.number,
    category: PropTypes.string,
    isActive: PropTypes.bool,
    duration_minutes: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    price: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    priceRange: PropTypes.shape({
      min: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
      max: PropTypes.oneOfType([PropTypes.number, PropTypes.string]),
    }),
  }),
};
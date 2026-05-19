const toneClasses = {
  indigo: "bg-indigo-50 text-indigo-600 border-indigo-100",
  blue: "bg-blue-50 text-blue-600 border-blue-100",
  green: "bg-green-50 text-green-600 border-green-100",
  amber: "bg-amber-50 text-amber-600 border-amber-100",
  rose: "bg-rose-50 text-rose-600 border-rose-100",
  slate: "bg-slate-50 text-slate-600 border-slate-100",
};

export const StatsCard = ({
  title,
  value,
  subtitle,
  icon: Icon,
  tone = "indigo",
  trend,
}) => {
  const toneClass = toneClasses[tone] || toneClasses.indigo;
  const normalizedTrend =
    typeof trend === "number"
      ? {
          value: trend,
          label: "к прошлой неделе",
        }
      : trend;

  return (
    <article className="rounded-2xl border border-white/70 bg-white/90 p-4 shadow-sm shadow-slate-200/80 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100/70">
      <div className="flex items-start justify-between gap-4">
        <div className="min-w-0">
          <p className="text-sm font-medium text-gray-500">{title}</p>
          <p className="mt-1 text-xl font-bold text-gray-900">{value}</p>
          {subtitle && <p className="mt-1 text-xs text-gray-400">{subtitle}</p>}
        </div>

        {Icon && (
          <div
            className={`shrink-0 w-10 h-10 rounded-xl border flex items-center justify-center ${toneClass}`}
            aria-hidden="true"
          >
            <Icon size={21} strokeWidth={2.2} />
          </div>
        )}
      </div>

      {normalizedTrend && (
        <div
          className={`mt-4 text-sm flex items-center gap-1 ${
            normalizedTrend.value >= 0 ? "text-green-600" : "text-red-600"
          }`}
        >
          <span className="font-medium">
            {normalizedTrend.value >= 0 ? "+" : ""}
            {normalizedTrend.value}%
          </span>
          <span className="text-gray-400">{normalizedTrend.label}</span>
        </div>
      )}
    </article>
  );
};

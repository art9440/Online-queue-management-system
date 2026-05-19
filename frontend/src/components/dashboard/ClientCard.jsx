import { Phone, Sparkles, User } from "lucide-react";

const getClientName = (client) =>
  [client.surname, client.name].filter(Boolean).join(" ") || client.name;

export const ClientCard = ({ client }) => {
  if (!client) return null;

  return (
    <article className="rounded-2xl border border-white/70 bg-white/90 p-4 shadow-sm shadow-slate-200/80 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100/70">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-indigo-50 text-indigo-600">
            <User size={18} />
          </div>
          <div className="min-w-0">
            <h3 className="font-semibold text-gray-900 truncate">
              {getClientName(client)}
            </h3>
            {client.phone && (
              <p className="mt-1 flex items-center gap-1.5 text-sm text-gray-500">
                <Phone size={14} />
                <span>{client.phone}</span>
              </p>
            )}
            {client.email && (
              <p className="mt-1 truncate text-sm text-gray-500">
                {client.email}
              </p>
            )}
            {client.branchName && (
              <p className="mt-1 text-xs font-medium text-indigo-600">
                {client.branchName}
              </p>
            )}
          </div>
        </div>
        {client.isNew && (
          <span className="inline-flex shrink-0 items-center gap-1 rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">
            <Sparkles size={12} />
            <span>Новый</span>
          </span>
        )}
      </div>

      {client.comment && (
        <p className="mt-4 border-t border-gray-100 pt-3 text-sm text-gray-500">
          {client.comment}
        </p>
      )}
    </article>
  );
};

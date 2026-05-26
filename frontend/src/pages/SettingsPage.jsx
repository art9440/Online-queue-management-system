import { useMemo, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Building2, Copy, ExternalLink, Link2, Store, X } from "lucide-react";
import { getBranches } from "../api/branches";
import { DashboardTopBar } from "../components/layouts/DashboardTopBar";
import { Sidebar } from "../components/layouts/Sidebar";
import { useAuth } from "../context/AuthContext";
import PropTypes from "prop-types";

const normalizeBranch = (branch) => ({
  id: branch.id,
  name: branch.name,
  address: branch.address,
});

const getIsManagerPath = () => globalThis.window.location.pathname.startsWith("/manager");

const DEMO_BOOKING_SLUG = "demo-business";

const getPublicBookingLink = (slug) => {
  if (!slug) return "";
  if (typeof window === "undefined") return `/public/${slug}`;
  return `${globalThis.window.location.origin}/public/${slug}`;
};

const BookingLinkModal = ({ bookingLink, isDemoLink, onClose }) => {
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    if (!bookingLink) return;

    await navigator.clipboard.writeText(bookingLink);
    setCopied(true);
    globalThis.window.setTimeout(() => setCopied(false), 1800);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/40 p-4">
      <section className="w-full max-w-xl overflow-hidden rounded-2xl bg-white shadow-2xl shadow-slate-950/20">
        <div className="flex items-start justify-between gap-4 border-b border-gray-100 p-5">
          <div>
            <p className="text-sm font-medium text-indigo-700">
              Онлайн-запись
            </p>
            <h2 className="mt-1 text-xl font-semibold text-gray-950">
              Ссылка для клиентов
            </h2>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="rounded-xl p-2 text-gray-400 transition hover:bg-gray-100 hover:text-gray-700"
            aria-label="Закрыть"
          >
            <X size={20} />
          </button>
        </div>

        <div className="space-y-4 p-5">
          {isDemoLink && (
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800">
              Endpoint для получения slug бизнеса пока не найден во фронтовых
              данных. Сейчас показана демо-ссылка из seed-миграции.
            </div>
          )}

          <div className="rounded-xl border border-indigo-100 bg-indigo-50/70 p-4">
            <p className="mb-2 text-sm font-medium text-indigo-700">
              Публичная ссылка
            </p>
            <div className="flex items-center gap-2 rounded-lg bg-white px-3 py-2 text-sm text-gray-700 ring-1 ring-indigo-100">
              <Link2 size={16} className="shrink-0 text-indigo-600" />
              <span className="min-w-0 flex-1 truncate">{bookingLink}</span>
            </div>
          </div>

          <div className="flex flex-col gap-2 sm:flex-row">
            <button
              type="button"
              onClick={handleCopy}
              className="inline-flex items-center justify-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-indigo-700"
            >
              <Copy size={16} />
              {copied ? "Скопировано" : "Скопировать"}
            </button>
            <a
              href={bookingLink}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center justify-center gap-2 rounded-lg border border-indigo-200 px-4 py-2.5 text-sm font-medium text-indigo-700 transition hover:bg-indigo-50"
            >
              <ExternalLink size={16} />
              Открыть
            </a>
          </div>
        </div>
      </section>
    </div>
  );
};

BookingLinkModal.propTypes = {
  bookingLink: PropTypes.string.isRequired,
  onClose: PropTypes.func.isRequired,
};

export const SettingsPage = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [bookingModalOpen, setBookingModalOpen] = useState(false);
  const { user, logout } = useAuth();
  const isManagerPath = getIsManagerPath();

  const { data: apiBranches = [], isLoading, isError } = useQuery({
    queryKey: ["branches", isManagerPath ? "manager" : "admin", "settings"],
    queryFn: getBranches,
    retry: 1,
  });

  const branches = useMemo(() => {
    const source = Array.isArray(apiBranches) ? apiBranches : [];
    return source.map(normalizeBranch);
  }, [apiBranches]);

  const businessName =
    user?.business_name ||
    user?.businessName ||
    user?.business ||
    (user?.business_id ? `Бизнес #${user.business_id}` : "Мой бизнес");
  const bookingSlug = user?.registration_slug || DEMO_BOOKING_SLUG;
  const bookingLink = getPublicBookingLink(bookingSlug);
  const isDemoLink = !user?.registration_slug;

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(99,102,241,0.14),transparent_32rem),linear-gradient(135deg,#f8fbff_0%,#eef6ff_48%,#f7fbf4_100%)]">
      <Sidebar
        role={isManagerPath ? "manager" : "admin"}
        isOpen={sidebarOpen}
        onClose={() => setSidebarOpen(false)}
        onLogout={logout}
      />

      <DashboardTopBar
        user={user}
        title="Настройки"
        onMenuClick={() => setSidebarOpen(true)}
        onLogout={logout}
      />

      <main className="p-4 lg:p-6">
        <div className="mx-auto max-w-5xl space-y-5">
          {isError && (
            <div className="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
              Не удалось загрузить филиалы для настроек.
            </div>
          )}

          <section className="overflow-hidden rounded-2xl border border-indigo-100 bg-white shadow-lg shadow-indigo-100/50">
            <div className="bg-indigo-700 p-6 text-white">
              <div className="flex items-center gap-3">
                <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-white/15">
                  <Building2 size={24} />
                </div>
                <div>
                  <p className="text-sm text-indigo-100">Название бизнеса</p>
                  <h2 className="text-2xl font-bold">{businessName}</h2>
                </div>
              </div>
            </div>

            <div className="grid gap-4 p-5 md:grid-cols-[260px_1fr]">
              <div className="space-y-4">
                <div className="rounded-2xl bg-indigo-50 p-5 ring-1 ring-indigo-100">
                  <p className="text-sm font-medium text-indigo-700">
                    Количество филиалов
                  </p>
                  <p className="mt-2 text-4xl font-bold text-indigo-950">
                    {branches.length}
                  </p>
                </div>

                <button
                  type="button"
                  onClick={() => setBookingModalOpen(true)}
                  className="w-full rounded-2xl bg-white p-5 text-left shadow-sm ring-1 ring-indigo-100 transition hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100/70"
                >
                  <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-indigo-600 text-white">
                    <Link2 size={20} />
                  </span>
                  <span className="mt-4 block text-sm font-medium text-indigo-700">
                    Онлайн-запись
                  </span>
                  <span className="mt-1 block font-semibold text-slate-950">
                    Ссылка для клиентов
                  </span>
                  <span className="mt-2 block truncate text-sm text-slate-500">
                    {bookingLink}
                  </span>
                </button>
              </div>

              <div>
                <div className="mb-3 flex items-center gap-2">
                  <Store size={18} className="text-indigo-600" />
                  <h3 className="font-semibold text-slate-950">Филиалы</h3>
                </div>

                {isLoading ? (
                  <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">
                    Загружаем филиалы...
                  </p>
                ) : (
                  <div className="grid gap-3 sm:grid-cols-2">
                    {branches.map((branch) => (
                      <div
                        key={branch.id}
                        className="rounded-xl border border-indigo-100 bg-indigo-50/60 p-4"
                      >
                        <p className="font-semibold text-slate-950">
                          {branch.name}
                        </p>
                        <p className="mt-1 text-sm text-slate-500">
                          {branch.address}
                        </p>
                      </div>
                    ))}

                    {!branches.length && (
                      <p className="rounded-xl bg-slate-50 p-4 text-sm text-slate-500">
                        Филиалы не найдены.
                      </p>
                    )}
                  </div>
                )}
              </div>
            </div>
          </section>
        </div>
      </main>

      {bookingModalOpen && (
        <BookingLinkModal
          bookingLink={bookingLink}
          isDemoLink={isDemoLink}
          onClose={() => setBookingModalOpen(false)}
        />
      )}
    </div>
  );
};

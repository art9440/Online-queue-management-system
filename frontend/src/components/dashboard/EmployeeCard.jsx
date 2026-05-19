import { UserRound } from "lucide-react";
import { EMPLOYEE_AVATAR_COLORS } from "../../constants/schedule";

const getAvatarColor = (employee) =>
  EMPLOYEE_AVATAR_COLORS[
    Number(employee.id || 0) % EMPLOYEE_AVATAR_COLORS.length
  ];

export const EmployeeCard = ({ employee, branchName, onClick }) => {
  if (!employee) return null;

  const fullName = [employee.surname, employee.name].filter(Boolean).join(" ");
  const initials = (fullName || employee.name || "?")
    .split(" ")
    .map((part) => part[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();

  return (
    <article
      onClick={onClick}
      className="cursor-pointer rounded-2xl border border-white/70 bg-white/90 p-4 shadow-sm shadow-slate-200/80 transition-all hover:-translate-y-0.5 hover:shadow-lg hover:shadow-indigo-100/70"
    >
      <div className="flex items-start gap-3">
        <div
          className={`flex h-11 w-11 shrink-0 items-center justify-center rounded-full text-sm font-semibold text-white ${getAvatarColor(employee)}`}
        >
          {initials ? initials : <UserRound size={18} />}
        </div>
        <div className="min-w-0 flex-1">
          <div className="flex items-start justify-between gap-3">
            <div className="min-w-0">
              <h3 className="font-semibold text-gray-900 truncate">
                {fullName || employee.name}
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                {employee.position || "Сотрудник"}
              </p>
              {branchName && (
                <p className="mt-1 text-xs font-medium text-indigo-600">
                  {branchName}
                </p>
              )}
            </div>
          </div>
        </div>
      </div>
    </article>
  );
};

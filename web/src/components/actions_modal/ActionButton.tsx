import type React from "react";

interface ActionButtonProps {
  icon: string;
  label: string;
  variant?: string;
  onClick: () => void;
}
const ActionButton: React.FC<ActionButtonProps> = ({
  icon,
  label,
  variant = "default",
  onClick 
}) => {
  const isError = variant === "error";
  return (
    <button className="flex flex-col items-center gap-1 group min-w-12"
      onClick={onClick}
    >
      <span className={`
        material-symbols-outlined text-[22px] transition-all duration-200
        ${isError ? 'text-red-400 group-hover:text-red-600' : 'text-slate-500 group-hover:text-blue-600'}
        group-hover:scale-110
      `}>
        {icon}
      </span>
      <span className={`
        text-[11px] font-medium tracking-wide transition-colors
        ${isError ? 'text-red-400 group-hover:text-red-600' : 'text-slate-500 group-hover:text-slate-900'}
      `}>
        {label}
      </span>
    </button>
  );
};

export default ActionButton;

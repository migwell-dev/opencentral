import type { Usage } from "../../data/types";

interface StorageBarProps {
  usage: Usage
}

const StorageBar: React.FC<StorageBarProps> = ({ usage }) => {
  const percent = usage?.percent ?? 0;
  const used = formatBytes(usage?.usedBytes ?? 0);
  const limit = usage?.limitBytes > 0 ? formatBytes(usage.limitBytes) : null;

  return (
    <div className="mt-auto p-4 bg-surface-container rounded-xl dark:bg-slate-800">
      <div className="flex justify-between items-center mb-2">
        <span className="text-xs font-bold text-on-surface-variant dark:text-blue-300">
          Storage
        </span>
        <span className="text-xs font-medium text-primary dark:text-blue-300">
          {limit ? `${Math.round(percent)}%` : used}
        </span>
      </div>
      <div className="w-full bg-outline-variant/30 h-1.5 rounded-full overflow-hidden">
        <div
          className="bg-primary-container h-full rounded-full dark:bg-blue-300 transition-all duration-500"
          style={{ width: limit ? `${Math.min(percent, 100)}%` : "0%" }}
        />
      </div>
      <p className="text-[10px] text-on-surface-variant mt-2 dark:text-blue-300">
        {limit ? `${used} of ${limit} used` : `${used} used — no limit set`}
      </p>
    </div>
  );
}

function formatBytes(bytes: number) {
  if (bytes === 0) return "0 B";
  const units = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`;
}

export default StorageBar;

import React from "react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "../ui/dropdown-menu";
import { Button } from "../ui/button";

interface SettingRowProps {
  label: string;
  hint?: string;
  value: number | string;
  onChange: (value: number | string) => void;
  onSave: (value: number) => void;
  saving: boolean;
  type?: string;
  min?: number;
  max?: number;
  isStorage?: boolean;
  storageUnit?: string;
  onUnitChange?: (unit: "GB" | "MB" | "KB") => void;
}

const SettingRow: React.FC<SettingRowProps> = ({
  label,
  hint,
  value,
  onChange,
  onSave,
  saving,
  type = "number",
  min,
  max,
  isStorage = false,
  storageUnit,
  onUnitChange,
}) => {
  return (
    <div>
      <label className="block text-sm font-medium text-slate-700 dark:text-slate-300 mb-1">
        {label}
      </label>

      {hint && <p className="text-xs text-slate-400 mb-2">{hint}</p>}

      <div className="flex gap-2">
        <input
          type={type}
          value={value}
          min={min}
          max={max}
          onChange={(e) => {
            const val = e.target.value;
            onChange(val === "" ? "" : Number(val));
          }}
          className="flex-1 px-3 py-2 text-sm rounded-lg border border-slate-200
          dark:border-slate-700 bg-slate-50 dark:bg-slate-800 text-slate-800 
          dark:text-slate-100 focus:outline-none focus:ring-2 
          focus:ring-primary/20 transition-all
          [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none
          [&::-webkit-inner-spin-button]:appearance-none"
        />

        {isStorage && storageUnit && onUnitChange && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline">{storageUnit}</Button>
            </DropdownMenuTrigger>

            <DropdownMenuContent>
              {["GB", "MB", "KB"].map((unit) => (
                <DropdownMenuItem
                  key={unit}
                  onClick={() => onUnitChange(unit as "GB" | "MB" | "KB")}
                >
                  {unit}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        )}

        <button
          onClick={() => onSave(Number(value))}
          disabled={saving}
          className="px-4 py-2 text-sm rounded-lg bg-primary-container
          dark:bg-slate-100 text-white dark:text-slate-900 hover:bg-on-surface
          dark:hover:bg-slate-200 disabled:opacity-50 transition-colors"
        >
          {saving ? "Saving..." : "Save"}
        </button>
      </div>
    </div>
  );
};

export default SettingRow;

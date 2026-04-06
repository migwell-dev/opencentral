import React, { useState, useEffect } from "react";
import SettingRow from "./SettingsRow";
import { toast } from "sonner";

interface SettingsModalProps {
  setOpenSettings: (value: boolean) => void
  refreshUsage: () => void
}

function toBytes(value: number | string, unit: string) {
  const numValue = typeof value === "string" ? parseFloat(value) : value;
  if (!numValue || isNaN(numValue)) return 0;
  let bytes = 0;

  switch (unit) {
    case "GB":
      bytes = (numValue * 1024 ** 3); break;
    case "MB":
      bytes = (numValue * 1024 ** 2); break;
    case "KB":
      bytes = (numValue * 1024); break;
    default:
      return value;
  }
  return Math.round(bytes);
}

const SettingsModal: React.FC<SettingsModalProps> = ({ setOpenSettings, refreshUsage }) => {
  const [settings, setSettings] = useState({
    port: 0 as number | string,
    storage_limit_bytes: 0 as number | string,
    storage_unit: "GB",
    trash_retention_days: 0 as number | string,
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState("");
  const [error, setError] = useState("");

  useEffect(() => {
    fetch("/api/settings")
      .then((r) => r.json())
      .then((data) => {
        const bytes = Number(data.storage_limit_bytes?.value ?? 0);

        let value: number = 0;
        let unit = "GB";

        if (bytes === 0) {
          value = 0;
        } else if (bytes >= 1024 ** 3) {
          value = (bytes / 1024 ** 3);
          unit = "GB";
        } else if (bytes >= 1024 ** 2) {
          value = (bytes / 1024 ** 2);
          unit = "MB";
        } else {
          value = (bytes / 1024);
          unit = "KB";
        }

        setSettings({
          port: data.port?.value ?? 0,
          storage_limit_bytes: value,
          storage_unit: unit,
          trash_retention_days: data.trash_retention_days?.value ?? 0,
        });

        setLoading(false);
      })
      .catch(() => {
        setError("Failed to load settings.");
        setLoading(false);
      });
  }, []);

  async function updateSetting(key: string, value: number | string) {
    setSaving(key);
    setError("");
    try {
      var valueString;
      if (typeof value !== 'string') {
        valueString = value.toString()
      }
      const res = await fetch("/api/settings/update", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ key,
          value: typeof value === 'string' ? value : valueString }),
      });
      if (!res.ok) {
        const msg = await res.text();
        setError(msg);
      } else {
        var keyName: string = "Port";
        switch (key) {
          case "port": break;
          case "storage_limit_bytes": keyName = "Storage Limit"; break;
          case "trash_retention_days": keyName = "Trash Retention Days"; break;
        }
        toast.success(keyName + " was updated successfully.")
        refreshUsage()
      }
    } catch {
      setError("Network error.");
    } finally {
      setSaving("");
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40"
      onClick={() => setOpenSettings(false)}
    >
      <div
        className="bg-white dark:bg-slate-900 rounded-2xl shadow-xl w-full max-w-md mx-4 p-6"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <h2 className="text-lg font-medium text-slate-800 dark:text-slate-100">
            Settings
          </h2>
          <button
            onClick={() => setOpenSettings(false)}
            className="p-1.5 rounded-full hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-400 transition-colors"
          >
            <span className="material-symbols-outlined text-xl">close</span>
          </button>
        </div>

        {loading ? (
          <p className="text-sm text-slate-400 text-center py-8">Loading...</p>
        ) : (
          <div className="space-y-6">
            {error && (
              <p className="text-sm text-red-500 bg-red-50 dark:bg-red-950/30 px-3 py-2 rounded-lg">
                {error}
              </p>
            )}

            {/* Port */}
            <SettingRow
              label="Port"
              hint="Requires a server restart to take effect."
              saving={saving === "port"}
              onSave={(val: number) => updateSetting("port", val)}
              value={settings.port}
              onChange={(v: number | string) => setSettings((s) => ({ ...s, port: v }))}
              type="number"
              min={1024}
              max={65535}
            />

            {/* Storage limit */}
            <SettingRow
              label="Storage limit"
              hint="Set to 0 for no limit."
              value={settings.storage_limit_bytes}
              onChange={(v) => setSettings((s) => ({ ...s, storage_limit_bytes: v }))}
              saving={saving === "storage_limit_bytes"}
              onSave={() =>
                updateSetting(
                  "storage_limit_bytes",
                  toBytes(settings.storage_limit_bytes, settings.storage_unit)
                )
              }
              type="number"
              min={0}
              isStorage
              storageUnit={settings.storage_unit}
              onUnitChange={(unit) => setSettings((s) => ({ ...s, storage_unit: unit }))}
            />

            {/* Trash retention */}
            <SettingRow
              label="Trash retention (days)"
              hint="Set to 0 to keep files in trash forever."
              saving={saving === "trash_retention_days"}
              onSave={(val) => updateSetting("trash_retention_days", val)}
              value={settings.trash_retention_days}
              onChange={(v) => setSettings((s) => ({ ...s, trash_retention_days: v }))}
              type="number"
              min={0}
            />
          </div>
        )}
      </div>
    </div>
  );
}


export default SettingsModal; 

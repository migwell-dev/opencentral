import { useState, useEffect, useCallback } from "react";
import type { Usage } from "../data/types";

export function useStorageUsage() {
  const [usage, setUsage] = useState<Usage | null>(null);

  const fetchUsage = useCallback(() => {
    fetch("/api/storage/usage")
      .then((r) => r.json())
      .then((data: Usage) => setUsage(data))
      .catch(console.error);
  }, []);

  useEffect(() => {
    fetchUsage();
  }, [fetchUsage]);

  return { usage, refreshUsage: fetchUsage };
}

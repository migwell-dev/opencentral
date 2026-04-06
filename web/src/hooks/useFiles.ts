import { useCallback, useEffect, useState } from "react";
import type { FileItem } from "../data/types";

export function useFiles(currentPath: string, overrideUrl?: string) {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchFiles = useCallback(async (signal?: AbortSignal) => {
    setLoading(true);
    setError(null);
    try {
      const url = overrideUrl ?? `/api/files?path=${encodeURIComponent(currentPath)}`;
      const res = await fetch(url, { signal });
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      console.log(data);
      const rawFiles = Array.isArray(data) ? data : [];
      const filteredFiles = rawFiles.filter(
        (file: FileItem) => file.name && !file.name.startsWith(".")
      );
      setFiles(filteredFiles);
    } catch (err: any) {
      if (err.name === "AbortError") return;
      setError("Failed to load files");
      setFiles([]);
    } finally {
      setLoading(false);
    }
  }, [currentPath, overrideUrl]);

  useEffect(() => {
    const controller = new AbortController();
    fetchFiles(controller.signal);
    return () => controller.abort();
  }, [fetchFiles]);

  return { files, loading, error, refetch: () => fetchFiles() };
}

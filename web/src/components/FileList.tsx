import React, { useRef } from "react";
import type { FileItem } from "../data/types";

interface FileListProps {
  loading: boolean;
  files: FileItem[];
  currentPath: string;
  setSelectedFiles: (files: FileItem[]) => void;
  selectedFiles: FileItem[];
  setActionsOpen: (value: boolean) => void;
  setPreviewFile: (value: string) => void;
  selectMode: boolean;
  setSelectMode: (value: boolean) => void;
  currentView: string;
  refetch: () => void;
}

const LONG_PRESS_MS = 500;

const FileList: React.FC<FileListProps> = ({
  loading,
  files,
  currentPath,
  selectedFiles,
  setSelectedFiles,
  setActionsOpen,
  setPreviewFile,
  selectMode,
  setSelectMode,
  currentView,
  refetch
}) => {
  const pressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const startPress = (file: FileItem) => {
    pressTimer.current = setTimeout(() => {
      setSelectMode(true);
      setSelectedFiles([file]);
      setActionsOpen(true);
    }, LONG_PRESS_MS);
  };

  const cancelPress = () => {
    if (pressTimer.current) {
      clearTimeout(pressTimer.current);
      pressTimer.current = null;
    }
  };

  if (loading) return <p className="p-4 text-slate-500">Loading...</p>;
  if (files.length === 0) return <p className="p-4 text-slate-500">No files</p>;

  const handleToggleFile = (file: FileItem) => {
    const isSelected = selectedFiles.some((f) => f.name === file.name);
    if (isSelected) {
      const updated = selectedFiles.filter((f) => f.name !== file.name);
      setSelectedFiles(updated);
      if (updated.length === 0) setActionsOpen(false);
    } else {
      setSelectedFiles([...selectedFiles, file]);
      setActionsOpen(true);
    }
  };

  const getDisplayFiles = () => {
    let filtered = files.filter((f) => !f.isDir);

    if (currentView === "starred") {
      filtered = filtered.filter((f) => f.starred);
    }

    if (currentView === "recent") {
      filtered = [...filtered].sort(
        (a, b) =>
          new Date(b.last_opened).getTime() - new Date(a.last_opened).getTime()
      );
    }

    return filtered;
  };

  const displayFiles = getDisplayFiles();

  const emptyMessages: Record<string, string> = {
    starred: "No starred files",
    trash: "Trash is empty",
    recent: "No recently opened files",
  };

  if (displayFiles.length === 0) {
    return (
      <p className="p-4 text-slate-500">
        {emptyMessages[currentView] ?? "No files"}
      </p>
    );
  }

  return (
    <div className="py-4">
      <p className="px-4 mb-2 text-xs font-bold uppercase tracking-wider text-slate-400">
        Files
      </p>
      <div className="flex flex-col border-t border-slate-100 dark:border-slate-700">
        {displayFiles.map((file) => {
          const isSelected = selectedFiles.some((f) => f.name === file.name);
          return (
            <div
              key={file.name}
              className="group flex items-center px-4 py-3 border-b
                border-slate-100 dark:border-slate-700
                hover:bg-slate-50 dark:hover:bg-slate-800
                transition-colors select-none"
              onPointerDown={() => startPress(file)}
              onPointerUp={cancelPress}
              onPointerLeave={cancelPress}
            >
              {selectMode && (
                <div
                  className="mr-4 flex items-center justify-center cursor-pointer"
                  onClick={() => handleToggleFile(file)}
                >
                  <div
                    className={`w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all
                      ${isSelected
                        ? "border-blue-600 bg-blue-600"
                        : "border-slate-300 dark:border-slate-600 group-hover:border-slate-400"
                      }`}
                  >
                    <div
                      className={`w-2 h-2 rounded-full bg-white transition-transform
                        ${isSelected ? "scale-100" : "scale-0"}`}
                    />
                  </div>
                </div>
              )}
              <div
                className="flex items-center space-x-3 flex-1 cursor-pointer"
                onClick={() => {
                  if (selectMode) {
                    handleToggleFile(file);
                    return;
                  }
                  const fullPath =
                    currentPath === "/"
                      ? file.name
                      : currentPath.slice(1) + "/" + file.name;
                  setPreviewFile(fullPath);
                }}
              >
                <span className="text-xl">📄</span>
                <span className={`text-sm font-medium ${isSelected ? "text-blue-600 dark:text-blue-400" : "text-slate-700 dark:text-slate-200"}`}>
                  {file.name}
                </span>
                <span
                  onClick={(e) => {
                    e.stopPropagation();
                    const path = currentPath === "/"
                      ? file.name
                      : currentPath.slice(1) + "/" + file.name;
                    fetch(`/api/files/star?path=${encodeURIComponent(path)}`)
                      .then(res => res.text())
                      .then(() => refetch())
                      .catch(console.error);
                  }}
                  className={`material-icons-round p-1 opacity-0
            group-hover:opacity-100 transition-opacity cursor-pointer text-base
                  ${file.starred ? "text-primary-container" : "text-slate-400"}`}
                >
                  {file.starred ? "star" : "star_border"}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default FileList;

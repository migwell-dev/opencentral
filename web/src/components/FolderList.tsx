import React, { useRef } from "react";
import type { FileItem } from "../data/types";

interface FolderListProps {
  loading: boolean;
  files: FileItem[];
  openFolder: (name: string) => void;
  selectedFiles: FileItem[];
  setSelectedFiles: (files: FileItem[]) => void;
  setActionsOpen: (value: boolean) => void;
  setSelectMode: (value: boolean) => void;
  selectMode: boolean;
  currentView: string;
  currentPath: string;
  refetch: () => void;
}

const LONG_PRESS_MS = 500;

const FolderList: React.FC<FolderListProps> = ({
  loading,
  files,
  openFolder,
  selectedFiles,
  setSelectedFiles,
  setActionsOpen,
  setSelectMode,
  selectMode,
  currentView,
  currentPath,
  refetch,
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

  const handleToggleFolder = (folder: FileItem) => {
    const isSelected = selectedFiles.some((f) => f.name === folder.name);
    if (isSelected) {
      const updated = selectedFiles.filter((f) => f.name !== folder.name);
      setSelectedFiles(updated);
      if (updated.length === 0) setActionsOpen(false);
    } else {
      setSelectedFiles([...selectedFiles, folder]);
      setActionsOpen(true);
    }
  };

  if (loading) return <p className="p-4 text-slate-500">Loading...</p>;

  if (currentView === "trash" || currentView === "recent") return null;

  let folders = files.filter((f) => f.isDir);

  if (currentView === "starred") {
    folders = folders.filter((f) => f.starred);
  }

  if (currentView === "recent") {
    folders = [...folders].sort(
      (a, b) =>
        new Date(b.last_opened).getTime() - new Date(a.last_opened).getTime()
    );
  }

  if (folders.length === 0) return <p className="p-4 text-slate-500">No folders</p>;

  return (
    <div className="py-4">
      <p className="px-4 mb-2 text-xs font-bold uppercase tracking-wider text-slate-400">
        Folders
      </p>
      <div className="flex flex-col border-t border-slate-100 dark:border-slate-700">
        {folders.map((folder) => {
          const isSelected = selectedFiles.some((f) => f.name === folder.name);
          return (
            <div
              key={folder.name}
              className="group flex items-center px-4 py-3 border-b
                border-slate-100 dark:border-slate-700
                hover:bg-slate-50 dark:hover:bg-slate-800
                transition-colors"
              onPointerDown={() => startPress(folder)}
              onPointerUp={cancelPress}
              onPointerLeave={cancelPress}
            >
              {selectMode && (
                <div
                  className="mr-4 flex items-center justify-center cursor-pointer"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleToggleFolder(folder);
                  }}
                >
                  <div
                    className={`w-5 h-5 rounded-full border-2 flex items-center justify-center transition-all
                      ${isSelected
                        ? "border-blue-600 bg-blue-600"
                        : "border-slate-300 dark:border-slate-600 group-hover:border-slate-400 dark:group-hover:border-slate-500"
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
                onClick={() => !selectMode && openFolder(folder.name)}
              >
                <span className="text-xl">📁</span>
                <span
                  className={`text-sm font-medium
                    ${isSelected
                      ? "text-blue-600 dark:text-blue-400"
                      : "text-slate-700 dark:text-slate-200"
                    }`}
                >
                  {folder.name}
                </span>
                <span
                  onClick={(e) => {
                    e.stopPropagation();
                    const path = currentPath === "/"
                      ? folder.name
                      : currentPath.slice(1) + "/" + folder.name;
                    fetch(`/api/files/star?path=${encodeURIComponent(path)}`)
                      .then(res => res.text())
                      .then(() => refetch())
                      .catch(console.error);
                  }}
                  className={`material-icons-round p-1 opacity-0
            group-hover:opacity-100 transition-opacity cursor-pointer text-base
    ${folder.starred ? "text-primary-container" : "text-slate-400"}`}
                >
                  {folder.starred ? "star" : "star_border"}
                </span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default FolderList;

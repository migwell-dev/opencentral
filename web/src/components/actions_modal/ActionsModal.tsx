import { toast } from "sonner";
import type { FileItem } from "../../data/types";
import ActionButton from "./ActionButton";

interface ActionsModalProps {
  currentPath: string;
  setActionsOpen: (value: boolean) => void;
  selectedFiles: FileItem[];
  setSelectedFiles: (files: FileItem[]) => void;
  setSelectMode: (value: boolean) => void;
  onRefresh: () => void;
}

const ActionsModal: React.FC<ActionsModalProps> = ({
  currentPath,
  setActionsOpen,
  selectedFiles: files,
  setSelectedFiles,
  setSelectMode,
  onRefresh,
}) => {
  const handleClose = () => {
    setSelectMode(false);
    setSelectedFiles([]);
    setActionsOpen(false);
  };

  async function handleDownload() {
    if (files.length === 1 && !files[0].isDir) {
      const a = document.createElement("a");
      a.href = `/api/file?path=${encodeURIComponent(currentPath + files[0].name)}`;
      a.download = files[0].name;
      a.click();
      handleClose();
      toast.success("Download success!");
      return;
    }

    // Batch — request a zip from the server
    const res = await fetch("/api/download/batch", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ paths: files.map((f) => currentPath + f.name) }),
    });

    if (!res.ok) return;

    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "download.zip";
    a.click();
    URL.revokeObjectURL(url);
    handleClose();
    toast.success("Download success!");
  }

  async function handleRename() {
    if (files.length === 1) {
      const newName = prompt("New name:", files[0].name);
      if (!newName || newName === files[0].name) return;

      await fetch("/api/rename", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ oldPath: currentPath + files[0].name, newName: currentPath + newName }),
      });

      onRefresh();
      handleClose();
      toast.success("Rename success!");
      return;
    }

    const baseName = prompt("Base name for all selected files:");
    if (!baseName) return;

    await Promise.all(
      files.map((file, i) => {
        const ext = file.isDir ? "" : extOf(file.name);
        const stem = file.isDir ? "" : baseName.replace(/\.[^.]+$/, "");
        const newName = i === 0
          ? `${baseName}${ext}`
          : `${stem || baseName}_${i + 1}${ext}`;

        return fetch("/api/rename", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ oldPath: currentPath + file.name, newName: currentPath + newName }),
        });
      })
    );

    onRefresh();
    handleClose();
    toast.success("Rename success!");
  }

  async function handleDelete() {
    const label = files.length === 1 ? files[0].name : `${files.length} items`;
    if (!confirm(`Move ${label} to trash?`)) return;

    await Promise.all(
      files.map((file) =>
        fetch("/api/trash", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ source: currentPath + file.name }),
        })
      )
    );

    onRefresh();
    handleClose();
    toast.success("Deletion success!");
  }

  const fileCount = files.length;

  return (
    <div className="fixed top-8 left-1/2 -translate-x-1/2 z-50 
      flex items-center gap-6 px-4 py-2.5
      bg-white/80 backdrop-blur-md border border-white/40
      rounded-full shadow-[0_8px_30px_rgb(0,0,0,0.12)]">
      {/* Left Section: Status & Close */}
      <div className="flex items-center gap-3 border-r border-slate-200 pr-5">
        <button
          className="p-1 hover:bg-slate-100 rounded-full transition-colors group"
          onClick={handleClose}
        >
          <span className="material-symbols-outlined text-slate-500 group-hover:text-slate-800 text-xl">
            close
          </span>
        </button>
        <span className="text-sm font-semibold text-primary-container tracking-tight whitespace-nowrap">
          {fileCount === 1 ? "1 Item Selected" : `${fileCount} Items Selected`}
        </span>
      </div>
      {/* Right Section: Action Items */}
      <div className="flex items-center gap-8 px-2">
        <ActionButton icon="download" label="Download" onClick={handleDownload} />
        <ActionButton icon="edit" label="Rename" onClick={handleRename} />
        <ActionButton icon="delete" label="Delete" variant="error" onClick={handleDelete} />
      </div>
    </div>
  );
};


function extOf(filename: string): string {
  const i = filename.lastIndexOf(".");
  return i > 0 ? filename.slice(i) : "";
}

export default ActionsModal;

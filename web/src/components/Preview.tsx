import React, { useEffect, useState } from "react";
import type { FileItem } from "../data/types";

interface PreviewProps {
  file: string | FileItem;
  onClose: () => void;
}

const Preview: React.FC<PreviewProps> = ({ file, onClose }) => {
  const [visible, setVisible] = useState(false);

  const filePath = typeof file === "string" ? file : file?.name;
  const fileName = filePath?.split("/").pop() ?? "Untitled";
  const url = filePath ? `/api/file?path=${filePath}&t=${Date.now()}` : "";

  useEffect(() => {
    if (!filePath) return;
    requestAnimationFrame(() => setVisible(true));
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") handleClose();
    };
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [filePath]);

  const handleClose = () => {
    setVisible(false);
    setTimeout(onClose, 250);
  };

  if (!filePath) return null;

  return (
    <div
      onClick={handleClose}
      className={`fixed inset-0 z-50 flex items-center justify-center p-6 transition-all duration-300 ${visible ? "bg-black/80 backdrop-blur-md" : "bg-black/0 backdrop-blur-none"
        }`}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        className={`flex flex-col w-full max-w-5xl h-[85vh] rounded-2xl overflow-hidden bg-neutral-900 border border-white/10 shadow-2xl transition-all duration-300 ${visible ? "opacity-100 scale-100 translate-y-0" : "opacity-0 scale-95 translate-y-4"
          }`}
      >
        {/* Top bar */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-white/6 bg-white/3 shrink-0">
          <div className="flex items-center gap-3 min-w-0">
            {/* File icon */}
            <div className="w-8 h-8 rounded-lg bg-white/6 border border-white/8 flex items-center justify-center shrink-0">
              <span className="material-icons-round text-white/50 text-base">insert_drive_file</span>
            </div>

            <div className="min-w-0">
              <p className="text-sm font-semibold text-white/90 truncate leading-tight">{fileName}</p>
            </div>
          </div>

          <button
            onClick={handleClose}
            className="w-8 h-8 rounded-lg bg-white/5 border border-white/8 flex items-center justify-center text-white/40 hover:bg-red-500/15 hover:border-red-500/30 hover:text-red-300 transition-all shrink-0"
            title="Close (Esc)"
          >
            <svg className="w-3.5 h-3.5" viewBox="0 0 14 14" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round">
              <path d="M1 1l12 12M13 1L1 13" />
            </svg>
          </button>
        </div>

        {/* iframe */}
        <div className="flex-1 bg-neutral-950 overflow-hidden">
          <iframe src={url} title={fileName} className="w-full h-full border-none block" />
        </div>
      </div>
    </div>
  );
};

export default Preview;

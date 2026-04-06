import React, { useState } from "react";
import { FileUploader } from "react-drag-drop-files";

interface UploadFile {
  file: File;
  name: string;
  ext: string;
  progress: number;
  done: boolean;
  error?: string;
}

interface UploadModalProps {
  setUploading: (uploading: boolean) => void;
  refetch: () => void;
}

function getExt(name: string): string {
  const parts = name.split(".");
  return parts.length > 1 ? parts.pop()!.toUpperCase().slice(0, 4) : "FILE";
}

const UploadModal: React.FC<UploadModalProps> = ({ setUploading, refetch }) => {
  const [files, setFiles] = useState<UploadFile[]>([]);

  const handleFiles = (incoming: FileList | File | File[]) => {
    const arr: File[] =
      incoming instanceof FileList
        ? Array.from(incoming)
        : Array.isArray(incoming)
          ? incoming
          : [incoming];

    setFiles((prev) => {
      const existing = new Set(prev.map((f) => f.name));
      const newFiles: UploadFile[] = arr
        .filter((f) => !existing.has(f.name))
        .map((f) => ({
          file: f,
          name: f.name,
          ext: getExt(f.name),
          progress: 0,
          done: false,
        }));
      return [...prev, ...newFiles];
    });
  };

  const removeFile = (name: string) => {
    setFiles((prev) => prev.filter((f) => f.name !== name));
  };

  const uploadFile = (file: UploadFile) => {
    const formData = new FormData();
    formData.append("file", file.file);

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `/api/upload?path=`);

    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable) {
        const progress = Math.round((event.loaded / event.total) * 100);
        setFiles((prev) =>
          prev.map((f) => (f.name === file.name ? { ...f, progress } : f))
        );
      }
    };

    xhr.onload = () => {
      if (xhr.status === 200) {
        setFiles((prev) =>
          prev.map((f) =>
            f.name === file.name ? { ...f, done: true, progress: 100 } : f
          )
        );
        setUploading(false)
      } else {
        setFiles((prev) =>
          prev.map((f) =>
            f.name === file.name ? { ...f, error: xhr.statusText } : f
          )
        );
      }
    };

    xhr.onerror = () => {
      setFiles((prev) =>
        prev.map((f) =>
          f.name === file.name ? { ...f, error: "Upload failed" } : f
        )
      );
    };

    xhr.send(formData);
  };

  const uploadAll = () => {
    files.forEach((file) => {
      if (!file.done && !file.error) uploadFile(file);
    });
  };

  const closeModal = () => {
    setUploading(false);
    refetch();
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center
      bg-black/25 backdrop-blur-sm">
      <div className="bg-white w-full max-w-125 rounded-[20px] border 
        border-black/8 shadow-sm overflow-hidden flex flex-col">

        {/* Header */}
        <div className="flex items-center justify-between px-6 py-5 border-b 
          border-black/6">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 bg-blue-50 rounded-lg flex items-center
              justify-center">
              <svg className="w-4 h-4 text-blue-500" viewBox="0 0 24 24"
                fill="none" stroke="currentColor" strokeWidth="2"
                strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4" />
                <polyline points="17 8 12 3 7 8" />
                <line x1="12" y1="3" x2="12" y2="15" />
              </svg>
            </div>
            <h2 className="text-[15px] font-medium tracking-tight">
              Upload files
            </h2>
          </div>
          <button
            onClick={closeModal}
            className="w-7 h-7 rounded-lg border border-black/10 flex
            items-center justify-center text-gray-400 text-xs hover:bg-gray-50 
            transition-colors"
          >
            ✕
          </button>
        </div>

        {/* Body */}
        <div className="px-6 py-5 flex flex-col gap-5">

          {/* Dropzone */}
          <FileUploader handleChange={handleFiles} name="files" multiple>
            <div className="border-[1.5px] border-dashed border-gray-200 
              rounded-[14px] px-6 py-8 text-center cursor-pointer 
              hover:border-blue-300 hover:bg-blue-50/50 transition-all group">
              <div className="w-10 h-10 bg-gray-100 rounded-[10px] flex 
                items-center justify-center mx-auto mb-3 group-hover:bg-blue-100
                transition-colors">
                <svg className="w-5 h-5 text-gray-400 group-hover:text-blue-400
                  transition-colors" viewBox="0 0 24 24" fill="none"
                  stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"
                  strokeLinejoin="round">
                  <path d="M4 14.899A7 7 0 1115.71 8h1.79a4.5 4.5 0 012.5 8.242" />
                  <path d="M12 12v9" />
                  <path d="M16 16l-4-4-4 4" />
                </svg>
              </div>
              <p className="text-[13px] font-medium text-gray-800">
                Drop files here or click to browse
              </p>
              <span className="text-[12px] text-gray-400 mt-1 block">
                Supports any file type · Max 2 GB per file
              </span>
            </div>
          </FileUploader>

          {/* File list */}
          <div>
            <div className="flex items-center justify-between mb-2.5">
              <h3 className="text-[13px] font-medium text-gray-900">Queued files</h3>
              <span className="text-[11px] font-medium bg-gray-100 text-gray-500 
                px-2 py-0.5 rounded-full border border-black/6">
                {files.length} {files.length === 1 ? "file" : "files"}
              </span>
            </div>

            <div className="flex flex-col gap-2 max-h-64 overflow-y-auto">
              {files.length === 0 ? (
                <p className="text-center text-[12px] text-gray-400 py-4">
                  No files added yet
                </p>
              ) : (
                files.map((file) => (
                  <div
                    key={file.name}
                    className="bg-gray-50 border border-black/6 rounded-[10px]
                      px-3 py-2.5 flex items-center gap-2.5"
                  >
                    <div className="w-7.5 h-9 bg-white border border-black/8 rounded-md flex items-center justify-center text-[9px] font-semibold text-gray-400 font-mono shrink-0 tracking-wide">
                      {file.ext}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-[12px] font-medium text-gray-800 truncate mb-1.5">
                        {file.name}
                      </p>
                      <div className="flex items-center gap-2">
                        <div className="flex-1 h-0.75 bg-gray-200 rounded-full overflow-hidden">
                          <div
                            className={`h-full rounded-full transition-all duration-300 ${file.done
                              ? "bg-green-500"
                              : file.error
                                ? "bg-red-400"
                                : "bg-blue-500"
                              }`}
                            style={{ width: `${file.progress}%` }}
                          />
                        </div>
                        <span
                          className={`text-[11px] font-mono min-w-6 text-right ${file.done
                            ? "text-green-500"
                            : file.error
                              ? "text-red-400"
                              : "text-gray-400"
                            }`}
                        >
                          {file.done ? "✓" : file.error ? "!" : `${file.progress}%`}
                        </span>
                      </div>
                    </div>
                    <button
                      onClick={() => removeFile(file.name)}
                      className="w-5.5 h-5.5 rounded-md flex items-center justify-center text-[11px] text-gray-300 hover:bg-red-50 hover:text-red-400 transition-colors shrink-0"
                    >
                      ✕
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-6 py-4 border-t border-black/6 flex items-center justify-between">
          <button
            onClick={() => setFiles([])}
            className="text-[13px] font-medium text-gray-400 hover:text-gray-600 px-3 py-1.5 rounded-lg hover:bg-gray-50 transition-colors"
          >
            Clear all
          </button>
          <button
            onClick={uploadAll}
            className="flex items-center gap-1.5 text-[13px] font-medium bg-gray-900 text-white px-5 py-2 rounded-[10px] hover:bg-gray-700 transition-colors"
          >
            <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <polyline points="16 16 12 12 8 16" />
              <line x1="12" y1="12" x2="12" y2="21" />
              <path d="M20.39 18.39A5 5 0 0018 9h-1.26A8 8 0 103 16.3" />
            </svg>
            Upload all
          </button>
        </div>
      </div>
    </div>
  );
};

export default UploadModal;

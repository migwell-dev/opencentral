import { useState, useEffect } from "react";

interface UseDragUploadProps {
  currentPath: string;
  onUploadComplete?: () => void;
}

export const useDragUpload = ({ currentPath, onUploadComplete }: UseDragUploadProps) => {
  const [dragging, setDragging] = useState(false);

  useEffect(() => {
    const handleDragEnter = (e: DragEvent) => {
      e.preventDefault();
      setDragging(true);
    };

    const handleDragLeave = (e: DragEvent) => {
      e.preventDefault();
      setDragging(false);
    };

    const handleDrop = async (e: DragEvent) => {
      e.preventDefault();
      setDragging(false);

      const files = e.dataTransfer?.files;
      if (!files) return;

      for (const file of files) {
        const formData = new FormData();
        formData.append("file", file);

        await fetch(`/api/upload?path=${encodeURIComponent(currentPath)}`, {
          method: "POST",
          body: formData,
        });
      }

      onUploadComplete?.();
    };

    window.addEventListener("dragenter", handleDragEnter);
    window.addEventListener("dragover", handleDragEnter);
    window.addEventListener("dragleave", handleDragLeave);
    window.addEventListener("drop", handleDrop);

    return () => {
      window.removeEventListener("dragenter", handleDragEnter);
      window.removeEventListener("dragover", handleDragEnter);
      window.removeEventListener("dragleave", handleDragLeave);
      window.removeEventListener("drop", handleDrop);
    };
  }, [currentPath, onUploadComplete]);

  return { dragging };
};

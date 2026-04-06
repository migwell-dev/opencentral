import React from "react";

interface UploadOverlayProps {
  dragging: boolean;
}

const UploadOverlay: React.FC<UploadOverlayProps> = ({ dragging }) => {
  if (!dragging) return null;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.4)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        fontSize: 24,
        color: "#fff",
        zIndex: 9999,
        pointerEvents: "none",
      }}
    >
      Drop in your files to upload
    </div>
  );
};

export default UploadOverlay;

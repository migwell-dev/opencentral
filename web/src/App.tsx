import { useState } from "react";
import FileList from "./components/FileList"
import Breadcrumbs from "./components/Breadcrumbs";
import Preview from "./components/Preview";
import { useDragUpload } from "./hooks/useDragUpload";
import UploadOverlay from "./components/UploadOverlay";
import type { FileItem, ViewType } from "./data/types";
import UploadModal from "./components/side_panel/UploadModal";
import FolderList from "./components/FolderList";
import { usePersistedState } from "./hooks/usePersistedState"
import { useFiles } from "./hooks/useFiles";
import Header from "./components/header/Header";
import SettingsModal from "./components/settings/SettingsModal";
import { Toaster } from "sonner"
import SidePanel from "./components/side_panel/SidePanel";
import { useStorageUsage } from "./hooks/useStorageUsage";
import ActionsModal from "./components/actions_modal/ActionsModal";

export default function App() {
  const [previewFile, setPreviewFile] = useState<string | null>(null);
  const [actionsOpen, setActionsOpen] = useState<boolean>(false);
  const [currentPath, setCurrentPath] = usePersistedState("currentPath", "/");
  const [currentView, setCurrentView] = usePersistedState<ViewType>("currentView", "home");
  const [uploading, setUploading] = useState<boolean>(false);
  const [searchTerm, setSearchTerm] = useState<string>("");
  const [settingsOpen, setSettingsOpen] = useState<boolean>(false);
  const [selectedFiles, setSelectedFiles] = useState<FileItem[]>([]);
  const [selectMode, setSelectMode] = useState<boolean>(false);
  const trashUrl = currentView === "trash" ? "/api/files/trash" : undefined;
  const { files, loading, refetch } = useFiles(currentPath, trashUrl);

  const { dragging } = useDragUpload({
    currentPath,
    onUploadComplete: refetch,
  });

  const { usage, refreshUsage } = useStorageUsage();

  const handleContextMenu = (e: any) => {
    e.preventDefault();
    console.log("Custom right-click at:", e.pageX, e.pageY);
  };

  const openFolder = (name: string) => {
    setCurrentPath(prev =>
      prev === "/" ? `/${name}` : `${prev}/${name}`
    );
  };

  const goBack = () => {
    if (currentPath === "/") return;

    const parts = currentPath.split("/").filter(Boolean);
    parts.pop();

    setCurrentPath("/" + parts.join("/"));
  };

  const toggleSelectMode = () => {
    setSelectMode((prev) => {
      const next = !prev;

      if (!next) {
        setSelectedFiles([]);
        setActionsOpen(false);
      }

      return next;
    });
  };

  const onRefresh = () => {
    refetch();
    refreshUsage();
  }

  const filteredFiles = files.filter((file) =>
    file.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const displayFiles = searchTerm ? filteredFiles : files;

  return (
    <div
      onContextMenu={handleContextMenu}
      className="flex h-screen bg-background text-on-background overflow-hidden
      dark:bg-slate-800"
    >
      {/* Sidebar */}
      <SidePanel
        setCurrentPath={setCurrentPath}
        currentView={currentView}
        setCurrentView={setCurrentView}
        setUploading={setUploading}
        usage={usage!}
      />

      <Toaster />
      {uploading && <UploadModal setUploading={setUploading} refetch={refetch} />}
      {settingsOpen && <SettingsModal setOpenSettings={setSettingsOpen} refreshUsage={refreshUsage} />}
      {actionsOpen && <ActionsModal
        setSelectedFiles={setSelectedFiles}
        selectedFiles={selectedFiles}
        setSelectMode={setSelectMode}
        setActionsOpen={setActionsOpen}
        onRefresh={onRefresh}
        currentPath={currentPath}
      />}

      {/* Main content */}
      <div className="flex flex-col flex-1 overflow-hidden">
        <Header
          setSearchTerm={setSearchTerm}
          searchTerm={searchTerm}
          setOpenSettings={setSettingsOpen}
          selectMode={selectMode}
          toggleSelectMode={toggleSelectMode}
        />
        {currentView === "home" &&
          <Breadcrumbs currentPath={currentPath} setCurrentPath={setCurrentPath} />}

        {currentView === "recent" && (
          <span className="ml-5 mt-5 text-lg font-medium text-slate-700 dark:text-slate-200">Recent</span>
        )}
        {currentView === "starred" && (
          <span className="ml-5 mt-5 text-lg font-medium text-slate-700 dark:text-slate-200">Starred</span>
        )}
        {currentView === "trash" && (
          <span className="ml-5 mt-5 text-lg font-medium text-slate-700 dark:text-slate-200">Trash</span>
        )}

        <div className="flex-1 overflow-y-auto p-5">
          {currentPath !== "/" && (
            <button onClick={goBack} className="flex items-center">
              <span className="material-symbols-outlined text-sm 
                text-primary-container dark:text-blue-300">arrow_back_ios</span>
            </button>
          )}


          <FolderList
            loading={loading}
            files={displayFiles}
            openFolder={openFolder}
            selectedFiles={selectedFiles}
            setSelectedFiles={setSelectedFiles}
            setActionsOpen={setActionsOpen}
            selectMode={selectMode}
            setSelectMode={setSelectMode}
            currentView={currentView}
            currentPath={currentPath}
            refetch={refetch}
          />

          <FileList
            loading={loading}
            files={displayFiles}
            currentPath={currentPath}
            selectedFiles={selectedFiles}
            setSelectedFiles={setSelectedFiles}
            setActionsOpen={setActionsOpen}
            setPreviewFile={setPreviewFile}
            selectMode={selectMode}
            setSelectMode={setSelectMode}
            currentView={currentView}
            refetch={refetch}
          />
        </div>
      </div>

      {/* Overlays */}
      <UploadOverlay dragging={dragging} />
      {previewFile && (
        <Preview
          file={previewFile}
          onClose={() => {
            setPreviewFile(null);
            refetch();
          }}
        />
      )}
    </div>
  );
}

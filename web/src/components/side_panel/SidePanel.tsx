import React from "react";
import type { Usage, ViewType } from "../../data/types";
import ThemeToggle from "../side_panel/ThemeToggle";
import { Button } from "../ui/button";
import StorageBar from "./StorageBar";

interface SidePanelProps {
  currentView: ViewType;
  setCurrentView: (view: ViewType) => void;
  setCurrentPath: (path: string) => void;
  setUploading: (uploading: boolean) => void;
  usage: Usage
}

const navItems: { label: string; value: ViewType; icon: string }[] = [
  { label: "Home", value: "home", icon: "home" },
  { label: "Recent", value: "recent", icon: "schedule" },
  { label: "Starred", value: "starred", icon: "star" },
  { label: "Trash", value: "trash", icon: "delete" },
];

const SidePanel: React.FC<SidePanelProps> =
  ({ currentView,
    setCurrentView,
    setCurrentPath,
    setUploading, usage }) => {
    return (
      <aside className="hidden md:flex h-screen w-64 flex-col shrink-0 p-4
      space-y-2 bg-surface-container-low dark:bg-slate-900">

        {/* Logo */}
        <div className="flex items-center space-x-3 px-2 mb-8 mt-2">
          <div className="w-10 h-10 rounded-lg bg-primary-container flex 
          items-center justify-center">
            <span
              className="material-symbols-outlined text-white"
              style={{ fontVariationSettings: "'FILL' 1" }}
            >
              cloud_done
            </span>
          </div>
          <div>
            <h1 className="text-lg font-black text-primary dark:text-blue-300 font-headline">
              OpenCentral
            </h1>
            <p className="text-[10px] text-on-surface-variant/60 font-medium tracking-widest uppercase dark:text-white">
              Local-first Storage
            </p>
          </div>
        </div>

        <Button
          className="bg-primary-container"
          size={"lg"}
          onClick={() => setUploading(true)}
        >
          <span className="material-symbols-outlined dark:text-white">add</span>
          <span className="dark:text-white">New</span>
        </Button>

        <nav className="flex flex-col space-y-1">
          {navItems.map((item) => {
            const isActive = currentView === item.value;
            return (
              <button
                key={`${item.value}-${item.icon}`}
                onClick={() => { setCurrentView(item.value); setCurrentPath("/") }}
                className={`flex items-center space-x-3 px-4 py-2.5 rounded-lg transition-all duration-200 w-full text-left
                ${isActive
                    ? "bg-white dark:bg-slate-800 dark:text-blue-300 shadow-sm font-semibold text-primary-container"
                    : "text-slate-600 dark:text-slate-400 hover:bg-slate-200/50 dark:hover:bg-slate-800/50 hover:translate-x-1"
                  }`}
              >
                <span
                  className="material-symbols-outlined"
                  style={isActive ? { fontVariationSettings: "'FILL' 1" } : undefined}
                >
                  {item.icon}
                </span>
                <span className="font-body text-sm font-medium">{item.label}</span>
              </button>
            );
          })}
        </nav>

        <StorageBar usage={usage} />

        <div className="flex justify-end mt-3 pt-3 border-t border-outline-variant/30">
          <ThemeToggle />
        </div>
      </aside >
    );
  };

export default SidePanel;

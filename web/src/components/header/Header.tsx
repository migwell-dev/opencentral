import type React from "react";
import SearchBar from "./SearchBar";
import { Button } from "../ui/button";

interface HeaderProps {
  searchTerm: string;
  setSearchTerm: (value: string) => void;
  setOpenSettings: (value: boolean) => void;

  selectMode: boolean;
  toggleSelectMode: () => void;
}

const Header: React.FC<HeaderProps> = ({
  searchTerm,
  setSearchTerm,
  setOpenSettings,
  selectMode,
  toggleSelectMode,
}) => {
  return (
    <header className="sticky top-0 z-40 flex justify-between items-center w-full px-6 py-3 bg-[#f8f9ff] dark:bg-slate-950">
      <div className="flex flex-1 items-center max-w-2xl">
        <SearchBar searchTerm={searchTerm} setSearchTerm={setSearchTerm} />
      </div>
      <div className="flex items-center space-x-4">
        <Button className="p-2 rounded-full hover:bg-primary-container
          bg-on-primary-container dark:hover:bg-slate-800
          transition-colors duration-200 text-primary-container
          hover:text-white"
          onClick={() => setOpenSettings(true)}
        >
          <span className="material-symbols-outlined">settings</span>
        </Button>
        <Button
          onClick={toggleSelectMode}
          className={`
              px-4 py-2 rounded-xl text-sm font-medium transition-all
              ${selectMode
              ? "bg-blue-600 text-white"
              : "bg-slate-100 text-slate-700 hover:bg-slate-200"}
              `}
        >
          {selectMode ? "Exit Select" : "Select"}
        </Button>
      </div>
    </header>
  );
}

export default Header;

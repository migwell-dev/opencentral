import React from "react";

interface SearchBarProps {
  searchTerm: string;
  setSearchTerm: (value: string) => void;
}

const SearchBar: React.FC<SearchBarProps> = ({ searchTerm, setSearchTerm }) => {
  return (
    <div className="relative w-full mr-4">
      <span className="material-symbols-outlined absolute left-3 top-1/2
        -translate-y-1/2 text-on-surface-variant/50 text-xl
        dark:text-on-primary-container">
        search
      </span>

      <input
        className="w-full pl-10 pr-4 py-2 rounded-lg text-sm
        dark:text-on-primary-container"
        type="text"
        placeholder="Search..."
        value={searchTerm}
        onChange={(e) => setSearchTerm(e.target.value)}
      />
    </div>
  );
};

export default SearchBar;

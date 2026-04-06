export type FileItem = {
  name: string;
  isDir: boolean;
  size: number;
  starred: boolean
  last_opened: string
};

export type Usage = {
  percent: number,
  usedBytes: number,
  limitBytes: number
}

export type ViewType = "home" | "recent" | "starred" | "trash";

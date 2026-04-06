import React from "react";
import {
  Breadcrumb,
  BreadcrumbEllipsis,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "./ui/breadcrumb";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { Button } from "./ui/button";

interface BreadcrumbsProps {
  currentPath: string;
  setCurrentPath: (path: string) => void;
}

const MAX_VISIBLE = 2;

const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ currentPath, setCurrentPath }) => {
  const parts = currentPath.split("/").filter(Boolean);

  const buildPath = (index: number) =>
    "/" + parts.slice(0, index + 1).join("/");

  const isCollapsed = parts.length > MAX_VISIBLE + 1;
  const hiddenParts = isCollapsed ? parts.slice(0, parts.length - MAX_VISIBLE) : [];
  const visibleParts = isCollapsed ? parts.slice(parts.length - MAX_VISIBLE) : parts;

  return (
    <div className="px-5 py-3 border-b border-black/6 dark:border-white/6">
      <Breadcrumb>
        <BreadcrumbList>

          {/* Home */}
          <BreadcrumbItem>
            <BreadcrumbLink
              className="cursor-pointer text-lg font-bold"
              onClick={() => setCurrentPath("/")}
            >
              Home
            </BreadcrumbLink>
          </BreadcrumbItem>

          {/* Collapsed middle segments */}
          {isCollapsed && (
            <>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button size="icon-sm" variant="ghost" className="w-6 h-6">
                      <BreadcrumbEllipsis className="w-4 h-4" />
                      <span className="sr-only">Show hidden folders</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="start">
                    {hiddenParts.map((part, i) => (
                      <DropdownMenuItem
                        key={buildPath(i)}
                        onClick={() => setCurrentPath(buildPath(i))}
                      >
                        {part}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
              </BreadcrumbItem>
            </>
          )}

          {/* Visible segments */}
          {visibleParts.map((part, i) => {
            const absoluteIndex = isCollapsed
              ? parts.length - MAX_VISIBLE + i
              : i;
            const path = buildPath(absoluteIndex);
            const isLast = absoluteIndex === parts.length - 1;

            return (
              <React.Fragment key={path}>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                  {isLast ? (
                    <BreadcrumbPage
                      className="text-lg font-bold text-primary-container dark:text-white"
                    >{part}</BreadcrumbPage>
                  ) : (
                    <BreadcrumbLink
                      className="cursor-pointer text-lg"
                      onClick={() => setCurrentPath(path)}
                    >
                      {part}
                    </BreadcrumbLink>
                  )}
                </BreadcrumbItem>
              </React.Fragment>
            );
          })}

        </BreadcrumbList>
      </Breadcrumb>
    </div>
  );
};

export default Breadcrumbs;

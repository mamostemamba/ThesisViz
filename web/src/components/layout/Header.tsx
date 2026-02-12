import Link from "next/link";
import { ChevronRight } from "lucide-react";

interface HeaderProps {
  projectTitle?: string;
}

export function Header({ projectTitle }: HeaderProps) {
  return (
    <header className="border-b bg-background">
      <div className="container mx-auto flex h-14 items-center px-4">
        <Link href="/" className="text-lg font-semibold">
          ThesisViz
        </Link>
        {projectTitle && (
          <div className="ml-2 flex items-center text-sm text-muted-foreground">
            <ChevronRight className="h-4 w-4" />
            <span className="ml-1 font-medium text-foreground">
              {projectTitle}
            </span>
          </div>
        )}
        <nav className="ml-auto flex gap-4 text-sm text-muted-foreground">
          <Link href="/" className="hover:text-foreground">
            项目列表
          </Link>
        </nav>
      </div>
    </header>
  );
}

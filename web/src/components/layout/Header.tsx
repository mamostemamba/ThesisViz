import Link from "next/link";

export function Header() {
  return (
    <header className="border-b bg-background">
      <div className="container mx-auto flex h-14 items-center px-4">
        <Link href="/" className="text-lg font-semibold">
          ThesisViz
        </Link>
        <nav className="ml-8 flex gap-4 text-sm text-muted-foreground">
          <Link href="/project" className="hover:text-foreground">
            Workspace
          </Link>
        </nav>
      </div>
    </header>
  );
}

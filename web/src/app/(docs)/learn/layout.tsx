import type { ReactNode } from "react";

import { DocsHeaderBar } from "./_components/docs-header-bar";
import { DocsSearchDialog } from "./_components/docs-search-dialog";
import { DocsSidebar } from "./_components/docs-sidebar";

export default function LearnLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-screen w-full bg-background">
      {/* 1. Left sticky sub-navigation */}
      <DocsSidebar />

      {/* 2. Main content area with top bar */}
      <div className="flex min-w-0 flex-1 flex-col">
        <DocsHeaderBar />
        <main className="flex-1 overflow-x-clip">{children}</main>
      </div>

      {/* 3. Global Docs Command Palette */}
      <DocsSearchDialog />
    </div>
  );
}

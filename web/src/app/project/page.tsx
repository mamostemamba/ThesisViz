"use client";

import { useSearchParams } from "next/navigation";
import { Suspense } from "react";
import { Header } from "@/components/layout/Header";
import { Sidebar } from "@/components/layout/Sidebar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ExpertToolbox } from "@/components/workspace/ExpertToolbox";
import { SmartMode } from "@/components/workspace/SmartMode";
import { HistoryPanel } from "@/components/workspace/HistoryPanel";
import { useProject } from "@/lib/queries";

function ProjectWorkspace() {
  const searchParams = useSearchParams();
  const projectId = searchParams.get("id") || "";
  const { data: project } = useProject(projectId);

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Header projectTitle={project?.title} />
      <div className="flex flex-1">
        <Sidebar />
        <main className="flex-1 p-6">
          <Tabs defaultValue="smart" className="w-full">
            <TabsList>
              <TabsTrigger value="smart">Smart Analysis</TabsTrigger>
              <TabsTrigger value="expert">Expert Toolbox</TabsTrigger>
              <TabsTrigger value="history">History</TabsTrigger>
            </TabsList>

            <TabsContent value="smart" className="mt-6">
              <SmartMode projectId={projectId} />
            </TabsContent>

            <TabsContent value="expert" className="mt-6">
              <ExpertToolbox projectId={projectId} />
            </TabsContent>

            <TabsContent value="history" className="mt-6">
              <HistoryPanel projectId={projectId} />
            </TabsContent>
          </Tabs>
        </main>
      </div>
    </div>
  );
}

export default function ProjectPage() {
  return (
    <Suspense>
      <ProjectWorkspace />
    </Suspense>
  );
}

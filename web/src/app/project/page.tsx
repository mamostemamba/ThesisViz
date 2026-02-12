"use client";

import { useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
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
  const [activeTab, setActiveTab] = useState("smart");

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Header projectTitle={project?.title} />
      <div className="flex flex-1">
        <Sidebar />
        <main className="flex-1 p-6">
          {/* key={projectId} forces full remount when switching projects,
              resetting all local component state */}
          <Tabs key={projectId} value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList>
              <TabsTrigger value="smart">智能分析</TabsTrigger>
              <TabsTrigger value="expert">专家工具箱</TabsTrigger>
              <TabsTrigger value="history">历史记录</TabsTrigger>
            </TabsList>

            <TabsContent value="smart" className="mt-6" forceMount hidden={activeTab !== "smart"}>
              <SmartMode projectId={projectId} />
            </TabsContent>

            <TabsContent value="expert" className="mt-6" forceMount hidden={activeTab !== "expert"}>
              <ExpertToolbox projectId={projectId} />
            </TabsContent>

            <TabsContent value="history" className="mt-6" forceMount hidden={activeTab !== "history"}>
              <HistoryPanel projectId={projectId} onLoadResult={() => setActiveTab("smart")} />
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

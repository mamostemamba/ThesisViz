"use client";

import { Header } from "@/components/layout/Header";
import { Sidebar } from "@/components/layout/Sidebar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { ExpertToolbox } from "@/components/workspace/ExpertToolbox";
import { SmartMode } from "@/components/workspace/SmartMode";

export default function ProjectPage() {
  return (
    <div className="flex min-h-screen flex-col bg-background">
      <Header />
      <div className="flex flex-1">
        <Sidebar />
        <main className="flex-1 p-6">
          <Tabs defaultValue="smart" className="w-full">
            <TabsList>
              <TabsTrigger value="smart">Smart Analysis</TabsTrigger>
              <TabsTrigger value="expert">Expert Toolbox</TabsTrigger>
            </TabsList>

            <TabsContent value="smart" className="mt-6">
              <SmartMode />
            </TabsContent>

            <TabsContent value="expert" className="mt-6">
              <ExpertToolbox />
            </TabsContent>
          </Tabs>
        </main>
      </div>
    </div>
  );
}

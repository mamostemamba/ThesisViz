"use client";

import { Header } from "@/components/layout/Header";
import { Sidebar } from "@/components/layout/Sidebar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ExpertToolbox } from "@/components/workspace/ExpertToolbox";

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
              <div className="grid gap-6 lg:grid-cols-2">
                <Card>
                  <CardHeader>
                    <CardTitle>Describe Your Figure</CardTitle>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <Textarea
                      placeholder="Describe the figure you want to create... e.g., 'A bar chart comparing model accuracy across 5 datasets'"
                      className="min-h-[200px]"
                    />
                    <Button className="w-full">Generate</Button>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader>
                    <CardTitle>Preview</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className="flex h-[300px] items-center justify-center rounded-md border border-dashed text-muted-foreground">
                      Generated figure will appear here
                    </div>
                  </CardContent>
                </Card>
              </div>
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

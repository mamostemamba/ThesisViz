"use client";

import { Header } from "@/components/layout/Header";
import { ApiKeySetup } from "@/components/projects/ApiKeySetup";
import { ProjectGallery } from "@/components/projects/ProjectGallery";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="container mx-auto px-4 py-8">
        <ApiKeySetup />
        <ProjectGallery />
      </main>
    </div>
  );
}

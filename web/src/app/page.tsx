import Link from "next/link";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Header } from "@/components/layout/Header";

export default function Home() {
  return (
    <div className="min-h-screen bg-background">
      <Header />
      <main className="container mx-auto px-4 py-16">
        <div className="mx-auto max-w-2xl text-center">
          <h1 className="text-4xl font-bold tracking-tight">ThesisViz</h1>
          <p className="mt-4 text-lg text-muted-foreground">
            AI-powered academic figure generation. Describe your figure in
            natural language and get publication-ready TikZ, Matplotlib, or
            Mermaid output.
          </p>
          <div className="mt-8">
            <Link href="/project">
              <Button size="lg">Open Workspace</Button>
            </Link>
          </div>
        </div>

        <div className="mx-auto mt-16 grid max-w-3xl gap-6 md:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle>TikZ</CardTitle>
              <CardDescription>
                Vector graphics for architecture diagrams and mathematical
                illustrations
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Matplotlib</CardTitle>
              <CardDescription>
                Data visualization charts and plots for experimental results
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>Mermaid</CardTitle>
              <CardDescription>
                Flowcharts, sequence diagrams, and other structured diagrams
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </main>
    </div>
  );
}

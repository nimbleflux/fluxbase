import { cn } from "@/lib/utils";
import { ThemeProvider } from "@/context/theme-provider";

interface MainProps {
  fixed?: boolean;
  fluid?: boolean;
  className?: string;
  children?: React.ReactNode;
}

export function Main({ fixed, className, fluid, children }: MainProps) {
  return (
    <ThemeProvider>
      <main
        className={cn(
          "flex-1",
          fixed && "fixed inset-0 overflow-auto",
          fluid ? "w-full" : "container mx-auto px-4",
          className,
        )}
      >
        {children}
      </main>
    </ThemeProvider>
  );
}

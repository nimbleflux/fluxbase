import { useQuery } from "@tanstack/react-query";
import api, { type KnowledgeBaseSummary } from "@/lib/api";
import { KnowledgeBaseCard } from "./-KnowledgeBaseCard";

function PublicKnowledgeBasesList() {
  const { data, isLoading } = useQuery({
    queryKey: ["my-knowledge-bases"],
    queryFn: async () => {
      const res = await api.get("/api/v1/ai/knowledge-bases");
      return res.data;
    },
  });

  if (isLoading) return <div className="text-center py-8">Loading...</div>;

  const publicKBs =
    data?.knowledge_bases?.filter(
      (kb: KnowledgeBaseSummary) => kb.visibility === "public",
    ) || [];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {publicKBs.map((kb: KnowledgeBaseSummary) => (
        <KnowledgeBaseCard key={kb.id} kb={kb} isOwner={false} />
      ))}
      {publicKBs.length === 0 && (
        <div className="col-span-full text-center py-12 text-muted-foreground">
          No public knowledge bases available.
        </div>
      )}
    </div>
  );
}

export { PublicKnowledgeBasesList };

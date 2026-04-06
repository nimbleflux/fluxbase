import { useQuery } from "@tanstack/react-query";
import api, { type KnowledgeBaseSummary } from "@/lib/api";
import { KnowledgeBaseCard } from "./-KnowledgeBaseCard";

function MyKnowledgeBasesList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ["my-knowledge-bases"],
    queryFn: async () => {
      const res = await api.get("/api/v1/ai/knowledge-bases");
      return res.data;
    },
  });

  if (isLoading) return <div className="text-center py-8">Loading...</div>;
  if (error)
    return (
      <div className="text-red-500 py-8">Error loading knowledge bases</div>
    );

  const myKBs =
    data?.knowledge_bases?.filter(
      (kb: KnowledgeBaseSummary) => kb.user_permission === "owner",
    ) || [];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {myKBs.map((kb: KnowledgeBaseSummary) => (
        <KnowledgeBaseCard key={kb.id} kb={kb} isOwner={true} />
      ))}
      {myKBs.length === 0 && (
        <div className="col-span-full text-center py-12 text-muted-foreground">
          No knowledge bases yet. Create your first one!
        </div>
      )}
    </div>
  );
}

export { MyKnowledgeBasesList };

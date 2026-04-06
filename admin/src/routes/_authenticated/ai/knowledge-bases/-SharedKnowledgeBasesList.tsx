import { useQuery } from "@tanstack/react-query";
import api, { type KnowledgeBaseSummary } from "@/lib/api";
import { KnowledgeBaseCard } from "./-KnowledgeBaseCard";

function SharedKnowledgeBasesList() {
  const { data, isLoading } = useQuery({
    queryKey: ["my-knowledge-bases"],
    queryFn: async () => {
      const res = await api.get("/api/v1/ai/knowledge-bases");
      return res.data;
    },
  });

  if (isLoading) return <div className="text-center py-8">Loading...</div>;

  const sharedKBs =
    data?.knowledge_bases?.filter(
      (kb: KnowledgeBaseSummary) =>
        kb.visibility === "shared" &&
        kb.user_permission &&
        kb.user_permission !== "owner",
    ) || [];

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {sharedKBs.map((kb: KnowledgeBaseSummary) => (
        <KnowledgeBaseCard key={kb.id} kb={kb} isOwner={false} />
      ))}
      {sharedKBs.length === 0 && (
        <div className="col-span-full text-center py-12 text-muted-foreground">
          No knowledge bases have been shared with you yet.
        </div>
      )}
    </div>
  );
}

export { SharedKnowledgeBasesList };

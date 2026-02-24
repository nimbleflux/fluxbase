import { useState, useEffect, useCallback, useMemo } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import {
  ReactFlow,
  type Node,
  type Edge,
  Controls,
  MiniMap,
  Background,
  useNodesState,
  useEdgesState,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import {
  RefreshCw,
  Search,
  Filter,
  Bot,
  User,
  Building2,
  MapPin,
  Lightbulb,
  Package,
  Calendar,
  Table2,
  Link2,
  Code,
  AlertCircle,
  HelpCircle,
  GitBranch,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  knowledgeBasesApi,
  type Entity,
  type KnowledgeGraphData,
  type EntityType,
  type ChatbotKnowledgeBaseLink,
  type KnowledgeBase,
} from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { KnowledgeBaseHeader } from '@/components/knowledge-bases/knowledge-base-header'

// Entity type colors
const ENTITY_COLORS: Record<EntityType, string> = {
  person: '#3b82f6',
  organization: '#8b5cf6',
  location: '#10b981',
  concept: '#f59e0b',
  product: '#ef4444',
  event: '#ec4899',
  table: '#6366f1',
  url: '#14b8a6',
  api_endpoint: '#f97316',
  datetime: '#06b6d4',
  code_reference: '#84cc16',
  error: '#dc2626',
  other: '#9ca3af',
}

// Entity type icons
const ENTITY_ICONS: Record<EntityType, React.ElementType> = {
  person: User,
  organization: Building2,
  location: MapPin,
  concept: Lightbulb,
  product: Package,
  event: Calendar,
  table: Table2,
  url: Link2,
  api_endpoint: Code,
  datetime: Calendar,
  code_reference: Code,
  error: AlertCircle,
  other: HelpCircle,
}

// Custom node component
function EntityNode({ data }: { data: { entity: Entity } }) {
  const entity = data.entity
  const color = ENTITY_COLORS[entity.entity_type] || ENTITY_COLORS.other
  const Icon = ENTITY_ICONS[entity.entity_type] || HelpCircle

  return (
    <div
      className='rounded-lg border-2 bg-white px-3 py-2 shadow-sm dark:bg-gray-800'
      style={{ borderColor: color }}
    >
      <div
        className='flex items-center gap-1.5 text-xs capitalize'
        style={{ color }}
      >
        <Icon className='h-3 w-3' />
        {entity.entity_type}
      </div>
      <div className='mt-0.5 max-w-[150px] truncate text-sm font-medium'>
        {entity.name}
      </div>
      {entity.document_count !== undefined && entity.document_count > 0 && (
        <div className='mt-1 text-xs text-gray-500'>
          {entity.document_count} docs
        </div>
      )}
    </div>
  )
}

const nodeTypes = {
  entity: EntityNode,
}

export const Route = createFileRoute(
  '/_authenticated/knowledge-bases/$id/graph'
)({
  component: KnowledgeGraphPage,
})

function KnowledgeGraphPage() {
  const params = Route.useParams()
  const id = params.id
  const [knowledgeBase, setKnowledgeBase] = useState<KnowledgeBase | null>(null)
  const [graphData, setGraphData] = useState<KnowledgeGraphData | null>(null)
  const [linkedChatbots, setLinkedChatbots] = useState<ChatbotKnowledgeBaseLink[]>(
    []
  )
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<string>('all')
  const [selectedEntity, setSelectedEntity] = useState<Entity | null>(null)

  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [kb, graph, chatbots] = await Promise.all([
        knowledgeBasesApi.get(id),
        knowledgeBasesApi.getKnowledgeGraph(id),
        knowledgeBasesApi.listLinkedChatbots(id),
      ])
      setKnowledgeBase(kb)
      setGraphData(graph)
      setLinkedChatbots(chatbots || [])
    } catch {
      toast.error('Failed to fetch knowledge graph')
    } finally {
      setLoading(false)
    }
  }, [id])

  // Convert graph data to ReactFlow nodes and edges
  useEffect(() => {
    if (!graphData) return

    const entities = graphData.entities || []
    const relationships = graphData.relationships || []

    // Filter entities based on search and type
    const filteredEntities = entities.filter((entity) => {
      const matchesSearch =
        !searchQuery ||
        entity.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        entity.canonical_name.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesType = typeFilter === 'all' || entity.entity_type === typeFilter
      return matchesSearch && matchesType
    })

    const entityIds = new Set(filteredEntities.map((e) => e.id))

    // Filter relationships to only include filtered entities
    const filteredRelationships = relationships.filter(
      (rel) => entityIds.has(rel.source_entity_id) && entityIds.has(rel.target_entity_id)
    )

    // Create nodes with simple layout
    const nodePositions: Record<string, { x: number; y: number }> = {}
    const typeGroups: Record<string, Entity[]> = {}

    filteredEntities.forEach((entity) => {
      const entityType = entity.entity_type
      if (!typeGroups[entityType]) typeGroups[entityType] = []
      typeGroups[entityType].push(entity)
    })

    let yOffset = 0
    Object.entries(typeGroups).forEach(([, entities]) => {
      entities.forEach((entity, idx) => {
        nodePositions[entity.id] = {
          x: idx * 200,
          y: yOffset,
        }
      })
      yOffset += 150
    })

    const flowNodes: Node[] = filteredEntities.map((entity) => ({
      id: entity.id,
      type: 'entity',
      position: nodePositions[entity.id] || { x: 0, y: 0 },
      data: { entity },
    }))

    const flowEdges: Edge[] = filteredRelationships.map((rel) => ({
      id: rel.id,
      source: rel.source_entity_id,
      target: rel.target_entity_id,
      label: rel.relationship_type,
      animated: false,
      style: { stroke: '#9ca3af' },
      labelStyle: { fontSize: 10, fill: '#6b7280' },
      labelBgStyle: { fill: 'hsl(var(--card))', fillOpacity: 0.95 },
      labelBgPadding: [4, 2] as [number, number],
      labelBgBorderRadius: 4,
    }))

    setNodes(flowNodes)
    setEdges(flowEdges)
  }, [graphData, searchQuery, typeFilter, setNodes, setEdges])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const uniqueEntityTypes = useMemo(() => {
    if (!graphData) return []
    const entities = graphData.entities || []
    const types = new Set(entities.map((e) => e.entity_type))
    return Array.from(types).sort()
  }, [graphData])

  if (loading) {
    return (
      <div className='flex h-96 items-center justify-center'>
        <RefreshCw className='text-muted-foreground h-8 w-8 animate-spin' />
      </div>
    )
  }

  if (!knowledgeBase) {
    return (
      <div className='flex h-96 flex-col items-center justify-center gap-4'>
        <p className='text-muted-foreground'>Knowledge base not found</p>
      </div>
    )
  }

  return (
    <div className='flex flex-1 flex-col gap-6 p-6'>
      <KnowledgeBaseHeader
        knowledgeBase={knowledgeBase}
        activeTab='graph'
        actions={
          <Button
            variant='outline'
            size='sm'
            onClick={() => {
              setSearchQuery('')
              setTypeFilter('all')
              setSelectedEntity(null)
            }}
          >
            <RefreshCw className='mr-2 h-4 w-4' />
            Reset View
          </Button>
        }
      />

      <div className='grid flex-1 gap-6 lg:grid-cols-[280px_1fr]'>
        {/* Left Sidebar */}
        <div className='space-y-4'>
          {/* Stats Card */}
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm'>Graph Statistics</CardTitle>
            </CardHeader>
            <CardContent className='space-y-2'>
              <div className='flex justify-between text-sm'>
                <span className='text-muted-foreground'>Entities</span>
                <Badge variant='secondary'>{graphData?.entity_count || 0}</Badge>
              </div>
              <div className='flex justify-between text-sm'>
                <span className='text-muted-foreground'>Relationships</span>
                <Badge variant='secondary'>
                  {graphData?.relationship_count || 0}
                </Badge>
              </div>
            </CardContent>
          </Card>

          {/* Filters */}
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-sm'>
                <Filter className='h-4 w-4' />
                Filters
              </CardTitle>
            </CardHeader>
            <CardContent className='space-y-3'>
              <div className='relative'>
                <Search className='text-muted-foreground absolute left-2.5 top-2.5 h-4 w-4' />
                <Input
                  placeholder='Search entities...'
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className='pl-8'
                />
              </div>
              <Select value={typeFilter} onValueChange={setTypeFilter}>
                <SelectTrigger>
                  <SelectValue placeholder='All types' />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='all'>All types</SelectItem>
                  {uniqueEntityTypes.map((entityType) => (
                    <SelectItem key={entityType} value={entityType}>
                      <div className='flex items-center gap-2'>
                        <div
                          className='h-2 w-2 rounded-full'
                          style={{ backgroundColor: ENTITY_COLORS[entityType] }}
                        />
                        <span className='capitalize'>{entityType}</span>
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </CardContent>
          </Card>

          {/* Entity Types Legend */}
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='text-sm'>Entity Types</CardTitle>
            </CardHeader>
            <CardContent>
              <ScrollArea className='h-[200px]'>
                <div className='space-y-1.5'>
                  {uniqueEntityTypes.map((entityType) => {
                    const count = (graphData?.entities || []).filter(
                      (e) => e.entity_type === entityType
                    ).length
                    return (
                      <div
                        key={entityType}
                        className='flex items-center justify-between rounded-md px-2 py-1.5 hover:bg-muted'
                      >
                        <div className='flex items-center gap-2'>
                          <div
                            className='h-3 w-3 rounded-full'
                            style={{ backgroundColor: ENTITY_COLORS[entityType] }}
                          />
                          <span className='capitalize text-sm'>{entityType}</span>
                        </div>
                        <Badge variant='outline' className='text-xs'>
                          {count}
                        </Badge>
                      </div>
                    )
                  })}
                </div>
              </ScrollArea>
            </CardContent>
          </Card>

          {/* Linked Chatbots */}
          <Card>
            <CardHeader className='pb-2'>
              <CardTitle className='flex items-center gap-2 text-sm'>
                <Bot className='h-4 w-4' />
                Linked Chatbots
              </CardTitle>
            </CardHeader>
            <CardContent>
              {linkedChatbots.length === 0 ? (
                <p className='text-muted-foreground text-sm'>
                  No chatbots using this knowledge base
                </p>
              ) : (
                <ScrollArea className='h-[150px]'>
                  <div className='space-y-2'>
                    {linkedChatbots.map((link) => (
                      <div
                        key={link.id}
                        className='flex items-center justify-between rounded-md border p-2'
                      >
                        <span className='text-sm font-medium'>
                          {link.chatbot_name || link.chatbot_id}
                        </span>
                        <Badge variant='outline' className='text-xs'>
                          P{link.priority}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </ScrollArea>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Main Graph Area */}
        <Card className='overflow-hidden'>
          <div className='h-[600px]'>
            {graphData && (graphData.entities || []).length > 0 ? (
              <ReactFlow
                nodes={nodes}
                edges={edges}
                onNodesChange={onNodesChange}
                onEdgesChange={onEdgesChange}
                nodeTypes={nodeTypes}
                fitView
                minZoom={0.1}
                maxZoom={2}
                onNodeClick={(_, node) => {
                  const entity = (node.data as { entity: Entity })?.entity
                  if (entity) setSelectedEntity(entity)
                }}
                className='react-flow-dark'
              >
                <Controls />
                <MiniMap
                  nodeColor={(node) => {
                    const entity = node.data?.entity as Entity | undefined
                    return entity
                      ? ENTITY_COLORS[entity.entity_type]
                      : '#9ca3af'
                  }}
                  maskColor='rgba(100,100,100,0.1)'
                />
                <Background color='#888' gap={16} />
              </ReactFlow>
            ) : (
              <div className='flex h-full flex-col items-center justify-center gap-4'>
                <GitBranch className='text-muted-foreground h-16 w-16' />
                <p className='text-lg font-medium'>No entities found</p>
                <p className='text-muted-foreground text-sm'>
                  Add documents to the knowledge base to extract entities
                </p>
              </div>
            )}
          </div>
        </Card>
      </div>

      {/* Selected Entity Details */}
      {selectedEntity && (
        <Card className='fixed bottom-4 right-4 w-[350px] shadow-lg'>
          <CardHeader className='pb-2'>
            <div className='flex items-center justify-between'>
              <CardTitle className='text-base'>{selectedEntity.name}</CardTitle>
              <Button
                variant='ghost'
                size='sm'
                className='h-6 w-6 p-0'
                onClick={() => setSelectedEntity(null)}
              >
                &times;
              </Button>
            </div>
            <Badge
              variant='outline'
              style={{
                borderColor: ENTITY_COLORS[selectedEntity.entity_type],
                color: ENTITY_COLORS[selectedEntity.entity_type],
              }}
            >
              {selectedEntity.entity_type}
            </Badge>
          </CardHeader>
          <CardContent className='space-y-2 text-sm'>
            {selectedEntity.canonical_name !== selectedEntity.name && (
              <div>
                <span className='text-muted-foreground'>Canonical:</span>{' '}
                {selectedEntity.canonical_name}
              </div>
            )}
            {selectedEntity.aliases && selectedEntity.aliases.length > 0 && (
              <div>
                <span className='text-muted-foreground'>Aliases:</span>{' '}
                {selectedEntity.aliases.join(', ')}
              </div>
            )}
            {selectedEntity.document_count !== undefined && (
              <div>
                <span className='text-muted-foreground'>Documents:</span>{' '}
                {selectedEntity.document_count}
              </div>
            )}
            {Object.keys(selectedEntity.metadata || {}).length > 0 && (
              <div className='space-y-1'>
                <span className='text-muted-foreground'>Metadata:</span>
                <pre className='bg-muted overflow-auto rounded p-2 text-xs'>
                  {JSON.stringify(selectedEntity.metadata, null, 2)}
                </pre>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}

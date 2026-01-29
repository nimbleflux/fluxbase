import { useState, useCallback, useMemo, useRef, useEffect } from 'react'
import { z } from 'zod'
import { useQuery } from '@tanstack/react-query'
import { createFileRoute, getRouteApi, Link } from '@tanstack/react-router'
import {
  GitFork,
  Loader2,
  AlertCircle,
  AlertTriangle,
  Search,
  Database,
  Key,
  Link as LinkIcon,
  Shield,
  ShieldOff,
  ArrowRight,
  ArrowLeft,
  Columns,
  LayoutGrid,
  List,
  ZoomIn,
  ZoomOut,
  Maximize2,
  Hash,
  Fingerprint,
} from 'lucide-react'
import {
  databaseApi,
  schemaApi,
  policyApi,
  type SchemaNode,
  type SchemaRelationship,
  type SecurityWarning,
} from '@/lib/api'
import { cn } from '@/lib/utils'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'

const schemaSearchSchema = z.object({
  schema: z.string().optional(),
})

export const Route = createFileRoute('/_authenticated/schema/')({
  validateSearch: schemaSearchSchema,
  component: SchemaViewerPage,
})

const routeApi = getRouteApi('/_authenticated/schema/')

type ViewMode = 'erd' | 'list'

function SchemaViewerPage() {
  const search = routeApi.useSearch()
  const navigate = routeApi.useNavigate()

  const [viewMode, setViewMode] = useState<ViewMode>('erd')
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [zoom, setZoom] = useState(1)

  // Fetch available schemas first
  const { data: availableSchemas = ['public'], isLoading: schemasLoading } =
    useQuery({
      queryKey: ['available-schemas'],
      queryFn: databaseApi.getSchemas,
      staleTime: 5 * 60 * 1000, // Cache for 5 minutes
    })

  // Use schema from URL or default to 'public'
  const selectedSchema = search.schema || 'public'

  // Fetch graph for selected schema
  const {
    data,
    isLoading: graphLoading,
    error,
  } = useQuery({
    queryKey: ['schema-graph', selectedSchema],
    queryFn: () => schemaApi.getGraph([selectedSchema]),
    enabled: !schemasLoading,
  })

  // Fetch security warnings for warning counts
  const { data: warningsData } = useQuery({
    queryKey: ['security-warnings'],
    queryFn: () => policyApi.getSecurityWarnings(),
    staleTime: 60000, // Cache for 1 minute
  })

  // Group warnings by table
  const warningsByTable = useMemo(() => {
    if (!warningsData?.warnings) return new Map<string, SecurityWarning[]>()

    const grouped = new Map<string, SecurityWarning[]>()
    for (const warning of warningsData.warnings) {
      const key = `${warning.schema}.${warning.table}`
      const existing = grouped.get(key) || []
      existing.push(warning)
      grouped.set(key, existing)
    }
    return grouped
  }, [warningsData])

  // Helper to get warning count for a table
  const getTableWarningCount = useCallback(
    (schema: string, table: string): number => {
      return warningsByTable.get(`${schema}.${table}`)?.length || 0
    },
    [warningsByTable]
  )

  // Helper to get highest severity for a table
  const getTableWarningSeverity = useCallback(
    (
      schema: string,
      table: string
    ): 'critical' | 'high' | 'medium' | 'low' | null => {
      const warnings = warningsByTable.get(`${schema}.${table}`)
      if (!warnings?.length) return null

      const severityOrder = ['critical', 'high', 'medium', 'low'] as const
      for (const severity of severityOrder) {
        if (warnings.some((w) => w.severity === severity)) return severity
      }
      return 'low'
    },
    [warningsByTable]
  )

  // Helper to get warnings for a table
  const getTableWarnings = useCallback(
    (schema: string, table: string): SecurityWarning[] => {
      return warningsByTable.get(`${schema}.${table}`) || []
    },
    [warningsByTable]
  )

  const isLoading = schemasLoading || graphLoading

  // Helper to get full name for a node
  const getFullName = (node: SchemaNode) => `${node.schema}.${node.name}`

  // Filter nodes based on search
  const filteredNodes = useMemo(() => {
    if (!data?.nodes) return []
    if (!searchQuery) return data.nodes
    const query = searchQuery.toLowerCase()
    return data.nodes.filter(
      (n) =>
        n.name.toLowerCase().includes(query) ||
        getFullName(n).toLowerCase().includes(query) ||
        n.columns.some((c) => c.name.toLowerCase().includes(query))
    )
  }, [data, searchQuery])

  // Get relationships for filtered nodes (use edges from API)
  // Only include relationships where BOTH tables are visible
  const filteredRelationships = useMemo(() => {
    if (!data?.edges || !filteredNodes.length) return []
    const nodeNames = new Set(filteredNodes.map((n) => getFullName(n)))
    return data.edges.filter(
      (r) =>
        nodeNames.has(`${r.source_schema}.${r.source_table}`) &&
        nodeNames.has(`${r.target_schema}.${r.target_table}`)
    )
  }, [data, filteredNodes])

  // Get selected table details
  const selectedTableData = useMemo(() => {
    if (!selectedTable || !data?.nodes) return null
    const node = data.nodes.find((n) => getFullName(n) === selectedTable)
    if (!node) return null

    // Deduplicate columns by name (safety net for backend issues)
    const seen = new Set<string>()
    const uniqueColumns = node.columns.filter((col) => {
      if (seen.has(col.name)) return false
      seen.add(col.name)
      return true
    })

    return { ...node, columns: uniqueColumns }
  }, [selectedTable, data])

  // Get relationships for selected table
  const selectedTableRelationships = useMemo(() => {
    if (!selectedTable || !data?.edges) return { incoming: [], outgoing: [] }
    const [schema, table] = selectedTable.split('.')
    return {
      incoming: data.edges.filter(
        (r) => r.target_schema === schema && r.target_table === table
      ),
      outgoing: data.edges.filter(
        (r) => r.source_schema === schema && r.source_table === table
      ),
    }
  }, [selectedTable, data])

  const handleSchemaChange = (schema: string) => {
    navigate({ search: { schema } })
  }

  if (error) {
    return (
      <div className='flex flex-1 flex-col gap-6 p-6'>
        <div className='text-destructive flex items-center gap-2'>
          <AlertCircle className='h-5 w-5' />
          <span>Failed to load schema graph</span>
        </div>
      </div>
    )
  }

  return (
    <div className='flex flex-1 flex-col gap-6 p-6'>
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='flex items-center gap-2 text-3xl font-bold tracking-tight'>
            <GitFork className='h-8 w-8' />
            Schema Viewer
          </h1>
          <p className='text-muted-foreground mt-2 text-sm'>
            Visualize database tables and their relationships
          </p>
        </div>
        <div className='flex items-center gap-2'>
          <Button
            variant={viewMode === 'erd' ? 'default' : 'outline'}
            size='sm'
            onClick={() => setViewMode('erd')}
          >
            <LayoutGrid className='mr-2 h-4 w-4' />
            ERD View
          </Button>
          <Button
            variant={viewMode === 'list' ? 'default' : 'outline'}
            size='sm'
            onClick={() => setViewMode('list')}
          >
            <List className='mr-2 h-4 w-4' />
            List View
          </Button>
        </div>
      </div>

      {/* Filters */}
      <div className='flex items-center gap-4'>
        <div className='relative max-w-sm flex-1'>
          <Search className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
          <Input
            placeholder='Search tables, columns...'
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className='pl-9'
          />
        </div>
        <Select
          value={selectedSchema}
          onValueChange={handleSchemaChange}
          disabled={schemasLoading}
        >
          <SelectTrigger className='w-[180px]'>
            <SelectValue
              placeholder={schemasLoading ? 'Loading...' : 'Select schema'}
            />
          </SelectTrigger>
          <SelectContent>
            {availableSchemas.map((schema) => (
              <SelectItem key={schema} value={schema}>
                {schema}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {viewMode === 'erd' && (
          <div className='flex items-center gap-1'>
            <Button
              variant='outline'
              size='icon'
              onClick={() => setZoom((z) => Math.max(0.25, z - 0.25))}
            >
              <ZoomOut className='h-4 w-4' />
            </Button>
            <span className='text-muted-foreground w-12 text-center text-sm'>
              {Math.round(zoom * 100)}%
            </span>
            <Button
              variant='outline'
              size='icon'
              onClick={() => setZoom((z) => Math.min(2, z + 0.25))}
            >
              <ZoomIn className='h-4 w-4' />
            </Button>
            <Button variant='outline' size='icon' onClick={() => setZoom(1)}>
              <Maximize2 className='h-4 w-4' />
            </Button>
          </div>
        )}
      </div>

      {isLoading ? (
        <div className='flex justify-center py-12'>
          <Loader2 className='text-muted-foreground h-8 w-8 animate-spin' />
        </div>
      ) : viewMode === 'erd' ? (
        <div className='flex flex-1 gap-6'>
          {/* ERD Canvas */}
          <div className='bg-muted/20 flex-1 overflow-auto rounded-lg border' style={{ minWidth: 0 }}>
            <ERDCanvas
              nodes={filteredNodes}
              relationships={filteredRelationships}
              zoom={zoom}
              onZoomChange={setZoom}
              selectedTable={selectedTable}
              onSelectTable={setSelectedTable}
              getTableWarningCount={getTableWarningCount}
              getTableWarningSeverity={getTableWarningSeverity}
              getTableWarnings={getTableWarnings}
            />
          </div>

          {/* Table Details Panel */}
          {selectedTable && selectedTableData && (
            <Card className='w-[420px] shrink-0'>
              <CardHeader className='pb-3'>
                <div className='flex items-center justify-between'>
                  <CardTitle className='flex items-center gap-2 text-lg'>
                    <Database className='h-4 w-4' />
                    {selectedTableData.name}
                  </CardTitle>
                  <div className='flex items-center gap-2'>
                    {(() => {
                      const count = getTableWarningCount(
                        selectedTableData.schema,
                        selectedTableData.name
                      )
                      const severity = getTableWarningSeverity(
                        selectedTableData.schema,
                        selectedTableData.name
                      )
                      const warnings = getTableWarnings(
                        selectedTableData.schema,
                        selectedTableData.name
                      )
                      if (count > 0) {
                        return (
                          <Popover>
                            <PopoverTrigger asChild>
                              <Badge
                                variant={
                                  severity === 'critical' || severity === 'high'
                                    ? 'destructive'
                                    : 'secondary'
                                }
                                className='cursor-pointer gap-1'
                              >
                                <AlertTriangle className='h-3 w-3' />
                                {count}
                              </Badge>
                            </PopoverTrigger>
                            <PopoverContent className='w-80' align='end'>
                              <div className='space-y-3'>
                                <h4 className='text-sm font-medium'>
                                  Security Warnings
                                </h4>
                                <div className='max-h-60 space-y-2 overflow-auto'>
                                  {warnings.map((w) => (
                                    <div
                                      key={w.id}
                                      className='border-l-2 border-l-orange-500 pl-2 text-sm'
                                    >
                                      <div className='flex items-center gap-1'>
                                        <Badge
                                          variant='outline'
                                          className='text-xs'
                                        >
                                          {w.severity}
                                        </Badge>
                                      </div>
                                      <p className='text-muted-foreground mt-1'>
                                        {w.message}
                                      </p>
                                      {w.suggestion && (
                                        <p className='text-muted-foreground mt-1 text-xs italic'>
                                          {w.suggestion}
                                        </p>
                                      )}
                                    </div>
                                  ))}
                                </div>
                                <Button
                                  variant='outline'
                                  size='sm'
                                  className='w-full'
                                  asChild
                                >
                                  <Link to='/policies'>Manage Policies</Link>
                                </Button>
                              </div>
                            </PopoverContent>
                          </Popover>
                        )
                      }
                      return null
                    })()}
                    {selectedTableData.rls_enabled ? (
                      <Badge variant='default' className='gap-1'>
                        <Shield className='h-3 w-3' />
                        RLS
                      </Badge>
                    ) : (
                      <Badge variant='secondary' className='gap-1'>
                        <ShieldOff className='h-3 w-3' />
                        No RLS
                      </Badge>
                    )}
                  </div>
                </div>
                <CardDescription>
                  {getFullName(selectedTableData)}
                </CardDescription>
              </CardHeader>
              <CardContent className='space-y-4'>
                {/* Columns */}
                <div>
                  <h4 className='mb-2 flex items-center gap-2 text-sm font-medium'>
                    <Columns className='h-4 w-4' />
                    Columns ({selectedTableData.columns.length})
                  </h4>
                  <div className='max-h-96 space-y-1 overflow-auto'>
                    {selectedTableData.columns.map((col) => (
                      <TooltipProvider key={col.name}>
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <div className='hover:bg-muted flex cursor-default items-start gap-1.5 rounded px-2 py-1 text-sm'>
                              <div className='mt-0.5 flex shrink-0 items-center gap-1'>
                                {col.is_primary_key && (
                                  <Key className='h-3 w-3 text-yellow-500' />
                                )}
                                {col.is_foreign_key && (
                                  <LinkIcon className='h-3 w-3 text-blue-500' />
                                )}
                                {col.is_unique && !col.is_primary_key && (
                                  <Fingerprint className='h-3 w-3 text-purple-500' />
                                )}
                                {col.is_indexed &&
                                  !col.is_primary_key &&
                                  !col.is_unique && (
                                    <Hash className='h-3 w-3 text-gray-500' />
                                  )}
                              </div>
                              <div className='min-w-0 flex-1'>
                                <div className='flex items-center justify-between gap-2'>
                                  <span
                                    className={cn(
                                      'truncate',
                                      col.is_primary_key && 'font-medium'
                                    )}
                                  >
                                    {col.name}
                                    {!col.nullable && (
                                      <span className='ml-0.5 text-xs text-red-500'>
                                        *
                                      </span>
                                    )}
                                  </span>
                                  <span className='text-muted-foreground shrink-0 text-xs'>
                                    {col.data_type}
                                  </span>
                                </div>
                                {col.comment && (
                                  <p className='text-muted-foreground truncate text-xs'>
                                    {col.comment}
                                  </p>
                                )}
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side='left' className='max-w-xs'>
                            <div className='space-y-1 text-xs'>
                              <div className='font-medium'>{col.name}</div>
                              <div>Type: {col.data_type}</div>
                              <div>Nullable: {col.nullable ? 'Yes' : 'No'}</div>
                              {col.default_value && (
                                <div>Default: {col.default_value}</div>
                              )}
                              {col.is_primary_key && (
                                <div className='text-yellow-500'>
                                  Primary Key
                                </div>
                              )}
                              {col.is_foreign_key && col.fk_target && (
                                <div className='text-blue-500'>
                                  FK → {col.fk_target}
                                </div>
                              )}
                              {col.is_unique && (
                                <div className='text-purple-500'>Unique</div>
                              )}
                              {col.is_indexed && (
                                <div className='text-gray-500'>Indexed</div>
                              )}
                              {col.comment && (
                                <div className='mt-1 italic'>{col.comment}</div>
                              )}
                            </div>
                          </TooltipContent>
                        </Tooltip>
                      </TooltipProvider>
                    ))}
                  </div>
                </div>

                {/* Relationships */}
                {(selectedTableRelationships.incoming.length > 0 ||
                  selectedTableRelationships.outgoing.length > 0) && (
                  <div>
                    <h4 className='mb-2 flex items-center gap-2 text-sm font-medium'>
                      <GitFork className='h-4 w-4' />
                      Relationships (
                      {selectedTableRelationships.outgoing.length +
                        selectedTableRelationships.incoming.length}
                      )
                    </h4>
                    <div className='max-h-48 space-y-2 overflow-auto'>
                      {selectedTableRelationships.outgoing.map((rel) => (
                        <TooltipProvider key={rel.id}>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <div
                                className='bg-muted/50 hover:bg-muted flex cursor-pointer items-center gap-2 rounded p-2 text-sm'
                                onClick={() =>
                                  setSelectedTable(
                                    `${rel.target_schema}.${rel.target_table}`
                                  )
                                }
                              >
                                <ArrowRight className='h-4 w-4 shrink-0 text-blue-500' />
                                <span className='text-muted-foreground truncate'>
                                  {rel.source_column}
                                </span>
                                <Badge
                                  variant='outline'
                                  className='shrink-0 px-1 text-[10px]'
                                >
                                  {rel.cardinality === 'one-to-one'
                                    ? '1:1'
                                    : 'N:1'}
                                </Badge>
                                <span className='truncate font-medium'>
                                  {rel.target_table}
                                </span>
                              </div>
                            </TooltipTrigger>
                            <TooltipContent side='left' className='max-w-xs'>
                              <div className='space-y-1 text-xs'>
                                <div className='font-medium'>Outgoing FK</div>
                                <div>
                                  {rel.source_column} → {rel.target_schema}.
                                  {rel.target_table}.{rel.target_column}
                                </div>
                                <div>Cardinality: {rel.cardinality}</div>
                                <div>ON DELETE: {rel.on_delete}</div>
                                <div>ON UPDATE: {rel.on_update}</div>
                                <div className='text-muted-foreground'>
                                  {rel.constraint_name}
                                </div>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      ))}
                      {selectedTableRelationships.incoming.map((rel) => (
                        <TooltipProvider key={rel.id}>
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <div
                                className='bg-muted/50 hover:bg-muted flex cursor-pointer items-center gap-2 rounded p-2 text-sm'
                                onClick={() =>
                                  setSelectedTable(
                                    `${rel.source_schema}.${rel.source_table}`
                                  )
                                }
                              >
                                <ArrowLeft className='h-4 w-4 shrink-0 text-green-500' />
                                <span className='truncate font-medium'>
                                  {rel.source_table}
                                </span>
                                <Badge
                                  variant='outline'
                                  className='shrink-0 px-1 text-[10px]'
                                >
                                  {rel.cardinality === 'one-to-one'
                                    ? '1:1'
                                    : '1:N'}
                                </Badge>
                                <span className='text-muted-foreground truncate'>
                                  {rel.target_column}
                                </span>
                              </div>
                            </TooltipTrigger>
                            <TooltipContent side='left' className='max-w-xs'>
                              <div className='space-y-1 text-xs'>
                                <div className='font-medium'>Incoming FK</div>
                                <div>
                                  {rel.source_schema}.{rel.source_table}.
                                  {rel.source_column} → {rel.target_column}
                                </div>
                                <div>
                                  Cardinality:{' '}
                                  {rel.cardinality === 'many-to-one'
                                    ? 'one-to-many'
                                    : rel.cardinality}
                                </div>
                                <div>ON DELETE: {rel.on_delete}</div>
                                <div>ON UPDATE: {rel.on_update}</div>
                                <div className='text-muted-foreground'>
                                  {rel.constraint_name}
                                </div>
                              </div>
                            </TooltipContent>
                          </Tooltip>
                        </TooltipProvider>
                      ))}
                    </div>
                  </div>
                )}

                {/* Row count */}
                {selectedTableData.row_estimate !== undefined && (
                  <div className='text-muted-foreground text-sm'>
                    ~{selectedTableData.row_estimate.toLocaleString()} rows
                  </div>
                )}
              </CardContent>
            </Card>
          )}
        </div>
      ) : (
        /* List View */
        <Card>
          <CardHeader>
            <CardTitle>Tables and Views</CardTitle>
            <CardDescription>
              {filteredNodes.length} items found
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Schema</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Columns</TableHead>
                  <TableHead>RLS</TableHead>
                  <TableHead>Warnings</TableHead>
                  <TableHead>Relationships</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredNodes.map((node) => {
                  const fullName = getFullName(node)
                  const relationships = data?.edges?.filter(
                    (r) =>
                      (r.source_schema === node.schema &&
                        r.source_table === node.name) ||
                      (r.target_schema === node.schema &&
                        r.target_table === node.name)
                  )
                  return (
                    <TableRow
                      key={fullName}
                      className='cursor-pointer'
                      onClick={() => {
                        setSelectedTable(fullName)
                        setViewMode('erd')
                      }}
                    >
                      <TableCell className='font-medium'>{node.name}</TableCell>
                      <TableCell>
                        <Badge variant='outline'>{node.schema}</Badge>
                      </TableCell>
                      <TableCell>
                        <Badge variant='default'>table</Badge>
                      </TableCell>
                      <TableCell>{node.columns.length}</TableCell>
                      <TableCell>
                        {node.rls_enabled ? (
                          <Shield className='h-4 w-4 text-green-500' />
                        ) : (
                          <ShieldOff className='text-muted-foreground h-4 w-4' />
                        )}
                      </TableCell>
                      <TableCell>
                        {(() => {
                          const count = getTableWarningCount(
                            node.schema,
                            node.name
                          )
                          const severity = getTableWarningSeverity(
                            node.schema,
                            node.name
                          )
                          const warnings = getTableWarnings(
                            node.schema,
                            node.name
                          )
                          if (count > 0) {
                            return (
                              <Popover>
                                <PopoverTrigger asChild>
                                  <Badge
                                    variant={
                                      severity === 'critical' ||
                                      severity === 'high'
                                        ? 'destructive'
                                        : 'secondary'
                                    }
                                    className='cursor-pointer gap-1'
                                    onClick={(e) => e.stopPropagation()}
                                  >
                                    <AlertTriangle className='h-3 w-3' />
                                    {count}
                                  </Badge>
                                </PopoverTrigger>
                                <PopoverContent className='w-72' align='start'>
                                  <div className='space-y-2'>
                                    <h4 className='text-sm font-medium'>
                                      Warnings for {node.name}
                                    </h4>
                                    <div className='max-h-48 space-y-2 overflow-auto'>
                                      {warnings.map((w) => (
                                        <div
                                          key={w.id}
                                          className='border-l-2 border-l-orange-500 pl-2 text-xs'
                                        >
                                          <Badge
                                            variant='outline'
                                            className='mb-1 text-xs'
                                          >
                                            {w.severity}
                                          </Badge>
                                          <p className='text-muted-foreground'>
                                            {w.message}
                                          </p>
                                        </div>
                                      ))}
                                    </div>
                                  </div>
                                </PopoverContent>
                              </Popover>
                            )
                          }
                          return (
                            <span className='text-muted-foreground'>-</span>
                          )
                        })()}
                      </TableCell>
                      <TableCell>
                        {relationships?.length ? (
                          <Badge variant='outline'>
                            {relationships.length}
                          </Badge>
                        ) : (
                          <span className='text-muted-foreground'>-</span>
                        )}
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}

// ERD Canvas Component - Simple CSS-based visualization
interface ERDCanvasProps {
  nodes: SchemaNode[]
  relationships: SchemaRelationship[]
  zoom: number
  onZoomChange: (zoom: number) => void
  selectedTable: string | null
  onSelectTable: (table: string | null) => void
  getTableWarningCount: (schema: string, table: string) => number
  getTableWarningSeverity: (
    schema: string,
    table: string
  ) => 'critical' | 'high' | 'medium' | 'low' | null
  getTableWarnings: (schema: string, table: string) => SecurityWarning[]
}

function ERDCanvas({
  nodes,
  relationships,
  zoom,
  onZoomChange,
  selectedTable,
  onSelectTable,
  getTableWarningCount,
  getTableWarningSeverity,
  getTableWarnings,
}: ERDCanvasProps) {
  // Drag state for moving table cards
  const [draggedNode, setDraggedNode] = useState<string | null>(null)
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 })
  const [nodeOffsets, setNodeOffsets] = useState<Record<string, { x: number; y: number }>>({})

  // Pan state for moving the canvas
  const [panOffset, setPanOffset] = useState({ x: 0, y: 0 })
  const [isPanning, setIsPanning] = useState(false)
  const [panStart, setPanStart] = useState({ x: 0, y: 0 })

  // Ref for the container to calculate mouse position for zoom
  const containerRef = useRef<HTMLDivElement>(null)

  // Helper to get full name for a node
  const getNodeFullName = (node: SchemaNode) => `${node.schema}.${node.name}`

  // Calculate node positions using relationship-based hierarchical layout
  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {}
    const nodeWidth = 340
    const headerHeight = 40
    const columnHeight = 24
    const basePadding = 16
    const gapX = 120 // Horizontal gap between related tables
    const gapY = 60 // Vertical gap between tables in same column

    // Helper to calculate dynamic node height based on column count
    const getNodeHeight = (node: SchemaNode) =>
      headerHeight + node.columns.length * columnHeight + basePadding

    // Build adjacency map from relationships - only for tables in current view
    const outgoing = new Map<string, string[]>() // table -> tables it references
    const incoming = new Map<string, string[]>() // table -> tables that reference it
    const nodeSet = new Set(nodes.map((n) => getNodeFullName(n)))

    relationships.forEach((rel) => {
      const source = `${rel.source_schema}.${rel.source_table}`
      const target = `${rel.target_schema}.${rel.target_table}`
      // Only include if BOTH tables are in current view
      if (nodeSet.has(source) && nodeSet.has(target)) {
        outgoing.set(source, [...(outgoing.get(source) || []), target])
        incoming.set(target, [...(incoming.get(target) || []), source])
      }
    })

    // Find root tables (no outgoing FKs - they are only referenced)
    const roots = nodes.filter((n) => {
      const key = getNodeFullName(n)
      return !outgoing.has(key) || outgoing.get(key)?.length === 0
    })

    // BFS to assign columns (depth from roots)
    const depths = new Map<string, number>()
    const visited = new Set<string>()
    const queue: { key: string; depth: number }[] = []

    // Start with roots at depth 0
    roots.forEach((n) => {
      const key = getNodeFullName(n)
      queue.push({ key, depth: 0 })
    })

    // Also add orphans (no relationships)
    nodes.forEach((n) => {
      const key = getNodeFullName(n)
      if (!outgoing.has(key) && !incoming.has(key)) {
        if (!queue.some((q) => q.key === key)) {
          queue.push({ key, depth: 0 })
        }
      }
    })

    while (queue.length > 0) {
      const { key, depth } = queue.shift()!
      if (visited.has(key)) continue
      visited.add(key)
      depths.set(key, depth)

      // Add tables that reference this one (they go deeper/to the right)
      const refs = incoming.get(key) || []
      refs.forEach((ref) => {
        if (!visited.has(ref)) {
          queue.push({ key: ref, depth: depth + 1 })
        }
      })
    }

    // Handle any unvisited nodes (cycles or disconnected)
    nodes.forEach((n) => {
      const key = getNodeFullName(n)
      if (!depths.has(key)) depths.set(key, 0)
    })

    // Check if all tables are at the same depth (no hierarchy)
    // This happens when tables have FKs pointing to other schemas, not to each other
    const uniqueDepths = new Set(depths.values())
    if (uniqueDepths.size === 1 && nodes.length > 1) {
      // Fall back to grid layout with dynamic row heights
      const cols = Math.ceil(Math.sqrt(nodes.length))
      // Group nodes by row to calculate cumulative heights
      const rows: SchemaNode[][] = []
      nodes.forEach((node, index) => {
        const rowIdx = Math.floor(index / cols)
        if (!rows[rowIdx]) rows[rowIdx] = []
        rows[rowIdx].push(node)
      })

      let cumulativeY = 40
      rows.forEach((rowNodes) => {
        const maxRowHeight = Math.max(...rowNodes.map(getNodeHeight))
        rowNodes.forEach((node, colIdx) => {
          positions[getNodeFullName(node)] = {
            x: colIdx * (nodeWidth + gapX) + 40,
            y: cumulativeY,
          }
        })
        cumulativeY += maxRowHeight + gapY
      })
      return positions
    }

    // Build a map from key to node for height lookups
    const nodeByKey = new Map<string, SchemaNode>()
    nodes.forEach((n) => nodeByKey.set(getNodeFullName(n), n))

    // Group by depth and assign positions
    const columns = new Map<number, string[]>()
    depths.forEach((depth, key) => {
      columns.set(depth, [...(columns.get(depth) || []), key])
    })

    // Sort columns for consistent ordering and use cumulative heights
    const sortedCols = Array.from(columns.keys()).sort((a, b) => a - b)
    sortedCols.forEach((col) => {
      const keys = columns.get(col) || []
      let cumulativeY = 40
      keys.forEach((key) => {
        positions[key] = {
          x: col * (nodeWidth + gapX) + 40,
          y: cumulativeY,
        }
        const node = nodeByKey.get(key)
        const nodeHeight = node ? getNodeHeight(node) : 200
        cumulativeY += nodeHeight + gapY
      })
    })

    return positions
  }, [nodes, relationships])

  // Calculate SVG dimensions based on actual node sizes
  const svgDimensions = useMemo(() => {
    const positions = Object.values(nodePositions)
    if (!positions.length) return { width: 800, height: 600 }
    const nodeWidth = 340
    const headerHeight = 40
    const columnHeight = 24
    const basePadding = 16

    // Find the max extent including the tallest node
    const maxColumns = Math.max(...nodes.map((n) => n.columns.length), 1)
    const maxNodeHeight = headerHeight + maxColumns * columnHeight + basePadding

    const maxX = Math.max(...positions.map((p) => p.x)) + nodeWidth + 40
    const maxY = Math.max(...positions.map((p) => p.y)) + maxNodeHeight + 40
    return { width: maxX, height: maxY }
  }, [nodePositions, nodes])

  // Compute final positions including user drag offsets
  const finalPositions = useMemo(() => {
    const result: Record<string, { x: number; y: number }> = {}
    for (const [key, pos] of Object.entries(nodePositions)) {
      const offset = nodeOffsets[key] || { x: 0, y: 0 }
      result[key] = {
        x: pos.x + offset.x,
        y: pos.y + offset.y,
      }
    }
    return result
  }, [nodePositions, nodeOffsets])

  // Pan handler for moving the canvas
  const handleCanvasMouseDown = useCallback(
    (e: React.MouseEvent) => {
      // Only start panning if clicking directly on the canvas background
      const target = e.target as HTMLElement
      if (target.closest('.table-card')) return // Don't pan when clicking on cards

      e.preventDefault()
      setIsPanning(true)
      setPanStart({
        x: e.clientX - panOffset.x,
        y: e.clientY - panOffset.y,
      })
    },
    [panOffset]
  )

  // Drag handlers for moving table cards
  const handleHeaderMouseDown = useCallback(
    (e: React.MouseEvent, nodeKey: string) => {
      e.preventDefault()
      e.stopPropagation()
      const pos = finalPositions[nodeKey]
      if (!pos) return

      setDraggedNode(nodeKey)
      setDragOffset({
        x: e.clientX / zoom - pos.x,
        y: e.clientY / zoom - pos.y,
      })
    },
    [finalPositions, zoom]
  )

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      // Handle panning
      if (isPanning) {
        setPanOffset({
          x: e.clientX - panStart.x,
          y: e.clientY - panStart.y,
        })
        return
      }

      // Handle card dragging
      if (!draggedNode) return

      const newX = e.clientX / zoom - dragOffset.x
      const newY = e.clientY / zoom - dragOffset.y

      const basePos = nodePositions[draggedNode]
      if (!basePos) return

      setNodeOffsets((prev) => ({
        ...prev,
        [draggedNode]: {
          x: newX - basePos.x,
          y: newY - basePos.y,
        },
      }))
    },
    [isPanning, panStart, draggedNode, dragOffset, nodePositions, zoom]
  )

  const handleMouseUp = useCallback(() => {
    setIsPanning(false)
    setDraggedNode(null)
  }, [])

  // Wheel handler - only zoom on pinch gesture, let regular scroll pan naturally
  // Uses native event listener with { passive: false } to properly prevent browser zoom
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const handleWheel = (e: WheelEvent) => {
      // Don't interfere with table card scrolling
      const target = e.target as HTMLElement
      if (target.closest('.table-card')) {
        return
      }

      // Only zoom on pinch gesture (ctrlKey is true for trackpad pinch)
      if (!e.ctrlKey) {
        // Regular scroll - let native scrolling handle it
        return
      }

      // Pinch-to-zoom - prevent browser zoom
      e.preventDefault()

      // Get mouse position relative to the transformed container's rendered position
      const rect = container.getBoundingClientRect()
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top

      // Calculate the world point under the mouse
      const worldX = mouseX / zoom
      const worldY = mouseY / zoom

      // Pinch zoom is typically more sensitive, use smaller delta
      const zoomDelta = e.deltaY > 0 ? -0.02 : 0.02
      const newZoom = Math.max(0.25, Math.min(2, zoom + zoomDelta))

      // Calculate new pan offset to keep the world point at the same screen position
      const newPanX = mouseX - worldX * newZoom + panOffset.x
      const newPanY = mouseY - worldY * newZoom + panOffset.y

      setPanOffset({ x: newPanX, y: newPanY })
      onZoomChange(newZoom)
    }

    // Prevent Safari's native gesture zoom
    const handleGestureStart = (e: Event) => {
      e.preventDefault()
    }

    // Add with { passive: false } to allow preventDefault
    container.addEventListener('wheel', handleWheel, { passive: false })
    container.addEventListener('gesturestart', handleGestureStart)

    return () => {
      container.removeEventListener('wheel', handleWheel)
      container.removeEventListener('gesturestart', handleGestureStart)
    }
  }, [zoom, onZoomChange, panOffset])

  // Draw relationship lines connecting at specific column positions
  const renderRelationships = useCallback(() => {
    return relationships.map((rel) => {
      const sourcePos =
        finalPositions[`${rel.source_schema}.${rel.source_table}`]
      const targetPos =
        finalPositions[`${rel.target_schema}.${rel.target_table}`]
      if (!sourcePos || !targetPos) return null

      // Find source and target nodes to get column indices
      const sourceNode = nodes.find(
        (n) => n.schema === rel.source_schema && n.name === rel.source_table
      )
      const targetNode = nodes.find(
        (n) => n.schema === rel.target_schema && n.name === rel.target_table
      )

      // Calculate column Y positions
      const headerHeight = 40 // Header section height
      const columnHeight = 24 // Height per column row

      const sourceColIndex = Math.max(
        sourceNode?.columns.findIndex((c) => c.name === rel.source_column) ?? 0,
        0
      )
      const targetColIndex = Math.max(
        targetNode?.columns.findIndex((c) => c.name === rel.target_column) ?? 0,
        0
      )

      // Determine which edges to connect based on relative positions
      const nodeWidth = 340
      const sourceIsLeft = sourcePos.x < targetPos.x

      let sourceX: number, targetX: number
      if (sourceIsLeft) {
        // Source is left of target: connect right edge → left edge
        sourceX = sourcePos.x + nodeWidth
        targetX = targetPos.x
      } else {
        // Source is right of target: connect left edge → right edge
        sourceX = sourcePos.x
        targetX = targetPos.x + nodeWidth
      }

      // Center Y on the column row
      const sourceY =
        sourcePos.y + headerHeight + sourceColIndex * columnHeight + columnHeight / 2
      const targetY =
        targetPos.y + headerHeight + targetColIndex * columnHeight + columnHeight / 2

      // Calculate control point offset based on distance for smooth curves
      const dx = Math.abs(targetX - sourceX)
      const controlOffset = Math.min(dx * 0.4, 80)

      // Create curved path with proper control points
      let path: string
      if (sourceIsLeft) {
        // Source left of target: curve goes right
        path = `M ${sourceX} ${sourceY} C ${sourceX + controlOffset} ${sourceY}, ${targetX - controlOffset} ${targetY}, ${targetX} ${targetY}`
      } else {
        // Source right of target: curve goes left
        path = `M ${sourceX} ${sourceY} C ${sourceX - controlOffset} ${sourceY}, ${targetX + controlOffset} ${targetY}, ${targetX} ${targetY}`
      }

      const isHighlighted =
        selectedTable === `${rel.source_schema}.${rel.source_table}` ||
        selectedTable === `${rel.target_schema}.${rel.target_table}`

      // Use colors visible in both light and dark modes
      const lineColor = isHighlighted ? '#3b82f6' : '#64748b'

      // Crow's foot direction based on table positions
      const footDir = sourceIsLeft ? 1 : -1

      return (
        <g key={rel.id}>
          <path
            d={path}
            fill='none'
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
            strokeDasharray={isHighlighted ? undefined : '6,3'}
          />
          {/* Target end: "one" marker (vertical line) */}
          <line
            x1={targetX + footDir * -8}
            y1={targetY - 6}
            x2={targetX + footDir * -8}
            y2={targetY + 6}
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
          />
          {/* Source end: "many" marker (crow's foot) */}
          <line
            x1={sourceX}
            y1={sourceY}
            x2={sourceX + footDir * 8}
            y2={sourceY - 6}
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
          />
          <line
            x1={sourceX}
            y1={sourceY}
            x2={sourceX + footDir * 8}
            y2={sourceY + 6}
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
          />
        </g>
      )
    })
  }, [relationships, finalPositions, nodes, selectedTable])

  if (!nodes.length) {
    return (
      <div className='text-muted-foreground flex h-full items-center justify-center'>
        No tables found
      </div>
    )
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative min-h-[500px]',
        isPanning && 'cursor-grabbing',
        draggedNode && 'cursor-grabbing',
        !isPanning && !draggedNode && 'cursor-grab'
      )}
      style={{
        transform: `translate(${panOffset.x}px, ${panOffset.y}px) scale(${zoom})`,
        transformOrigin: 'top left',
        width: svgDimensions.width,
        height: svgDimensions.height,
        willChange: 'transform',
      }}
      onMouseDown={handleCanvasMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
    >
      {/* SVG for relationship lines - rendered behind cards */}
      <svg
        className='pointer-events-none'
        style={{
          position: 'absolute',
          left: 0,
          top: 0,
          zIndex: 0,
          overflow: 'visible',
        }}
        width={svgDimensions.width}
        height={svgDimensions.height}
      >
        {renderRelationships()}
      </svg>

      {/* Table nodes */}
      {nodes.map((node) => {
        const fullName = getNodeFullName(node)
        const pos = finalPositions[fullName]
        if (!pos) return null

        const isSelected = selectedTable === fullName
        const isDragging = draggedNode === fullName

        return (
          <div
            key={fullName}
            className={cn(
              'table-card bg-card absolute z-10 w-[340px] rounded-lg border shadow-sm',
              isSelected && 'ring-primary shadow-lg ring-2',
              !isSelected && !isDragging && 'hover:border-primary/50 hover:shadow-md',
              isDragging && 'z-20 shadow-xl ring-primary/50 ring-2'
            )}
            style={{ left: pos.x, top: pos.y }}
            onClick={() => !isDragging && onSelectTable(isSelected ? null : fullName)}
          >
            {/* Header - draggable */}
            <div
              className={cn(
                'flex items-center justify-between rounded-t-lg border-b bg-blue-500/10 px-3 py-2',
                isDragging ? 'cursor-grabbing' : 'cursor-grab'
              )}
              onMouseDown={(e) => handleHeaderMouseDown(e, fullName)}
            >
              <div className='flex items-center gap-2'>
                <Database className='h-4 w-4' />
                <span className='truncate text-sm font-medium'>
                  {node.name}
                </span>
              </div>
              <div className='flex items-center gap-1'>
                {(() => {
                  const count = getTableWarningCount(node.schema, node.name)
                  const severity = getTableWarningSeverity(
                    node.schema,
                    node.name
                  )
                  const warnings = getTableWarnings(node.schema, node.name)
                  if (count > 0) {
                    return (
                      <Popover>
                        <PopoverTrigger
                          asChild
                          onClick={(e) => e.stopPropagation()}
                        >
                          <button className='hover:bg-muted rounded p-0.5'>
                            <AlertTriangle
                              className={cn(
                                'h-3 w-3',
                                severity === 'critical' && 'text-red-500',
                                severity === 'high' && 'text-orange-500',
                                severity === 'medium' && 'text-yellow-500',
                                severity === 'low' && 'text-blue-500'
                              )}
                            />
                          </button>
                        </PopoverTrigger>
                        <PopoverContent className='w-64' align='end'>
                          <div className='space-y-2'>
                            <h4 className='text-sm font-medium'>
                              {count} Warning{count !== 1 ? 's' : ''}
                            </h4>
                            <div className='max-h-40 space-y-2 overflow-auto'>
                              {warnings.map((w) => (
                                <div
                                  key={w.id}
                                  className='border-l-2 border-l-orange-500 pl-2 text-xs'
                                >
                                  <Badge
                                    variant='outline'
                                    className='mb-1 text-xs'
                                  >
                                    {w.severity}
                                  </Badge>
                                  <p className='text-muted-foreground'>
                                    {w.message}
                                  </p>
                                </div>
                              ))}
                            </div>
                          </div>
                        </PopoverContent>
                      </Popover>
                    )
                  }
                  return null
                })()}
                {node.rls_enabled && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Shield className='h-3 w-3 text-green-500' />
                      </TooltipTrigger>
                      <TooltipContent>RLS Enabled</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
                <Badge variant='outline' className='text-xs'>
                  table
                </Badge>
              </div>
            </div>

            {/* Columns */}
            <div className='px-2 py-1'>
              {node.columns.map((col) => (
                <div
                  key={col.name}
                  className='hover:bg-muted flex items-center justify-between rounded px-1 py-1 text-xs'
                >
                  <div className='flex items-center gap-1.5 truncate'>
                    {col.is_primary_key && (
                      <Key className='h-3 w-3 shrink-0 text-yellow-500' />
                    )}
                    {col.is_foreign_key && (
                      <LinkIcon className='h-3 w-3 shrink-0 text-blue-500' />
                    )}
                    <span
                      className={cn(
                        'truncate',
                        col.is_primary_key && 'font-medium'
                      )}
                    >
                      {col.name}
                    </span>
                  </div>
                  <span className='text-muted-foreground ml-2 truncate'>
                    {col.data_type}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}

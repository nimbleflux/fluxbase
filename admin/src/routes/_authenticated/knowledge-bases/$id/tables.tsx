import { useState, useEffect, useCallback } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import {
  RefreshCw,
  Database,
  Loader2,
  Plus,
  Key,
  Link2,
  Settings,
  X,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  knowledgeBasesApi,
  type TableSummary,
  type KnowledgeBase,
  type TableDetails,
  type TableColumn,
  type TableExportSyncConfig,
} from '@/lib/api'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
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
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import { ScrollArea, ScrollBar } from '@/components/ui/scroll-area'
import { KnowledgeBaseHeader } from '@/components/knowledge-bases/knowledge-base-header'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/knowledge-bases/$id/tables')({
  component: KnowledgeBaseTablesPage,
})

interface ExportDialogState {
  open: boolean
  table: TableSummary | null
  tableDetails: TableDetails | null
  selectedColumns: Set<string>
  loading: boolean
  exporting: boolean
}

interface ExportOptions {
  includeForeignKeys: boolean
  includeIndexes: boolean
}

const defaultExportOptions: ExportOptions = {
  includeForeignKeys: true,
  includeIndexes: false,
}

function KnowledgeBaseTablesPage() {
  const { id } = Route.useParams()
  const [knowledgeBase, setKnowledgeBase] = useState<KnowledgeBase | null>(null)
  const [tables, setTables] = useState<TableSummary[]>([])
  const [syncConfigs, setSyncConfigs] = useState<TableExportSyncConfig[]>([])
  const [loading, setLoading] = useState(true)
  const [schemaFilter, setSchemaFilter] = useState<string>('all')
  const [exportDialog, setExportDialog] = useState<ExportDialogState>({
    open: false,
    table: null,
    tableDetails: null,
    selectedColumns: new Set(),
    loading: false,
    exporting: false,
  })
  const [exportOptions, setExportOptions] = useState<ExportOptions>(defaultExportOptions)

  const fetchData = useCallback(async () => {
    setLoading(true)
    try {
      const [kb, tableData, syncs] = await Promise.all([
        knowledgeBasesApi.get(id),
        knowledgeBasesApi.listTables(id, schemaFilter === 'all' ? undefined : schemaFilter),
        knowledgeBasesApi.listTableExportSyncs(id),
      ])
      setKnowledgeBase(kb)
      setTables(tableData)
      setSyncConfigs(syncs)
    } catch {
      toast.error('Failed to fetch data')
    } finally {
      setLoading(false)
    }
  }, [id, schemaFilter])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  const openExportDialog = async (table: TableSummary) => {
    setExportDialog({
      open: true,
      table,
      tableDetails: null,
      selectedColumns: new Set(),
      loading: true,
      exporting: false,
    })
    setExportOptions(defaultExportOptions)

    try {
      const details = await knowledgeBasesApi.getTableDetails(table.schema, table.name)
      const allColumns = new Set(details.columns.map((c) => c.name))
      setExportDialog((prev) => ({
        ...prev,
        tableDetails: details,
        selectedColumns: allColumns,
        loading: false,
      }))
    } catch {
      toast.error('Failed to load table details')
      setExportDialog((prev) => ({ ...prev, open: false, loading: false }))
    }
  }

  const closeExportDialog = () => {
    setExportDialog({
      open: false,
      table: null,
      tableDetails: null,
      selectedColumns: new Set(),
      loading: false,
      exporting: false,
    })
  }

  const toggleColumn = (columnName: string) => {
    setExportDialog((prev) => {
      const newSelected = new Set(prev.selectedColumns)
      if (newSelected.has(columnName)) {
        newSelected.delete(columnName)
      } else {
        newSelected.add(columnName)
      }
      return { ...prev, selectedColumns: newSelected }
    })
  }

  const selectAllColumns = () => {
    if (exportDialog.tableDetails) {
      setExportDialog((prev) => ({
        ...prev,
        selectedColumns: new Set(prev.tableDetails?.columns.map((c) => c.name) || []),
      }))
    }
  }

  const deselectAllColumns = () => {
    setExportDialog((prev) => ({ ...prev, selectedColumns: new Set() }))
  }

  const handleExport = async () => {
    if (!exportDialog.table) return

    const columns = Array.from(exportDialog.selectedColumns)
    if (columns.length === 0) {
      toast.error('Please select at least one column')
      return
    }

    setExportDialog((prev) => ({ ...prev, exporting: true }))

    try {
      await knowledgeBasesApi.exportTable(id, {
        schema: exportDialog.table.schema,
        table: exportDialog.table.name,
        columns,
        include_foreign_keys: exportOptions.includeForeignKeys,
        include_indexes: exportOptions.includeIndexes,
        include_sample_rows: false,
      })
      toast.success(`Table ${exportDialog.table.schema}.${exportDialog.table.name} exported successfully`)
      closeExportDialog()
    } catch {
      toast.error('Failed to export table')
    } finally {
      setExportDialog((prev) => ({ ...prev, exporting: false }))
    }
  }

  const handleTriggerSync = async (sync: TableExportSyncConfig) => {
    try {
      await knowledgeBasesApi.triggerTableExportSync(id, sync.id)
      toast.success(`Sync triggered for ${sync.schema_name}.${sync.table_name}`)
      // Refresh to show updated last_sync_at
      const syncs = await knowledgeBasesApi.listTableExportSyncs(id)
      setSyncConfigs(syncs)
    } catch {
      toast.error('Failed to trigger sync')
    }
  }

  const handleDeleteSync = async (sync: TableExportSyncConfig) => {
    try {
      await knowledgeBasesApi.deleteTableExportSync(id, sync.id)
      toast.success(`Sync configuration deleted`)
      setSyncConfigs((prev) => prev.filter((s) => s.id !== sync.id))
    } catch {
      toast.error('Failed to delete sync configuration')
    }
  }

  // Get unique schemas for filter
  const schemas = ['all', ...Array.from(new Set(tables.map((t) => t.schema))).sort()]

  // Helper to get sync config for a table
  const getSyncForTable = (schema: string, table: string) =>
    syncConfigs.find((s) => s.schema_name === schema && s.table_name === table)

  if (!knowledgeBase && !loading) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <p className="text-muted-foreground">Knowledge base not found</p>
      </div>
    )
  }

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      {knowledgeBase && (
        <KnowledgeBaseHeader
          knowledgeBase={knowledgeBase}
          activeTab="tables"
          actions={
            <>
              <Select value={schemaFilter} onValueChange={setSchemaFilter}>
                <SelectTrigger className="w-[180px]">
                  <SelectValue placeholder="Filter by schema" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Schemas</SelectItem>
                  {schemas
                    .filter((s) => s !== 'all')
                    .map((schema) => (
                      <SelectItem key={schema} value={schema}>
                        {schema}
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
              <Button onClick={fetchData} variant="outline" size="sm">
                <RefreshCw className="mr-2 h-4 w-4" />
                Refresh
              </Button>
            </>
          }
        />
      )}

      <Card>
        <CardHeader>
          <CardTitle>Exportable Tables</CardTitle>
          <CardDescription>
            Export database tables as documents with schema information for AI to
            search and understand your database structure.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="flex h-96 items-center justify-center">
              <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
            </div>
          ) : tables.length === 0 ? (
            <div className="py-12 text-center">
              <Database className="text-muted-foreground mx-auto mb-4 h-12 w-12" />
              <p className="mb-2 text-lg font-medium">No tables found</p>
              <p className="text-muted-foreground text-sm">
                No tables available for export in the selected schema
              </p>
            </div>
          ) : (
            <ScrollArea className="h-[600px]">
              <div className="min-w-[700px]">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Schema</TableHead>
                      <TableHead>Table Name</TableHead>
                      <TableHead>Columns</TableHead>
                      <TableHead>Foreign Keys</TableHead>
                      <TableHead>Last Export</TableHead>
                      <TableHead className="w-[200px]"></TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {tables.map((table) => {
                      const sync = getSyncForTable(table.schema, table.name)
                      // Use last_export from table data (if exported) or last_sync_at from sync config
                      const lastExportTime = table.last_export || sync?.last_sync_at
                      return (
                        <TableRow key={`${table.schema}.${table.name}`}>
                          <TableCell className="font-medium">
                            <Badge variant="outline">{table.schema}</Badge>
                          </TableCell>
                          <TableCell>
                            <code className="text-sm">{table.name}</code>
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary">{table.columns}</Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant="secondary">{table.foreign_keys}</Badge>
                          </TableCell>
                          <TableCell>
                            {lastExportTime ? (
                              <span className="text-sm">
                                {new Date(lastExportTime).toLocaleString()}
                              </span>
                            ) : (
                              <span className="text-muted-foreground text-sm">-</span>
                            )}
                          </TableCell>
                          <TableCell>
                            <div className="flex gap-2">
                              <Button
                                variant="outline"
                                size="sm"
                                onClick={() => openExportDialog(table)}
                              >
                                <Settings className="mr-2 h-4 w-4" />
                                Configure
                              </Button>
                              {sync && (
                                <>
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => handleTriggerSync(sync)}
                                  >
                                    Sync
                                  </Button>
                                  <Button
                                    variant="outline"
                                    size="sm"
                                    onClick={() => handleDeleteSync(sync)}
                                  >
                                    <X className="h-4 w-4" />
                                  </Button>
                                </>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </div>
              <ScrollBar orientation="horizontal" />
            </ScrollArea>
          )}
        </CardContent>
      </Card>

      {/* Export Configuration Dialog */}
      <Dialog open={exportDialog.open} onOpenChange={(open) => !open && closeExportDialog()}>
        <DialogContent className="max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle>
              Export Table: {exportDialog.table?.schema}.{exportDialog.table?.name}
            </DialogTitle>
            <DialogDescription>
              Select columns to include in the export. The table schema will be
              exported as a document for AI to search.
            </DialogDescription>
          </DialogHeader>

          {exportDialog.loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin" />
            </div>
          ) : (
            <div className="flex-1 overflow-auto space-y-6">
              {/* Column Selection */}
              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <Label className="text-base font-semibold">Columns to Export</Label>
                  <div className="flex gap-2">
                    <Button variant="outline" size="sm" onClick={selectAllColumns}>
                      Select All
                    </Button>
                    <Button variant="outline" size="sm" onClick={deselectAllColumns}>
                      Deselect All
                    </Button>
                  </div>
                </div>
                <ScrollArea className="h-[200px] border rounded-md p-2">
                  <div className="space-y-1">
                    {exportDialog.tableDetails?.columns.map((col: TableColumn) => (
                      <label
                        key={col.name}
                        className={cn(
                          'flex items-center gap-3 p-2 rounded hover:bg-muted cursor-pointer',
                          exportDialog.selectedColumns.has(col.name) && 'bg-muted/50'
                        )}
                      >
                        <Checkbox
                          checked={exportDialog.selectedColumns.has(col.name)}
                          onCheckedChange={() => toggleColumn(col.name)}
                        />
                        <span className="font-mono text-sm flex-1">{col.name}</span>
                        <div className="flex items-center gap-1">
                          {col.is_primary_key && (
                            <Badge variant="outline" className="text-xs">
                              <Key className="h-3 w-3 mr-1" />PK
                            </Badge>
                          )}
                          {col.is_foreign_key && (
                            <Badge variant="outline" className="text-xs">
                              <Link2 className="h-3 w-3 mr-1" />FK
                            </Badge>
                          )}
                          <Badge variant="secondary" className="text-xs">
                            {col.data_type}
                          </Badge>
                        </div>
                      </label>
                    ))}
                  </div>
                </ScrollArea>
                <p className="text-muted-foreground text-sm">
                  {exportDialog.selectedColumns.size} of{' '}
                  {exportDialog.tableDetails?.columns.length || 0} columns selected
                </p>
              </div>

              {/* Export Options */}
              <div className="space-y-3">
                <Label className="text-base font-semibold">Export Options</Label>
                <div className="flex gap-4">
                  <label className="flex items-center gap-2">
                    <Checkbox
                      checked={exportOptions.includeForeignKeys}
                      onCheckedChange={(checked) =>
                        setExportOptions((prev) => ({
                          ...prev,
                          includeForeignKeys: checked === true,
                        }))
                      }
                    />
                    <span className="text-sm">Include foreign keys</span>
                  </label>
                  <label className="flex items-center gap-2">
                    <Checkbox
                      checked={exportOptions.includeIndexes}
                      onCheckedChange={(checked) =>
                        setExportOptions((prev) => ({
                          ...prev,
                          includeIndexes: checked === true,
                        }))
                      }
                    />
                    <span className="text-sm">Include indexes</span>
                  </label>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={closeExportDialog}>
              Cancel
            </Button>
            <Button
              onClick={handleExport}
              disabled={
                exportDialog.exporting ||
                exportDialog.selectedColumns.size === 0 ||
                exportDialog.loading
              }
            >
              {exportDialog.exporting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Exporting...
                </>
              ) : (
                <>
                  <Plus className="mr-2 h-4 w-4" />
                  Export Now
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}

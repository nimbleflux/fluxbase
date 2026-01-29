import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import {
  Shield,
  ShieldCheck,
  ShieldOff,
  Plus,
  Trash2,
  Pencil,
  Loader2,
  AlertCircle,
  AlertTriangle,
  Search,
  ChevronDown,
  ChevronRight,
  FileCode,
  Info,
  CheckCircle2,
  XCircle,
  Copy,
} from 'lucide-react'
import { toast } from 'sonner'
import {
  policyApi,
  type RLSPolicy,
  type SecurityWarning,
  type PolicyTemplate,
  type CreatePolicyRequest,
} from '@/lib/api'
import { cn } from '@/lib/utils'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
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
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'

export const Route = createFileRoute('/_authenticated/policies/')({
  component: PoliciesPage,
})

function PoliciesPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [activeTab, setActiveTab] = useState('tables')
  // Policy management modal - replaces the narrow side panel
  const [policyModal, setPolicyModal] = useState<{
    open: boolean
    schema: string
    table: string
    warning?: SecurityWarning | null
  } | null>(null)
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [deleteDialog, setDeleteDialog] = useState<{
    open: boolean
    policy: RLSPolicy | null
  }>({ open: false, policy: null })
  const [editDialog, setEditDialog] = useState<{
    open: boolean
    policy: RLSPolicy | null
  }>({ open: false, policy: null })
  // Template application dialog with table selector
  const [templateDialog, setTemplateDialog] = useState<{
    open: boolean
    template: PolicyTemplate | null
    selectedTable: string
  }>({ open: false, template: null, selectedTable: '' })

  const queryClient = useQueryClient()

  // Fetch tables with RLS status (returns array directly)
  const { data: tablesData, isLoading: tablesLoading } = useQuery({
    queryKey: ['tables-rls'],
    queryFn: () => policyApi.getTablesWithRLS('public'),
  })

  // Fetch security warnings
  const { data: warningsData, isLoading: warningsLoading } = useQuery({
    queryKey: ['security-warnings'],
    queryFn: () => policyApi.getSecurityWarnings(),
  })

  // Fetch policy templates
  const { data: templates } = useQuery({
    queryKey: ['policy-templates'],
    queryFn: () => policyApi.getTemplates(),
  })

  // Fetch selected table details for the modal
  const { data: tableDetails, isLoading: detailsLoading } = useQuery({
    queryKey: ['table-rls-status', policyModal],
    queryFn: () =>
      policyModal
        ? policyApi.getTableRLSStatus(policyModal.schema, policyModal.table)
        : null,
    enabled: !!policyModal?.open,
  })

  // Toggle RLS mutation
  const toggleRLSMutation = useMutation({
    mutationFn: ({
      schema,
      table,
      enable,
      forceRLS,
    }: {
      schema: string
      table: string
      enable: boolean
      forceRLS?: boolean
    }) => policyApi.toggleTableRLS(schema, table, enable, forceRLS),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['tables-rls'] })
      queryClient.invalidateQueries({ queryKey: ['table-rls-status'] })
      queryClient.invalidateQueries({ queryKey: ['security-warnings'] })
      toast.success(data.message)
    },
    onError: () => {
      toast.error('Failed to toggle RLS')
    },
  })

  // Create policy mutation
  const createPolicyMutation = useMutation({
    mutationFn: (data: CreatePolicyRequest) => policyApi.create(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['table-rls-status'] })
      queryClient.invalidateQueries({ queryKey: ['security-warnings'] })
      setCreateDialogOpen(false)
      toast.success(data.message)
    },
    onError: () => {
      toast.error('Failed to create policy')
    },
  })

  // Delete policy mutation
  const deletePolicyMutation = useMutation({
    mutationFn: ({
      schema,
      table,
      name,
    }: {
      schema: string
      table: string
      name: string
    }) => policyApi.delete(schema, table, name),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['table-rls-status'] })
      queryClient.invalidateQueries({ queryKey: ['security-warnings'] })
      setDeleteDialog({ open: false, policy: null })
      toast.success(data.message)
    },
    onError: () => {
      toast.error('Failed to delete policy')
    },
  })

  // Update policy mutation
  const updatePolicyMutation = useMutation({
    mutationFn: ({
      schema,
      table,
      name,
      data,
    }: {
      schema: string
      table: string
      name: string
      data: {
        roles?: string[]
        using?: string | null
        with_check?: string | null
      }
    }) => policyApi.update(schema, table, name, data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ['table-rls-status'] })
      queryClient.invalidateQueries({ queryKey: ['security-warnings'] })
      setEditDialog({ open: false, policy: null })
      toast.success(data.message)
    },
    onError: () => {
      toast.error('Failed to update policy')
    },
  })

  // Filter tables based on search
  const filteredTables = useMemo(() => {
    if (!tablesData) return []
    if (!searchQuery) return tablesData
    const query = searchQuery.toLowerCase()
    return tablesData.filter(
      (t) =>
        t.table.toLowerCase().includes(query) ||
        t.schema.toLowerCase().includes(query)
    )
  }, [tablesData, searchQuery])

  // Sort warnings by severity (descending: critical > high > medium > low)
  const sortedWarnings = useMemo(() => {
    if (!warningsData?.warnings) return []
    const severityOrder: Record<string, number> = {
      critical: 0,
      high: 1,
      medium: 2,
      low: 3,
    }
    return [...warningsData.warnings].sort(
      (a, b) =>
        (severityOrder[a.severity] ?? 4) - (severityOrder[b.severity] ?? 4)
    )
  }, [warningsData])

  // Copy all warnings to clipboard as CSV
  const copyWarningsToClipboard = () => {
    if (!sortedWarnings.length) return
    const header = 'severity,policy_name,table,message,suggestion'
    const csv = sortedWarnings
      .map((w) => {
        // Escape commas and quotes in fields
        const escape = (s: string | undefined) => {
          if (!s) return ''
          if (s.includes(',') || s.includes('"') || s.includes('\n')) {
            return `"${s.replace(/"/g, '""')}"`
          }
          return s
        }
        return [
          escape(w.severity),
          escape(w.policy_name || ''),
          escape(`${w.schema}.${w.table}`),
          escape(w.message),
          escape(w.suggestion),
        ].join(',')
      })
      .join('\n')
    navigator.clipboard.writeText(header + '\n' + csv)
    toast.success(`Copied ${sortedWarnings.length} warnings to clipboard`)
  }

  return (
    <div className='flex flex-1 flex-col gap-6 p-6'>
      {/* Header */}
      <div className='flex items-center justify-between'>
        <div>
          <h1 className='flex items-center gap-2 text-3xl font-bold tracking-tight'>
            <Shield className='h-8 w-8' />
            Row Level Security
          </h1>
          <p className='text-muted-foreground mt-2 text-sm'>
            Manage RLS policies and security settings for your tables
          </p>
        </div>
      </div>

      {/* Security Summary */}
      {warningsData && (
        <div className='grid grid-cols-4 gap-4'>
          <Card
            className={cn(
              warningsData.summary.critical > 0 &&
                'border-red-500/50 bg-red-500/5'
            )}
          >
            <CardHeader className='pb-2'>
              <CardDescription>Critical Issues</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <AlertCircle className='h-5 w-5 text-red-500' />
                {warningsData.summary.critical}
              </CardTitle>
            </CardHeader>
          </Card>
          <Card
            className={cn(
              warningsData.summary.high > 0 &&
                'border-orange-500/50 bg-orange-500/5'
            )}
          >
            <CardHeader className='pb-2'>
              <CardDescription>High Priority</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <AlertTriangle className='h-5 w-5 text-orange-500' />
                {warningsData.summary.high}
              </CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Medium Priority</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <Info className='h-5 w-5 text-yellow-500' />
                {warningsData.summary.medium}
              </CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className='pb-2'>
              <CardDescription>Tables with RLS</CardDescription>
              <CardTitle className='flex items-center gap-2 text-2xl'>
                <ShieldCheck className='h-5 w-5 text-green-500' />
                {tablesData?.filter((t) => t.rls_enabled).length || 0}/
                {tablesData?.length || 0}
              </CardTitle>
            </CardHeader>
          </Card>
        </div>
      )}

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value='tables'>Tables</TabsTrigger>
          <TabsTrigger value='warnings' className='gap-2'>
            Security Warnings
            {warningsData && warningsData.summary.total > 0 && (
              <Badge variant='destructive' className='ml-1'>
                {warningsData.summary.total}
              </Badge>
            )}
          </TabsTrigger>
          <TabsTrigger value='templates'>Policy Templates</TabsTrigger>
        </TabsList>

        {/* Tables Tab */}
        <TabsContent value='tables' className='space-y-4'>
          <div className='flex items-center gap-4'>
            <div className='relative max-w-sm flex-1'>
              <Search className='text-muted-foreground absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2' />
              <Input
                placeholder='Search tables...'
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className='pl-9'
              />
            </div>
          </div>

          <div>
            {/* Tables List */}
            <Card>
              <CardHeader>
                <CardTitle>Tables</CardTitle>
                <CardDescription>
                  Click a table to view and manage its RLS policies
                </CardDescription>
              </CardHeader>
              <CardContent>
                {tablesLoading ? (
                  <div className='flex justify-center py-8'>
                    <Loader2 className='text-muted-foreground h-6 w-6 animate-spin' />
                  </div>
                ) : (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Table</TableHead>
                        <TableHead>Schema</TableHead>
                        <TableHead>RLS</TableHead>
                        <TableHead>Force RLS</TableHead>
                        <TableHead>Policies</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredTables.map((table) => (
                        <TableRow
                          key={`${table.schema}.${table.table}`}
                          className='hover:bg-muted cursor-pointer'
                          onClick={() => {
                            setPolicyModal({
                              open: true,
                              schema: table.schema,
                              table: table.table,
                            })
                          }}
                        >
                          <TableCell className='font-medium'>
                            {table.table}
                          </TableCell>
                          <TableCell>
                            <Badge variant='outline'>{table.schema}</Badge>
                          </TableCell>
                          <TableCell>
                            <Switch
                              checked={table.rls_enabled}
                              onCheckedChange={(checked) =>
                                toggleRLSMutation.mutate({
                                  schema: table.schema,
                                  table: table.table,
                                  enable: checked,
                                })
                              }
                              onClick={(e) => e.stopPropagation()}
                            />
                          </TableCell>
                          <TableCell>
                            {table.rls_forced ? (
                              <CheckCircle2 className='h-4 w-4 text-green-500' />
                            ) : (
                              <XCircle className='text-muted-foreground h-4 w-4' />
                            )}
                          </TableCell>
                          <TableCell>
                            <Badge
                              variant={
                                table.policy_count > 0 ? 'default' : 'secondary'
                              }
                            >
                              {table.policy_count}
                            </Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Security Warnings Tab */}
        <TabsContent value='warnings'>
          <Card>
            <CardHeader>
              <div className='flex items-center justify-between'>
                <div>
                  <CardTitle>Security Warnings</CardTitle>
                  <CardDescription>
                    Issues that may indicate security vulnerabilities in your
                    RLS configuration
                  </CardDescription>
                </div>
                {sortedWarnings.length > 0 && (
                  <Button
                    variant='outline'
                    size='sm'
                    onClick={copyWarningsToClipboard}
                  >
                    <Copy className='mr-2 h-4 w-4' />
                    Copy All
                  </Button>
                )}
              </div>
            </CardHeader>
            <CardContent>
              {warningsLoading ? (
                <div className='flex justify-center py-8'>
                  <Loader2 className='text-muted-foreground h-6 w-6 animate-spin' />
                </div>
              ) : sortedWarnings.length === 0 ? (
                <div className='text-muted-foreground py-12 text-center'>
                  <ShieldCheck className='mx-auto mb-4 h-12 w-12 text-green-500' />
                  <h3 className='text-lg font-medium'>
                    No Security Issues Found
                  </h3>
                  <p className='mt-1 text-sm'>
                    Your RLS configuration looks good
                  </p>
                </div>
              ) : (
                <div className='space-y-3'>
                  {sortedWarnings.map((warning, index) => (
                    <WarningCard
                      key={`${warning.id}-${index}`}
                      warning={warning}
                      onNavigate={() => {
                        setPolicyModal({
                          open: true,
                          schema: warning.schema,
                          table: warning.table,
                          warning: warning,
                        })
                      }}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Templates Tab */}
        <TabsContent value='templates'>
          <Card>
            <CardHeader>
              <CardTitle>Policy Templates</CardTitle>
              <CardDescription>
                Common policy patterns you can use as starting points
              </CardDescription>
            </CardHeader>
            <CardContent>
              {templates?.length === 0 ? (
                <div className='text-muted-foreground py-12 text-center'>
                  <FileCode className='mx-auto mb-4 h-12 w-12' />
                  <h3 className='text-lg font-medium'>
                    No Templates Available
                  </h3>
                </div>
              ) : (
                <div className='grid gap-4 md:grid-cols-2'>
                  {templates?.map((template) => (
                    <TemplateCard
                      key={template.id}
                      template={template}
                      onUse={() => {
                        setTemplateDialog({
                          open: true,
                          template,
                          selectedTable: '',
                        })
                      }}
                    />
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      {/* Policy Management Modal */}
      <PolicyManagementModal
        open={!!policyModal?.open}
        onOpenChange={(open) => !open && setPolicyModal(null)}
        schema={policyModal?.schema || ''}
        table={policyModal?.table || ''}
        warning={policyModal?.warning}
        tableDetails={tableDetails}
        detailsLoading={detailsLoading}
        onToggleRLS={(enable) => {
          if (policyModal) {
            toggleRLSMutation.mutate({
              schema: policyModal.schema,
              table: policyModal.table,
              enable,
            })
          }
        }}
        onEditPolicy={(policy) => setEditDialog({ open: true, policy })}
        onDeletePolicy={(policy) => setDeleteDialog({ open: true, policy })}
        onCreatePolicy={() => setCreateDialogOpen(true)}
      />

      {/* Create Policy Dialog */}
      {policyModal && (
        <CreatePolicyDialog
          open={createDialogOpen}
          onOpenChange={setCreateDialogOpen}
          schema={policyModal.schema}
          table={policyModal.table}
          templates={templates || []}
          onSubmit={(data) => createPolicyMutation.mutate(data)}
          isLoading={createPolicyMutation.isPending}
        />
      )}

      {/* Template Application Dialog */}
      <TemplateApplicationDialog
        open={templateDialog.open}
        onOpenChange={(open) =>
          setTemplateDialog({
            open,
            template: open ? templateDialog.template : null,
            selectedTable: '',
          })
        }
        template={templateDialog.template}
        tables={tablesData || []}
        selectedTable={templateDialog.selectedTable}
        onTableSelect={(table) =>
          setTemplateDialog({ ...templateDialog, selectedTable: table })
        }
        onApply={() => {
          if (templateDialog.template && templateDialog.selectedTable) {
            const [schema, table] = templateDialog.selectedTable.split('.')
            setPolicyModal({ open: true, schema, table })
            setCreateDialogOpen(true)
            setTemplateDialog({
              open: false,
              template: null,
              selectedTable: '',
            })
          }
        }}
      />

      {/* Delete Policy Confirmation */}
      <AlertDialog
        open={deleteDialog.open}
        onOpenChange={(open) =>
          setDeleteDialog({ open, policy: open ? deleteDialog.policy : null })
        }
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Policy</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete the policy &quot;
              {deleteDialog.policy?.policy_name}&quot;? This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => {
                if (deleteDialog.policy) {
                  deletePolicyMutation.mutate({
                    schema: deleteDialog.policy.schema,
                    table: deleteDialog.policy.table,
                    name: deleteDialog.policy.policy_name,
                  })
                }
              }}
              className='bg-destructive text-destructive-foreground hover:bg-destructive/90'
            >
              {deletePolicyMutation.isPending ? (
                <Loader2 className='h-4 w-4 animate-spin' />
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Edit Policy Dialog */}
      {editDialog.policy && (
        <EditPolicyDialog
          open={editDialog.open}
          onOpenChange={(open) =>
            setEditDialog({ open, policy: open ? editDialog.policy : null })
          }
          policy={editDialog.policy}
          onSubmit={(data) => {
            if (editDialog.policy) {
              updatePolicyMutation.mutate({
                schema: editDialog.policy.schema,
                table: editDialog.policy.table,
                name: editDialog.policy.policy_name,
                data,
              })
            }
          }}
          isLoading={updatePolicyMutation.isPending}
        />
      )}
    </div>
  )
}

// Policy Card Component
function PolicyCard({
  policy,
  onEdit,
  onDelete,
}: {
  policy: RLSPolicy
  onEdit: () => void
  onDelete: () => void
}) {
  const [expanded, setExpanded] = useState(false)
  const isPermissive = policy.permissive === 'PERMISSIVE'

  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <div className='rounded-lg border'>
        <CollapsibleTrigger className='hover:bg-muted/50 flex w-full items-center justify-between p-3'>
          <div className='flex items-center gap-3'>
            {expanded ? (
              <ChevronDown className='h-4 w-4' />
            ) : (
              <ChevronRight className='h-4 w-4' />
            )}
            <span className='font-medium'>{policy.policy_name}</span>
            <Badge variant='outline'>{policy.command}</Badge>
            <Badge variant={isPermissive ? 'default' : 'secondary'}>
              {policy.permissive}
            </Badge>
          </div>
          <div className='flex items-center gap-1'>
            <Button
              variant='ghost'
              size='icon'
              className='h-8 w-8'
              onClick={(e) => {
                e.stopPropagation()
                onEdit()
              }}
            >
              <Pencil className='h-4 w-4' />
            </Button>
            <Button
              variant='ghost'
              size='icon'
              className='h-8 w-8'
              onClick={(e) => {
                e.stopPropagation()
                onDelete()
              }}
            >
              <Trash2 className='text-destructive h-4 w-4' />
            </Button>
          </div>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <div className='space-y-3 border-t px-3 pt-0 pb-3'>
            <div className='pt-3'>
              <Label className='text-muted-foreground text-xs'>Roles</Label>
              <div className='mt-1 flex gap-1'>
                {policy.roles.map((role) => (
                  <Badge key={role} variant='secondary'>
                    {role}
                  </Badge>
                ))}
              </div>
            </div>
            {policy.using && (
              <div>
                <Label className='text-muted-foreground text-xs'>
                  USING Expression
                </Label>
                <pre className='bg-muted mt-1 overflow-auto rounded p-2 text-xs'>
                  {policy.using}
                </pre>
              </div>
            )}
            {policy.with_check && (
              <div>
                <Label className='text-muted-foreground text-xs'>
                  WITH CHECK Expression
                </Label>
                <pre className='bg-muted mt-1 overflow-auto rounded p-2 text-xs'>
                  {policy.with_check}
                </pre>
              </div>
            )}
          </div>
        </CollapsibleContent>
      </div>
    </Collapsible>
  )
}

// Warning Card Component
function WarningCard({
  warning,
  onNavigate,
}: {
  warning: SecurityWarning
  onNavigate: () => void
}) {
  const severityColors = {
    critical: 'border-red-500/50 bg-red-500/5',
    high: 'border-orange-500/50 bg-orange-500/5',
    medium: 'border-yellow-500/50 bg-yellow-500/5',
    low: 'border-blue-500/50 bg-blue-500/5',
  }

  const severityIcons = {
    critical: <AlertCircle className='h-5 w-5 text-red-500' />,
    high: <AlertTriangle className='h-5 w-5 text-orange-500' />,
    medium: <Info className='h-5 w-5 text-yellow-500' />,
    low: <Info className='h-5 w-5 text-blue-500' />,
  }

  return (
    <div
      className={cn(
        'cursor-pointer rounded-lg border p-4 transition-shadow hover:shadow-md',
        severityColors[warning.severity]
      )}
      onClick={onNavigate}
    >
      <div className='flex items-start gap-3'>
        {severityIcons[warning.severity]}
        <div className='flex-1'>
          <div className='flex flex-wrap items-center gap-2'>
            <Badge
              variant={
                warning.severity === 'critical' || warning.severity === 'high'
                  ? 'destructive'
                  : 'secondary'
              }
            >
              {warning.severity}
            </Badge>
            <Badge variant='outline'>{warning.category}</Badge>
          </div>
          <p className='mt-2 text-sm'>{warning.message}</p>
          <div className='mt-2 flex items-center gap-2'>
            <Badge variant='outline'>
              {warning.schema}.{warning.table}
            </Badge>
            {warning.policy_name && (
              <Badge variant='secondary'>{warning.policy_name}</Badge>
            )}
          </div>
          <p className='bg-muted mt-2 rounded p-2 text-sm'>
            <strong>Suggestion:</strong> {warning.suggestion}
          </p>
          {warning.fix_sql && (
            <pre className='bg-muted mt-2 overflow-auto rounded p-2 text-xs'>
              {warning.fix_sql}
            </pre>
          )}
        </div>
      </div>
    </div>
  )
}

// Template Card Component
function TemplateCard({
  template,
  onUse,
}: {
  template: PolicyTemplate
  onUse: () => void
}) {
  return (
    <Card>
      <CardHeader className='pb-2'>
        <CardTitle className='text-base'>{template.name}</CardTitle>
        <CardDescription>{template.description}</CardDescription>
      </CardHeader>
      <CardContent>
        <div className='space-y-2'>
          <Badge variant='outline'>{template.command}</Badge>
          <pre className='bg-muted overflow-auto rounded p-2 text-xs'>
            {template.using}
          </pre>
          {template.with_check && (
            <pre className='bg-muted overflow-auto rounded p-2 text-xs'>
              WITH CHECK: {template.with_check}
            </pre>
          )}
          <Button size='sm' onClick={onUse} className='w-full'>
            Use Template
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

// Edit Policy Dialog
// Note: PostgreSQL's ALTER POLICY can only change roles, USING, and WITH CHECK.
// It cannot change the policy name, command type, or permissive/restrictive mode.
function EditPolicyDialog({
  open,
  onOpenChange,
  policy,
  onSubmit,
  isLoading,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  policy: RLSPolicy
  onSubmit: (data: {
    roles?: string[]
    using?: string | null
    with_check?: string | null
  }) => void
  isLoading: boolean
}) {
  const [formData, setFormData] = useState({
    roles: policy.roles,
    using: policy.using || '',
    with_check: policy.with_check || '',
  })

  // Reset form when policy changes
  const [prevPolicy, setPrevPolicy] = useState(policy)
  if (policy !== prevPolicy) {
    setFormData({
      roles: policy.roles,
      using: policy.using || '',
      with_check: policy.with_check || '',
    })
    setPrevPolicy(policy)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      roles: formData.roles,
      using: formData.using || null,
      with_check: formData.with_check || null,
    })
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl'>
        <DialogHeader>
          <DialogTitle>Edit Policy</DialogTitle>
          <DialogDescription>
            Edit the policy &quot;{policy.policy_name}&quot; on {policy.schema}.
            {policy.table}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className='space-y-4'>
          {/* Read-only info */}
          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label>Policy Name</Label>
              <Input value={policy.policy_name} disabled className='bg-muted' />
              <p className='text-muted-foreground text-xs'>
                Policy name cannot be changed
              </p>
            </div>
            <div className='space-y-2'>
              <Label>Command</Label>
              <Input value={policy.command} disabled className='bg-muted' />
              <p className='text-muted-foreground text-xs'>
                Command type cannot be changed
              </p>
            </div>
          </div>

          <div className='space-y-2'>
            <Label>Mode</Label>
            <Input value={policy.permissive} disabled className='bg-muted' />
            <p className='text-muted-foreground text-xs'>
              Permissive/restrictive mode cannot be changed
            </p>
          </div>

          {/* Editable fields */}
          <div className='space-y-2'>
            <Label htmlFor='edit-roles'>Roles (comma-separated)</Label>
            <Input
              id='edit-roles'
              value={formData.roles.join(', ')}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  roles: e.target.value
                    .split(',')
                    .map((r) => r.trim())
                    .filter(Boolean),
                })
              }
              placeholder='e.g., authenticated, anon'
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='edit-using'>USING Expression</Label>
            <Textarea
              id='edit-using'
              value={formData.using}
              onChange={(e) =>
                setFormData({ ...formData, using: e.target.value })
              }
              placeholder='e.g., auth.uid() = user_id'
              className='min-h-[80px] font-mono text-sm'
            />
            <p className='text-muted-foreground text-xs'>
              Controls which rows can be selected, updated, or deleted
            </p>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='edit-with-check'>WITH CHECK Expression</Label>
            <Textarea
              id='edit-with-check'
              value={formData.with_check}
              onChange={(e) =>
                setFormData({ ...formData, with_check: e.target.value })
              }
              placeholder='e.g., auth.uid() = user_id'
              className='min-h-[80px] font-mono text-sm'
            />
            <p className='text-muted-foreground text-xs'>
              Controls which rows can be inserted or updated (new values)
            </p>
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type='submit' disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className='mr-2 h-4 w-4 animate-spin' />
                  Saving...
                </>
              ) : (
                'Save Changes'
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// Create Policy Dialog
function CreatePolicyDialog({
  open,
  onOpenChange,
  schema,
  table,
  templates,
  onSubmit,
  isLoading,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  schema: string
  table: string
  templates: PolicyTemplate[]
  onSubmit: (data: CreatePolicyRequest) => void
  isLoading: boolean
}) {
  const [formData, setFormData] = useState<CreatePolicyRequest>({
    schema,
    table,
    name: '',
    command: 'ALL',
    roles: ['authenticated'],
    using: '',
    with_check: '',
    permissive: true,
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    onSubmit({
      ...formData,
      schema,
      table,
    })
  }

  const handleTemplateSelect = (templateId: string) => {
    const template = templates.find((t) => t.id === templateId)
    if (template) {
      setFormData({
        ...formData,
        command: template.command,
        using: template.using,
        with_check: template.with_check || '',
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-w-2xl'>
        <DialogHeader>
          <DialogTitle>Create Policy</DialogTitle>
          <DialogDescription>
            Create a new RLS policy for {schema}.{table}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={handleSubmit} className='space-y-4'>
          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label htmlFor='name'>Policy Name</Label>
              <Input
                id='name'
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder='e.g., users_select_own'
                required
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='command'>Command</Label>
              <Select
                value={formData.command}
                onValueChange={(value) =>
                  setFormData({ ...formData, command: value })
                }
              >
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value='ALL'>ALL</SelectItem>
                  <SelectItem value='SELECT'>SELECT</SelectItem>
                  <SelectItem value='INSERT'>INSERT</SelectItem>
                  <SelectItem value='UPDATE'>UPDATE</SelectItem>
                  <SelectItem value='DELETE'>DELETE</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>

          {templates.length > 0 && (
            <div className='space-y-2'>
              <Label>Use Template</Label>
              <Select onValueChange={handleTemplateSelect}>
                <SelectTrigger>
                  <SelectValue placeholder='Select a template...' />
                </SelectTrigger>
                <SelectContent>
                  {templates.map((t) => (
                    <SelectItem key={t.id} value={t.id}>
                      {t.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}

          <div className='space-y-2'>
            <Label htmlFor='using'>USING Expression</Label>
            <Textarea
              id='using'
              value={formData.using || ''}
              onChange={(e) =>
                setFormData({ ...formData, using: e.target.value })
              }
              placeholder='e.g., auth.uid() = user_id'
              rows={3}
              className='font-mono text-sm'
            />
            <p className='text-muted-foreground text-xs'>
              Expression that returns true for rows the user can access
            </p>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='check'>WITH CHECK Expression (optional)</Label>
            <Textarea
              id='check'
              value={formData.with_check || ''}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  with_check: e.target.value,
                })
              }
              placeholder='e.g., auth.uid() = user_id'
              rows={3}
              className='font-mono text-sm'
            />
            <p className='text-muted-foreground text-xs'>
              Expression that must be true for new/modified rows (INSERT/UPDATE)
            </p>
          </div>

          <div className='flex items-center justify-between'>
            <div className='flex items-center gap-2'>
              <Switch
                id='permissive'
                checked={formData.permissive}
                onCheckedChange={(checked) =>
                  setFormData({ ...formData, permissive: checked })
                }
              />
              <Label htmlFor='permissive'>Permissive</Label>
            </div>
            <p className='text-muted-foreground text-xs'>
              Permissive policies are combined with OR, restrictive with AND
            </p>
          </div>

          <DialogFooter>
            <Button
              type='button'
              variant='outline'
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type='submit' disabled={isLoading}>
              {isLoading ? (
                <Loader2 className='mr-2 h-4 w-4 animate-spin' />
              ) : null}
              Create Policy
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

// Policy Management Modal - Full-width modal for managing table policies
function PolicyManagementModal({
  open,
  onOpenChange,
  schema,
  table,
  warning,
  tableDetails,
  detailsLoading,
  onToggleRLS,
  onEditPolicy,
  onDeletePolicy,
  onCreatePolicy,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  schema: string
  table: string
  warning?: SecurityWarning | null
  tableDetails:
    | { rls_enabled: boolean; rls_forced: boolean; policies: RLSPolicy[] }
    | null
    | undefined
  detailsLoading: boolean
  onToggleRLS: (enable: boolean) => void
  onEditPolicy: (policy: RLSPolicy) => void
  onDeletePolicy: (policy: RLSPolicy) => void
  onCreatePolicy: () => void
}) {
  if (!schema || !table) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='max-h-[90vh] w-full sm:max-w-5xl overflow-y-auto'>
        <DialogHeader>
          <DialogTitle className='flex items-center gap-2'>
            <Shield className='h-5 w-5' />
            Manage Policies: {table}
          </DialogTitle>
          <DialogDescription>
            {schema}.{table}
          </DialogDescription>
        </DialogHeader>

        {/* Warning context banner (if from warning) */}
        {warning && (
          <div className='rounded-lg border border-orange-500/50 bg-orange-500/10 p-4'>
            <div className='flex items-start gap-3'>
              <AlertTriangle className='mt-0.5 h-5 w-5 shrink-0 text-orange-500' />
              <div className='flex-1'>
                <div className='mb-1 flex items-center gap-2'>
                  <Badge
                    variant={
                      warning.severity === 'critical' ||
                      warning.severity === 'high'
                        ? 'destructive'
                        : 'secondary'
                    }
                  >
                    {warning.severity}
                  </Badge>
                  <Badge variant='outline'>{warning.category}</Badge>
                </div>
                <p className='font-medium'>{warning.message}</p>
                {warning.suggestion && (
                  <p className='text-muted-foreground mt-1 text-sm'>
                    {warning.suggestion}
                  </p>
                )}
                {warning.fix_sql && (
                  <pre className='bg-muted mt-2 overflow-auto rounded p-2 text-xs'>
                    {warning.fix_sql}
                  </pre>
                )}
              </div>
            </div>
          </div>
        )}

        {detailsLoading ? (
          <div className='flex justify-center py-12'>
            <Loader2 className='text-muted-foreground h-8 w-8 animate-spin' />
          </div>
        ) : tableDetails ? (
          <div className='space-y-6'>
            {/* RLS Status */}
            <div className='bg-muted/50 flex items-center justify-between rounded-lg p-4'>
              <div className='flex items-center gap-3'>
                {tableDetails.rls_enabled ? (
                  <ShieldCheck className='h-6 w-6 text-green-500' />
                ) : (
                  <ShieldOff className='text-muted-foreground h-6 w-6' />
                )}
                <div>
                  <div className='text-lg font-medium'>
                    RLS {tableDetails.rls_enabled ? 'Enabled' : 'Disabled'}
                  </div>
                  <div className='text-muted-foreground text-sm'>
                    Force RLS: {tableDetails.rls_forced ? 'Yes' : 'No'}
                  </div>
                </div>
              </div>
              <Switch
                checked={tableDetails.rls_enabled}
                onCheckedChange={onToggleRLS}
              />
            </div>

            {/* Policies Section */}
            <div>
              <div className='mb-4 flex items-center justify-between'>
                <h3 className='text-lg font-semibold'>
                  Policies ({tableDetails.policies.length})
                </h3>
                <Button onClick={onCreatePolicy}>
                  <Plus className='mr-2 h-4 w-4' />
                  Add Policy
                </Button>
              </div>

              {tableDetails.policies.length === 0 ? (
                <div className='rounded-lg border py-12 text-center'>
                  <ShieldOff className='text-muted-foreground mx-auto mb-3 h-12 w-12' />
                  <h4 className='text-lg font-medium'>No policies defined</h4>
                  {tableDetails.rls_enabled && (
                    <p className='text-muted-foreground mt-1 text-sm'>
                      All access will be denied by default when RLS is enabled
                    </p>
                  )}
                  <Button onClick={onCreatePolicy} className='mt-4'>
                    <Plus className='mr-2 h-4 w-4' />
                    Create First Policy
                  </Button>
                </div>
              ) : (
                <div className='space-y-3'>
                  {tableDetails.policies.map((policy) => (
                    <PolicyCard
                      key={policy.policy_name}
                      policy={policy}
                      onEdit={() => onEditPolicy(policy)}
                      onDelete={() => onDeletePolicy(policy)}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}

// Template Application Dialog - Select table for applying a template
function TemplateApplicationDialog({
  open,
  onOpenChange,
  template,
  tables,
  selectedTable,
  onTableSelect,
  onApply,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  template: PolicyTemplate | null
  tables: { schema: string; table: string }[]
  selectedTable: string
  onTableSelect: (table: string) => void
  onApply: () => void
}) {
  if (!template) return null

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className='w-full sm:max-w-2xl'>
        <DialogHeader>
          <DialogTitle>Apply Template: {template.name}</DialogTitle>
          <DialogDescription>
            Select the table to apply this policy template to
          </DialogDescription>
        </DialogHeader>

        <div className='space-y-4'>
          {/* Template Preview */}
          <div className='bg-muted/30 rounded-lg border p-4'>
            <div className='mb-2 flex items-center gap-2'>
              <Badge variant='outline'>{template.command}</Badge>
              <span className='text-muted-foreground text-sm'>
                {template.description}
              </span>
            </div>
            <div className='space-y-2'>
              <div>
                <Label className='text-muted-foreground text-xs'>
                  USING Expression
                </Label>
                <pre className='bg-muted mt-1 overflow-auto rounded p-2 text-xs'>
                  {template.using}
                </pre>
              </div>
              {template.with_check && (
                <div>
                  <Label className='text-muted-foreground text-xs'>
                    WITH CHECK Expression
                  </Label>
                  <pre className='bg-muted mt-1 overflow-auto rounded p-2 text-xs'>
                    {template.with_check}
                  </pre>
                </div>
              )}
            </div>
          </div>

          {/* Table Selector */}
          <div className='space-y-2'>
            <Label>Select Table</Label>
            <Select value={selectedTable} onValueChange={onTableSelect}>
              <SelectTrigger>
                <SelectValue placeholder='Choose a table...' />
              </SelectTrigger>
              <SelectContent>
                {tables.map((t) => (
                  <SelectItem
                    key={`${t.schema}.${t.table}`}
                    value={`${t.schema}.${t.table}`}
                  >
                    {t.schema}.{t.table}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <DialogFooter>
          <Button variant='outline' onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={onApply} disabled={!selectedTable}>
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

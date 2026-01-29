import { useEffect } from 'react'
import { getRouteApi } from '@tanstack/react-router'
import { Panel, PanelGroup, PanelResizeHandle } from 'react-resizable-panels'
import { ImpersonationBanner } from '@/components/impersonation-banner'
import { Main } from '@/components/layout/main'
import { TableSelector } from './components/table-selector'
import { TableViewer } from './components/table-viewer'

const route = getRouteApi('/_authenticated/tables/')

export function Tables() {
  const navigate = route.useNavigate()
  const search = route.useSearch()
  const selectedTable = search.table
  const selectedSchema = search.schema || 'public'

  // Auto-select first table if none selected
  useEffect(() => {
    if (!selectedTable) {
      // This will be handled by the TableSelector component
    }
  }, [selectedTable])

  const handleTableSelect = (table: string) => {
    navigate({
      search: (prev) => ({ ...prev, table, page: 1 }),
    })
  }

  const handleSchemaChange = (schema: string) => {
    navigate({
      search: (prev) => ({ ...prev, schema, table: undefined, page: 1 }),
    })
  }

  return (
    <>
      <ImpersonationBanner />

      <Main className='h-[calc(100vh-4rem)] p-0'>
        <PanelGroup direction='horizontal'>
          <Panel defaultSize={20} minSize={15} maxSize={40}>
            <TableSelector
              selectedTable={selectedTable}
              selectedSchema={selectedSchema}
              onTableSelect={handleTableSelect}
              onSchemaChange={handleSchemaChange}
            />
          </Panel>
          <PanelResizeHandle className='bg-border hover:bg-primary w-1 transition-colors' />
          <Panel>
            <main className='h-full overflow-auto'>
              {selectedTable ? (
                <TableViewer
                  tableName={selectedTable}
                  schema={selectedSchema}
                />
              ) : (
                <div className='flex h-full items-center justify-center'>
                  <p className='text-muted-foreground'>
                    Select a table from the sidebar to view its data
                  </p>
                </div>
              )}
            </main>
          </Panel>
        </PanelGroup>
      </Main>
    </>
  )
}

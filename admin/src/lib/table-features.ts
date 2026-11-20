import {
  createFacetedRowModel,
  createFacetedUniqueValues,
  createFilteredRowModel,
  createPaginatedRowModel,
  createSortedRowModel,
  stockFeatures,
  tableFeatures,
} from '@tanstack/react-table'

/**
 * Shared TanStack Table v9 feature set.
 *
 * v9 requires an explicit `features` option; this bundles the stock features
 * plus the row models every admin table uses (filtered/sorted/paginated +
 * faceted for column filters). Create one per table instance (useMemo) —
 * the object is cheap and keeps per-table memoization scoped correctly.
 *
 * Use `AppTableFeatures` as the leading type argument for `ColumnDef`,
 * `createColumnHelper`, and meta augmentations.
 */
export function createAppTableFeatures() {
  return tableFeatures({
    ...stockFeatures,
    facetedRowModel: createFacetedRowModel(),
    facetedUniqueValues: createFacetedUniqueValues(),
    filteredRowModel: createFilteredRowModel(),
    paginatedRowModel: createPaginatedRowModel(),
    sortedRowModel: createSortedRowModel(),
  })
}

export type AppTableFeatures = ReturnType<typeof createAppTableFeatures>

/**
 * Mirrors table-core's `RowData` (not re-exported by @tanstack/react-table
 * v9). Generic table components must constrain `TData extends TableRowData`
 * with this exact type: v9's `in out` variance requires bidirectional
 * compatibility with the library's own constraint, so anything narrower
 * (e.g. `Record<string, unknown>`) is rejected.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type TableRowData = Record<string, any> | Array<any>

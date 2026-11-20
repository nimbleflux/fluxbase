import type { AppTableFeatures } from '@/lib/table-features'

declare module '@tanstack/react-table' {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TFeatures extends AppTableFeatures, TData, TValue> {
    className?: string // apply to both th and td
    tdClassName?: string
    thClassName?: string
  }
}

export {}

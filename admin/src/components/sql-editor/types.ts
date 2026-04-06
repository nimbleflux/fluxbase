export type EditorMode = "sql" | "graphql";

export interface SQLResult {
  columns?: string[];
  rows?: Record<string, unknown>[];
  row_count: number;
  affected_rows?: number;
  execution_time_ms: number;
  error?: string;
  statement: string;
}

export interface SQLExecutionResponse {
  results: SQLResult[];
}

export interface GraphQLError {
  message: string;
  locations?: Array<{ line: number; column: number }>;
  path?: (string | number)[];
}

export interface GraphQLResponse {
  data?: unknown;
  errors?: GraphQLError[];
}

export interface QueryHistory {
  id: string;
  timestamp: Date;
  mode: EditorMode;
  results?: SQLResult[];
  graphqlResponse?: GraphQLResponse;
  query: string;
  executionTime?: number;
}

export interface HistoryItemProps {
  history: QueryHistory;
  isSelected: boolean;
  onSelect: () => void;
  onRemove: () => void;
}

export interface SQLResultViewProps {
  result: SQLResult;
  currentPage: number;
  totalPages: number;
  paginatedRows: Record<string, unknown>[];
  onExportCSV: () => void;
  onExportJSON: () => void;
  onPrevPage: () => void;
  onNextPage: () => void;
}

export interface GraphQLResultViewProps {
  response: GraphQLResponse;
  executionTime?: number;
}

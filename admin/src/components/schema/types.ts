import type {
  SchemaNode,
  SchemaRelationship,
  SecurityWarning,
} from "@/lib/api";

export type { SchemaNode, SchemaRelationship, SecurityWarning };

export type ViewMode = "erd" | "list";

export type WarningHelpers = {
  getTableWarningCount: (schema: string, table: string) => number;
  getTableWarningSeverity: (
    schema: string,
    table: string,
  ) => "critical" | "high" | "medium" | "low" | null;
  getTableWarnings: (schema: string, table: string) => SecurityWarning[];
};

export interface ERDCanvasProps extends WarningHelpers {
  nodes: SchemaNode[];
  relationships: SchemaRelationship[];
  zoom: number;
  onZoomChange: (zoom: number) => void;
  selectedTable: string | null;
  onSelectTable: (table: string | null) => void;
}

export interface TableDetailsPanelProps extends WarningHelpers {
  selectedTableData: SchemaNode;
  selectedTableRelationships: {
    incoming: SchemaRelationship[];
    outgoing: SchemaRelationship[];
  };
  onSelectTable: (table: string) => void;
}

export interface ListViewProps extends WarningHelpers {
  nodes: SchemaNode[];
  relationships: SchemaRelationship[];
  onSelectTable: (table: string) => void;
}

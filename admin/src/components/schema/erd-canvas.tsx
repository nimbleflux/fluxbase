import { useState, useCallback, useMemo, useRef, useEffect } from "react";
import {
  Database,
  Key,
  Link as LinkIcon,
  Shield,
  AlertTriangle,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { ERDCanvasProps, SchemaNode } from "./types";

function getFullName(node: SchemaNode) {
  return `${node.schema}.${node.name}`;
}

export function ERDCanvas({
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
  const [draggedNode, setDraggedNode] = useState<string | null>(null);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [nodeOffsets, setNodeOffsets] = useState<
    Record<string, { x: number; y: number }>
  >({});

  const [panOffset, setPanOffset] = useState({ x: 0, y: 0 });
  const [isPanning, setIsPanning] = useState(false);
  const [panStart, setPanStart] = useState({ x: 0, y: 0 });

  const containerRef = useRef<HTMLDivElement>(null);

  const nodePositions = useMemo(() => {
    const positions: Record<string, { x: number; y: number }> = {};
    const nodeWidth = 340;
    const headerHeight = 40;
    const columnHeight = 24;
    const basePadding = 16;
    const gapX = 120;
    const gapY = 60;

    const getNodeHeight = (node: SchemaNode) =>
      headerHeight + node.columns.length * columnHeight + basePadding;

    const outgoing = new Map<string, string[]>();
    const incoming = new Map<string, string[]>();
    const nodeSet = new Set(nodes.map((n) => getFullName(n)));

    relationships.forEach((rel) => {
      const source = `${rel.source_schema}.${rel.source_table}`;
      const target = `${rel.target_schema}.${rel.target_table}`;
      if (nodeSet.has(source) && nodeSet.has(target)) {
        outgoing.set(source, [...(outgoing.get(source) || []), target]);
        incoming.set(target, [...(incoming.get(target) || []), source]);
      }
    });

    const roots = nodes.filter((n) => {
      const key = getFullName(n);
      return !outgoing.has(key) || outgoing.get(key)?.length === 0;
    });

    const depths = new Map<string, number>();
    const visited = new Set<string>();
    const queue: { key: string; depth: number }[] = [];

    roots.forEach((n) => {
      const key = getFullName(n);
      queue.push({ key, depth: 0 });
    });

    nodes.forEach((n) => {
      const key = getFullName(n);
      if (!outgoing.has(key) && !incoming.has(key)) {
        if (!queue.some((q) => q.key === key)) {
          queue.push({ key, depth: 0 });
        }
      }
    });

    while (queue.length > 0) {
      const { key, depth } = queue.shift()!;
      if (visited.has(key)) continue;
      visited.add(key);
      depths.set(key, depth);

      const refs = incoming.get(key) || [];
      refs.forEach((ref) => {
        if (!visited.has(ref)) {
          queue.push({ key: ref, depth: depth + 1 });
        }
      });
    }

    nodes.forEach((n) => {
      const key = getFullName(n);
      if (!depths.has(key)) depths.set(key, 0);
    });

    const uniqueDepths = new Set(depths.values());
    if (uniqueDepths.size === 1 && nodes.length > 1) {
      const cols = Math.ceil(Math.sqrt(nodes.length));
      const rows: SchemaNode[][] = [];
      nodes.forEach((node, index) => {
        const rowIdx = Math.floor(index / cols);
        if (!rows[rowIdx]) rows[rowIdx] = [];
        rows[rowIdx].push(node);
      });

      let cumulativeY = 40;
      rows.forEach((rowNodes) => {
        const maxRowHeight = Math.max(...rowNodes.map(getNodeHeight));
        rowNodes.forEach((node, colIdx) => {
          positions[getFullName(node)] = {
            x: colIdx * (nodeWidth + gapX) + 40,
            y: cumulativeY,
          };
        });
        cumulativeY += maxRowHeight + gapY;
      });
      return positions;
    }

    const nodeByKey = new Map<string, SchemaNode>();
    nodes.forEach((n) => nodeByKey.set(getFullName(n), n));

    const columns = new Map<number, string[]>();
    depths.forEach((depth, key) => {
      columns.set(depth, [...(columns.get(depth) || []), key]);
    });

    const sortedCols = Array.from(columns.keys()).sort((a, b) => a - b);
    sortedCols.forEach((col) => {
      const keys = columns.get(col) || [];
      let cumulativeY = 40;
      keys.forEach((key) => {
        positions[key] = {
          x: col * (nodeWidth + gapX) + 40,
          y: cumulativeY,
        };
        const node = nodeByKey.get(key);
        const nodeHeight = node ? getNodeHeight(node) : 200;
        cumulativeY += nodeHeight + gapY;
      });
    });

    return positions;
  }, [nodes, relationships]);

  const svgDimensions = useMemo(() => {
    const positions = Object.values(nodePositions);
    if (!positions.length) return { width: 800, height: 600 };
    const nodeWidth = 340;
    const headerHeight = 40;
    const columnHeight = 24;
    const basePadding = 16;

    const maxColumns = Math.max(...nodes.map((n) => n.columns.length), 1);
    const maxNodeHeight =
      headerHeight + maxColumns * columnHeight + basePadding;

    const maxX = Math.max(...positions.map((p) => p.x)) + nodeWidth + 40;
    const maxY = Math.max(...positions.map((p) => p.y)) + maxNodeHeight + 40;
    return { width: maxX, height: maxY };
  }, [nodePositions, nodes]);

  const finalPositions = useMemo(() => {
    const result: Record<string, { x: number; y: number }> = {};
    for (const [key, pos] of Object.entries(nodePositions)) {
      const offset = nodeOffsets[key] || { x: 0, y: 0 };
      result[key] = {
        x: pos.x + offset.x,
        y: pos.y + offset.y,
      };
    }
    return result;
  }, [nodePositions, nodeOffsets]);

  const handleCanvasMouseDown = useCallback(
    (e: React.MouseEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest(".table-card")) return;

      e.preventDefault();
      setIsPanning(true);
      setPanStart({
        x: e.clientX - panOffset.x,
        y: e.clientY - panOffset.y,
      });
    },
    [panOffset],
  );

  const handleHeaderMouseDown = useCallback(
    (e: React.MouseEvent, nodeKey: string) => {
      e.preventDefault();
      e.stopPropagation();
      const pos = finalPositions[nodeKey];
      if (!pos) return;

      setDraggedNode(nodeKey);
      setDragOffset({
        x: e.clientX / zoom - pos.x,
        y: e.clientY / zoom - pos.y,
      });
    },
    [finalPositions, zoom],
  );

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      if (isPanning) {
        setPanOffset({
          x: e.clientX - panStart.x,
          y: e.clientY - panStart.y,
        });
        return;
      }

      if (!draggedNode) return;

      const newX = e.clientX / zoom - dragOffset.x;
      const newY = e.clientY / zoom - dragOffset.y;

      const basePos = nodePositions[draggedNode];
      if (!basePos) return;

      setNodeOffsets((prev) => ({
        ...prev,
        [draggedNode]: {
          x: newX - basePos.x,
          y: newY - basePos.y,
        },
      }));
    },
    [isPanning, panStart, draggedNode, dragOffset, nodePositions, zoom],
  );

  const handleMouseUp = useCallback(() => {
    setIsPanning(false);
    setDraggedNode(null);
  }, []);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleWheel = (e: WheelEvent) => {
      const target = e.target as HTMLElement;
      if (target.closest(".table-card")) {
        return;
      }

      if (!e.ctrlKey) {
        return;
      }

      e.preventDefault();

      const rect = container.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const mouseY = e.clientY - rect.top;

      const worldX = mouseX / zoom;
      const worldY = mouseY / zoom;

      const zoomDelta = e.deltaY > 0 ? -0.02 : 0.02;
      const newZoom = Math.max(0.25, Math.min(2, zoom + zoomDelta));

      const newPanX = mouseX - worldX * newZoom + panOffset.x;
      const newPanY = mouseY - worldY * newZoom + panOffset.y;

      setPanOffset({ x: newPanX, y: newPanY });
      onZoomChange(newZoom);
    };

    const handleGestureStart = (e: Event) => {
      e.preventDefault();
    };

    container.addEventListener("wheel", handleWheel, { passive: false });
    container.addEventListener("gesturestart", handleGestureStart);

    return () => {
      container.removeEventListener("wheel", handleWheel);
      container.removeEventListener("gesturestart", handleGestureStart);
    };
  }, [zoom, onZoomChange, panOffset]);

  const renderRelationships = useCallback(() => {
    return relationships.map((rel) => {
      const sourcePos =
        finalPositions[`${rel.source_schema}.${rel.source_table}`];
      const targetPos =
        finalPositions[`${rel.target_schema}.${rel.target_table}`];
      if (!sourcePos || !targetPos) return null;

      const sourceNode = nodes.find(
        (n) => n.schema === rel.source_schema && n.name === rel.source_table,
      );
      const targetNode = nodes.find(
        (n) => n.schema === rel.target_schema && n.name === rel.target_table,
      );

      const headerHeight = 40;
      const columnHeight = 24;

      const sourceColIndex = Math.max(
        sourceNode?.columns.findIndex((c) => c.name === rel.source_column) ?? 0,
        0,
      );
      const targetColIndex = Math.max(
        targetNode?.columns.findIndex((c) => c.name === rel.target_column) ?? 0,
        0,
      );

      const nodeWidth = 340;
      const sourceIsLeft = sourcePos.x < targetPos.x;

      let sourceX: number, targetX: number;
      if (sourceIsLeft) {
        sourceX = sourcePos.x + nodeWidth;
        targetX = targetPos.x;
      } else {
        sourceX = sourcePos.x;
        targetX = targetPos.x + nodeWidth;
      }

      const sourceY =
        sourcePos.y +
        headerHeight +
        sourceColIndex * columnHeight +
        columnHeight / 2;
      const targetY =
        targetPos.y +
        headerHeight +
        targetColIndex * columnHeight +
        columnHeight / 2;

      const dx = Math.abs(targetX - sourceX);
      const controlOffset = Math.min(dx * 0.4, 80);

      let path: string;
      if (sourceIsLeft) {
        path = `M ${sourceX} ${sourceY} C ${sourceX + controlOffset} ${sourceY}, ${targetX - controlOffset} ${targetY}, ${targetX} ${targetY}`;
      } else {
        path = `M ${sourceX} ${sourceY} C ${sourceX - controlOffset} ${sourceY}, ${targetX + controlOffset} ${targetY}, ${targetX} ${targetY}`;
      }

      const isHighlighted =
        selectedTable === `${rel.source_schema}.${rel.source_table}` ||
        selectedTable === `${rel.target_schema}.${rel.target_table}`;

      const lineColor = isHighlighted ? "#3b82f6" : "#64748b";

      const footDir = sourceIsLeft ? 1 : -1;

      return (
        <g key={rel.id}>
          <path
            d={path}
            fill="none"
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
            strokeDasharray={isHighlighted ? undefined : "6,3"}
          />
          <line
            x1={targetX + footDir * -8}
            y1={targetY - 6}
            x2={targetX + footDir * -8}
            y2={targetY + 6}
            stroke={lineColor}
            strokeWidth={isHighlighted ? 2.5 : 1.5}
          />
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
      );
    });
  }, [relationships, finalPositions, nodes, selectedTable]);

  if (!nodes.length) {
    return (
      <div className="text-muted-foreground flex h-full items-center justify-center">
        No tables found
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      className={cn(
        "relative min-h-[500px]",
        isPanning && "cursor-grabbing",
        draggedNode && "cursor-grabbing",
        !isPanning && !draggedNode && "cursor-grab",
      )}
      style={{
        transform: `translate(${panOffset.x}px, ${panOffset.y}px) scale(${zoom})`,
        transformOrigin: "top left",
        width: svgDimensions.width,
        height: svgDimensions.height,
        willChange: "transform",
      }}
      onMouseDown={handleCanvasMouseDown}
      onMouseMove={handleMouseMove}
      onMouseUp={handleMouseUp}
      onMouseLeave={handleMouseUp}
    >
      <svg
        className="pointer-events-none"
        style={{
          position: "absolute",
          left: 0,
          top: 0,
          zIndex: 0,
          overflow: "visible",
        }}
        width={svgDimensions.width}
        height={svgDimensions.height}
      >
        {renderRelationships()}
      </svg>

      {nodes.map((node) => {
        const fullName = getFullName(node);
        const pos = finalPositions[fullName];
        if (!pos) return null;

        const isSelected = selectedTable === fullName;
        const isDragging = draggedNode === fullName;

        return (
          <div
            key={fullName}
            className={cn(
              "table-card bg-card absolute z-10 w-[340px] rounded-lg border shadow-sm",
              isSelected && "ring-primary shadow-lg ring-2",
              !isSelected &&
                !isDragging &&
                "hover:border-primary/50 hover:shadow-md",
              isDragging && "z-20 shadow-xl ring-primary/50 ring-2",
            )}
            style={{ left: pos.x, top: pos.y }}
            onClick={() =>
              !isDragging && onSelectTable(isSelected ? null : fullName)
            }
          >
            <div
              className={cn(
                "flex items-center justify-between rounded-t-lg border-b bg-blue-500/10 px-3 py-2",
                isDragging ? "cursor-grabbing" : "cursor-grab",
              )}
              onMouseDown={(e) => handleHeaderMouseDown(e, fullName)}
            >
              <div className="flex items-center gap-2">
                <Database className="h-4 w-4" />
                <span className="truncate text-sm font-medium">
                  {node.name}
                </span>
              </div>
              <div className="flex items-center gap-1">
                {(() => {
                  const count = getTableWarningCount(node.schema, node.name);
                  const severity = getTableWarningSeverity(
                    node.schema,
                    node.name,
                  );
                  const warnings = getTableWarnings(node.schema, node.name);
                  if (count > 0) {
                    return (
                      <Popover>
                        <PopoverTrigger
                          asChild
                          onClick={(e) => e.stopPropagation()}
                        >
                          <button className="hover:bg-muted rounded p-0.5">
                            <AlertTriangle
                              className={cn(
                                "h-3 w-3",
                                severity === "critical" && "text-red-500",
                                severity === "high" && "text-orange-500",
                                severity === "medium" && "text-yellow-500",
                                severity === "low" && "text-blue-500",
                              )}
                            />
                          </button>
                        </PopoverTrigger>
                        <PopoverContent className="w-64" align="end">
                          <div className="space-y-2">
                            <h4 className="text-sm font-medium">
                              {count} Warning{count !== 1 ? "s" : ""}
                            </h4>
                            <div className="max-h-40 space-y-2 overflow-auto">
                              {warnings.map((w) => (
                                <div
                                  key={w.id}
                                  className="border-l-2 border-l-orange-500 pl-2 text-xs"
                                >
                                  <Badge
                                    variant="outline"
                                    className="mb-1 text-xs"
                                  >
                                    {w.severity}
                                  </Badge>
                                  <p className="text-muted-foreground">
                                    {w.message}
                                  </p>
                                </div>
                              ))}
                            </div>
                          </div>
                        </PopoverContent>
                      </Popover>
                    );
                  }
                  return null;
                })()}
                {node.rls_enabled && (
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger>
                        <Shield className="h-3 w-3 text-green-500" />
                      </TooltipTrigger>
                      <TooltipContent>RLS Enabled</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                )}
                <Badge variant="outline" className="text-xs">
                  table
                </Badge>
              </div>
            </div>

            <div className="px-2 py-1">
              {node.columns.map((col) => (
                <div
                  key={col.name}
                  className="hover:bg-muted flex items-center justify-between rounded px-1 py-1 text-xs"
                >
                  <div className="flex items-center gap-1.5 truncate">
                    {col.is_primary_key && (
                      <Key className="h-3 w-3 shrink-0 text-yellow-500" />
                    )}
                    {col.is_foreign_key && (
                      <LinkIcon className="h-3 w-3 shrink-0 text-blue-500" />
                    )}
                    <span
                      className={cn(
                        "truncate",
                        col.is_primary_key && "font-medium",
                      )}
                    >
                      {col.name}
                    </span>
                  </div>
                  <span className="text-muted-foreground ml-2 truncate">
                    {col.data_type}
                  </span>
                </div>
              ))}
            </div>
          </div>
        );
      })}
    </div>
  );
}

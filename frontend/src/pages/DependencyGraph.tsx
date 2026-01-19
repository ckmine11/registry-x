import React, { useCallback, useMemo, useEffect } from 'react';
import ReactFlow, {
    Background,
    Controls,
    MiniMap,
    useNodesState,
    useEdgesState,
    addEdge,
    MarkerType,
    Node,
    Edge,
    Handle,
    Position
} from 'reactflow';
import 'reactflow/dist/style.css';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import { GitBranch, Info, AlertTriangle, Cpu, Zap, Maximize2, Crosshair } from 'lucide-react';
import clsx from 'clsx';

// Custom Node Component for a more futuristic look
const CyberNode = ({ data }: any) => (
    <div className="cyber-card p-4 min-w-[220px] relative overflow-hidden group border-blue-500/30">
        <div className="absolute top-0 left-0 w-full h-0.5 bg-gradient-to-r from-blue-500 to-transparent opacity-50" />
        <Handle type="target" position={Position.Top} className="!bg-blue-500 !border-none !w-2 !h-2" />

        <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform shadow-[0_0_15px_rgba(59,130,246,0.2)]">
                <Cpu size={18} />
            </div>
            <div className="flex-1 min-w-0">
                <div className="text-[10px] font-black text-blue-500 uppercase tracking-widest mb-0.5">IMAGE_ENTITY</div>
                <div className="text-xs font-bold text-white truncate uppercase tracking-tighter">{data.label}</div>
            </div>
        </div>

        <div className="mt-4 flex items-center justify-between text-[8px] font-mono text-gray-500 uppercase tracking-[0.2em]">
            <span>Layer Status</span>
            <span className="text-green-500">Verified</span>
        </div>

        <Handle type="source" position={Position.Bottom} className="!bg-blue-500 !border-none !w-2 !h-2" />
    </div>
);

const nodeTypes = {
    cyber: CyberNode
};

export default function DependencyGraph() {
    const { data: graphData, isLoading } = useQuery({
        queryKey: ['dependency-graph'],
        queryFn: () => api.getDependencyGraph(),
    });

    const initialNodes: Node[] = useMemo(() => {
        if (!graphData?.data?.nodes) return [];
        return graphData.data.nodes.map((node: any, index: number) => ({
            id: node.id,
            type: 'cyber',
            data: { label: `${node.name}:${node.tag}` },
            position: { x: index * 300, y: index * 150 },
        }));
    }, [graphData]);

    const initialEdges: Edge[] = useMemo(() => {
        if (!graphData?.data?.edges) return [];
        return graphData.data.edges.map((edge: any) => ({
            id: `e-${edge.source}-${edge.target}`,
            source: edge.source,
            target: edge.target,
            label: edge.label,
            animated: true,
            labelStyle: { fill: '#3b82f6', fontWeight: 700, fontSize: 10, fontFamily: 'monospace' },
            labelBgPadding: [8, 4],
            labelBgBorderRadius: 4,
            labelBgStyle: { fill: 'rgba(15, 23, 42, 0.8)', stroke: 'rgba(59, 130, 246, 0.2)' },
            style: { stroke: '#3b82f6', strokeWidth: 2 },
            markerEnd: {
                type: MarkerType.ArrowClosed,
                color: '#3b82f6',
                width: 20,
                height: 20
            },
        }));
    }, [graphData]);

    const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);

    useEffect(() => {
        if (initialNodes.length > 0) setNodes(initialNodes);
        if (initialEdges.length > 0) setEdges(initialEdges);
    }, [initialNodes, initialEdges, setNodes, setEdges]);

    const onConnect = useCallback(
        (params: any) => setEdges((eds: any) => addEdge(params, eds)),
        [setEdges]
    );

    return (
        <div className="h-[calc(100vh-120px)] flex flex-col space-y-6">
            {/* Header Readout */}
            <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-6">
                <div>
                    <h1 className="text-4xl font-black uppercase tracking-tighter text-white">Lineage Core</h1>
                    <p className="text-blue-400 font-mono text-xs tracking-[0.3em] uppercase opacity-70">Hierarchical Dependency Mapping System</p>
                </div>

                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-6 px-6 py-3 bg-white/5 border border-white/5 rounded-2xl">
                        <div className="flex flex-col">
                            <span className="text-[8px] font-black text-gray-500 uppercase tracking-widest">Active Nodes</span>
                            <span className="text-lg font-black text-white leading-tight">{nodes.length}</span>
                        </div>
                        <div className="w-px h-8 bg-white/10" />
                        <div className="flex flex-col">
                            <span className="text-[8px] font-black text-gray-500 uppercase tracking-widest">Edges Linked</span>
                            <span className="text-lg font-black text-white leading-tight">{edges.length}</span>
                        </div>
                        <div className="w-px h-8 bg-white/10" />
                        <div className="flex items-center gap-2 px-3 py-1 bg-blue-500/10 rounded-lg text-blue-400 font-mono text-[10px] uppercase font-bold">
                            <Zap size={10} className="animate-pulse" />
                            Stream Live
                        </div>
                    </div>
                </div>
            </div>

            {/* Graph Canvas */}
            <div className="flex-1 cyber-card p-1 relative overflow-hidden group">
                <div className="absolute top-4 left-4 z-50 flex flex-col gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                    <button className="p-2 bg-black/50 hover:bg-black text-gray-400 hover:text-white rounded-lg border border-white/5 transition-all shadow-xl">
                        <Maximize2 size={16} />
                    </button>
                    <button className="p-2 bg-black/50 hover:bg-black text-gray-400 hover:text-white rounded-lg border border-white/5 transition-all shadow-xl">
                        <Crosshair size={16} />
                    </button>
                </div>

                {isLoading ? (
                    <div className="absolute inset-0 flex flex-col items-center justify-center bg-slate-950/50 backdrop-blur-sm z-50">
                        <div className="w-16 h-16 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin mb-4" />
                        <span className="text-[10px] font-mono text-blue-400 uppercase tracking-widest animate-pulse">Syncing Lineage Data...</span>
                    </div>
                ) : nodes.length === 0 ? (
                    <div className="absolute inset-0 flex flex-col items-center justify-center bg-slate-950/50 backdrop-blur-sm z-50 text-center px-10">
                        <div className="w-20 h-20 rounded-3xl bg-white/5 flex items-center justify-center text-gray-700 mb-8 border border-white/5">
                            <GitBranch size={40} />
                        </div>
                        <h2 className="text-2xl font-black text-white uppercase tracking-tighter mb-2">No Lineage Detected</h2>
                        <p className="max-w-xs text-[10px] font-mono text-gray-500 uppercase tracking-widest leading-relaxed">
                            System is standing by. Lineage mappings will initialize automatically upon multi-stage push events.
                        </p>
                    </div>
                ) : (
                    <ReactFlow
                        nodes={nodes}
                        edges={edges}
                        onNodesChange={onNodesChange}
                        onEdgesChange={onEdgesChange}
                        onConnect={onConnect}
                        nodeTypes={nodeTypes}
                        fitView
                        className="bg-slate-950"
                    >
                        <Background
                            color="#334155"
                            gap={40}
                            size={1}
                            variant={BackgroundVariant.Dots}
                            className="opacity-20"
                        />
                        <Controls
                            className="!bg-black/50 !border-white/5 !rounded-xl !overflow-hidden !shadow-2xl"
                            showInteractive={false}
                        />
                        <MiniMap
                            className="!bg-black/80 !border-white/10 !rounded-2xl !shadow-2xl !bottom-6 !right-6"
                            nodeStrokeWidth={3}
                            maskColor="rgba(0, 0, 0, 0.7)"
                            nodeStrokeColor="#3b82f6"
                            nodeColor="#1e293b"
                        />
                    </ReactFlow>
                )}
            </div>

            {/* Tactical Footer Overlay */}
            <div className="flex items-center justify-between text-[10px] font-mono font-bold text-gray-600 uppercase tracking-widest px-2">
                <div className="flex items-center gap-4">
                    <span className="flex items-center gap-1.5"><div className="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.5)]" /> Base Node</span>
                    <span className="flex items-center gap-1.5"><div className="w-1.5 h-1.5 rounded-full bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.5)]" /> Derived Image</span>
                </div>
                <div className="flex items-center gap-4">
                    <span className="animate-pulse flex items-center gap-2">
                        <span className="w-1 h-1 bg-blue-500 rounded-full" />
                        ENGINE_STATE: SECURE
                    </span>
                </div>
            </div>
        </div>
    );
}

// ReactFlow BackgroundVariant enum is needed
import { BackgroundVariant } from 'reactflow';

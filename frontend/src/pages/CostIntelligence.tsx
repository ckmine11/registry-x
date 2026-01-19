import React, { useEffect, useState } from 'react';
import { api } from '../lib/api';
import { DollarSign, TrendingDown, Trash2, RefreshCcw, ShieldAlert, Zap, Layers, BarChart3, ArrowUpRight, Scale } from 'lucide-react';
import { Modal } from '../components/Modal';
import clsx from 'clsx';

interface CostDashboard {
    total_storage_cost_usd: number;
    total_bandwidth_cost_usd: number;
    total_cost_usd: number;
    total_images: number;
    zombie_images: number;
    potential_savings_usd: number;
    top_expensive_images: ImageCost[];
    cost_trend: string;
}

interface ImageCost {
    manifest_id: string;
    repository: string;
    tag: string;
    size_bytes: number;
    storage_cost_usd: number;
    bandwidth_cost_usd: number;
    total_cost_usd: number;
    pull_count_30d: number;
    cost_per_pull: number;
}

interface ZombieImage {
    manifest_id: string;
    repository: string;
    tag: string;
    days_since_last_pull: number;
    storage_cost_usd: number;
    recommended_action: string;
}

export default function CostIntelligence() {
    const [dashboard, setDashboard] = useState<CostDashboard | null>(null);
    const [zombies, setZombies] = useState<ZombieImage[]>([]);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);

    // Modal States
    const [showPurgeModal, setShowPurgeModal] = useState(false);
    const [zombieToDelete, setZombieToDelete] = useState<ZombieImage | null>(null);
    const [alertMessage, setAlertMessage] = useState<{ type: 'success' | 'error', message: string } | null>(null);

    useEffect(() => {
        loadData();
    }, []);

    const loadData = async () => {
        try {
            setLoading(true);
            const [dashData, zombieData] = await Promise.all([
                api.get('/api/v1/costs/dashboard'),
                api.get('/api/v1/costs/zombie-images')
            ]);

            const dashboard = dashData.data || {};
            if (dashboard.top_expensive_images && !Array.isArray(dashboard.top_expensive_images)) {
                dashboard.top_expensive_images = [];
            }

            setDashboard(dashboard);
            setZombies(Array.isArray(zombieData.data) ? zombieData.data : []);
        } catch (error) {
            console.error('Failed to load cost data:', error);
            setDashboard({
                total_storage_cost_usd: 0,
                total_bandwidth_cost_usd: 0,
                total_cost_usd: 0,
                total_images: 0,
                zombie_images: 0,
                potential_savings_usd: 0,
                top_expensive_images: [],
                cost_trend: 'stable'
            });
            setZombies([]);
        } finally {
            setLoading(false);
        }
    };

    const handleRefresh = async () => {
        try {
            setRefreshing(true);
            await api.post('/api/v1/costs/refresh');
            setTimeout(() => {
                loadData();
                setRefreshing(false);
                setAlertMessage({ type: 'success', message: 'COST_METRICS: RECALCULATION_COMPLETE -> DATA_SYNCED' });
            }, 3000);
        } catch (error) {
            console.error('Failed to refresh costs:', error);
            setAlertMessage({ type: 'error', message: 'SYNC_FAILURE: Metric consolidation interrupted.' });
            setRefreshing(false);
        }
    };


    // Renamed to separate trigger from action
    const handleCleanupZombies = (dryRun: boolean) => {
        if (dryRun) {
            executeCleanup(true);
        } else {
            setShowPurgeModal(true);
        }
    };

    const confirmCleanupZombies = () => {
        setShowPurgeModal(false);
        executeCleanup(false);
    };

    const executeCleanup = async (dryRun: boolean) => {
        try {
            const result = await api.post(`/api/v1/costs/cleanup-zombies?dry_run=${dryRun}`);
            const count = result.data?.deleted_count || 0;
            const message = `SYSTEM_RECOVERY: ${dryRun ? 'PROJECTION_ONLY' : 'PURGE_COMPLETE'} -> ${count} ENTITIES`;
            setAlertMessage({ type: 'success', message });
            if (!dryRun) loadData();
        } catch (error) {
            console.error('Failed to cleanup zombies:', error);
            setAlertMessage({ type: 'error', message: 'RECOVERY_FAILURE: Check system protocols.' });
        }
    };

    const handleSingleDelete = async () => {
        if (!zombieToDelete) return;
        try {
            await api.delete(`/api/v1/repositories/${encodeURIComponent(zombieToDelete.repository)}/manifests/${zombieToDelete.manifest_id}`);
            loadData();
            setZombieToDelete(null);
            setAlertMessage({ type: 'success', message: 'Entity Reclaimed Successfully' });
        } catch (error) {
            setAlertMessage({ type: 'error', message: 'ACTION_DENIED: Protocol failure' });
            setZombieToDelete(null);
        }
    };

    const formatCurrency = (amount: number) => {
        return new Intl.NumberFormat('en-US', {
            style: 'currency',
            currency: 'USD',
            minimumFractionDigits: 2
        }).format(amount);
    };

    const formatBytes = (bytes: number) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
    };

    if (loading) {
        return (
            <div className="h-96 flex flex-col items-center justify-center space-y-4">
                <div className="w-16 h-16 border-4 border-blue-500/20 border-t-blue-500 rounded-full animate-spin" />
                <span className="text-sm font-mono text-blue-400 uppercase tracking-widest animate-pulse">Calculating Fiscal Metrics...</span>
            </div>
        );
    }

    if (!dashboard) return null;

    return (
        <div className="space-y-10 pb-20">
            {/* Header Readout */}
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <div>
                    <h1 className="text-4xl font-black uppercase tracking-tighter text-white">Finance Core</h1>
                    <p className="text-blue-400 font-mono text-sm tracking-[0.2em] uppercase opacity-70">Infrastructure Cost Analysis & Yield Optimization</p>
                </div>

                <button
                    onClick={handleRefresh}
                    disabled={refreshing}
                    className={clsx(
                        "flex items-center gap-3 px-6 py-3 rounded-2xl font-black uppercase text-sm tracking-widest transition-all shadow-lg active:scale-95",
                        refreshing ? "bg-white/5 text-gray-500" : "bg-blue-600 text-white hover:bg-blue-500 shadow-blue-500/20"
                    )}
                >
                    <RefreshCcw size={16} className={clsx(refreshing && "animate-spin")} />
                    {refreshing ? 'Syncing...' : 'Force Metric Refresh'}
                </button>
            </div>

            {/* Summary Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    icon={DollarSign}
                    label="ESTIMATED OVERHEAD"
                    value={formatCurrency(dashboard.total_cost_usd)}
                    sub={`Storage: ${formatCurrency(dashboard.total_storage_cost_usd)}`}
                    accent="blue"
                    sub2={`Egress: ${formatCurrency(dashboard.total_bandwidth_cost_usd)}`}
                />
                <StatCard
                    icon={Layers}
                    label="ALLOCATED ASSETS"
                    value={dashboard.total_images.toString()}
                    sub="Active Metadata Entities"
                    accent="green"
                />
                <StatCard
                    icon={ShieldAlert}
                    label="ZOMBIE DIFF"
                    value={dashboard.zombie_images.toString()}
                    sub="Idle > 90 Cycles"
                    accent="red"
                />
                <StatCard
                    icon={TrendingDown}
                    label="RECLAIMABLE YIELD"
                    value={formatCurrency(dashboard.potential_savings_usd)}
                    sub="Optimized Projection"
                    accent="yellow"
                />
            </div>

            {/* Main Analysis Block */}
            <div className="grid grid-cols-1 xl:grid-cols-5 gap-8">
                {/* Expensive Images Table */}
                <div className="xl:col-span-3 cyber-card overflow-hidden">
                    <div className="p-8 border-b border-white/5 flex items-center justify-between">
                        <div className="flex items-center gap-4">
                            <div className="w-10 h-10 rounded-xl bg-blue-600/10 flex items-center justify-center text-blue-400 border border-blue-500/20">
                                <BarChart3 size={20} />
                            </div>
                            <h2 className="text-2xl font-black uppercase tracking-tight text-white">Top Cost Consumers</h2>
                        </div>
                    </div>
                    <div className="overflow-x-auto">
                        <table className="w-full text-sm font-mono uppercase tracking-wider">
                            <thead className="bg-white/[0.02]">
                                <tr>
                                    <th className="px-8 py-5 text-left text-gray-500 font-black">Entity Repository</th>
                                    <th className="px-6 py-5 text-left text-gray-500 font-black">Tag</th>
                                    <th className="px-6 py-5 text-left text-gray-500 font-black">Density</th>
                                    <th className="px-6 py-5 text-left text-gray-500 font-black text-right">Yield Impact</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-white/5">
                                {(dashboard.top_expensive_images || []).map((image: ImageCost, idx: number) => (
                                    <tr key={idx} className="hover:bg-white/[0.01] transition-colors group">
                                        <td className="px-8 py-5 text-white font-bold">{image.repository}</td>
                                        <td className="px-6 py-5 text-gray-400">{image.tag}</td>
                                        <td className="px-6 py-5 text-gray-400">{formatBytes(image.size_bytes)}</td>
                                        <td className="px-10 py-5 text-right font-black text-blue-400 group-hover:text-blue-300">
                                            {formatCurrency(image.total_cost_usd)}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>

                {/* Optimization Console */}
                <div className="xl:col-span-2 space-y-8">
                    <div className="cyber-card p-8 relative overflow-hidden">
                        <div className="absolute top-0 right-0 p-4 opacity-5 text-white"><Scale size={120} /></div>

                        <div className="flex items-center gap-4 mb-8">
                            <div className="w-10 h-10 rounded-xl bg-red-600/10 flex items-center justify-center text-red-500 border border-red-500/20">
                                <TrendingDown size={20} />
                            </div>
                            <h2 className="text-2xl font-black uppercase tracking-tight text-white">Yield Optimizer</h2>
                        </div>

                        <p className="text-sm font-mono text-gray-500 uppercase tracking-widest leading-relaxed mb-8">
                            Detected <span className="text-red-500 font-black">{dashboard.zombie_images} redundant entities</span> consuming <span className="text-white font-black">{formatCurrency(dashboard.potential_savings_usd)}/mo</span>.
                            Automatic recovery protocols ready for initialization.
                        </p>

                        <div className="grid grid-cols-2 gap-4">
                            <button
                                onClick={() => handleCleanupZombies(true)}
                                className="flex flex-col items-center gap-3 p-6 bg-white/5 rounded-2xl border border-white/5 hover:bg-white/10 transition group"
                            >
                                <Zap size={24} className="text-gray-500 group-hover:text-yellow-500 transition-colors" />
                                <span className="text-xs font-black text-white uppercase tracking-widest">Projection Only</span>
                            </button>
                            <button
                                onClick={() => handleCleanupZombies(false)}
                                className="flex flex-col items-center gap-3 p-6 bg-red-600/10 rounded-2xl border border-red-500/20 hover:bg-red-600/20 transition group"
                            >
                                <Trash2 size={24} className="text-red-500 group-hover:scale-110 transition-transform" />
                                <span className="text-xs font-black text-red-500 uppercase tracking-widest">Secure Purge</span>
                            </button>
                        </div>
                    </div>

                    <div className="cyber-card p-8 border-yellow-500/20">
                        <div className="flex items-center gap-3 mb-4">
                            <ShieldAlert size={16} className="text-yellow-500" />
                            <span className="text-xs font-black text-yellow-500 uppercase tracking-widest">Protocol Advisory</span>
                        </div>
                        <p className="text-sm font-mono text-gray-400 uppercase tracking-widest leading-relaxed italic">
                            Cleanup actions are permanent and affect all associated blob manifests. Confirm integrity via lineage map before execution.
                        </p>
                    </div>
                </div>
            </div>

            {/* Zombie List */}
            <div className="cyber-card overflow-hidden">
                <div className="p-8 border-b border-white/5 flex items-center gap-4">
                    <div className="w-10 h-10 rounded-xl bg-yellow-600/10 flex items-center justify-center text-yellow-500 border border-yellow-500/20">
                        <Zap size={20} />
                    </div>
                    <div>
                        <h2 className="text-xl font-black uppercase tracking-tight text-white mb-0.5">Stagnant Lifecycle Entities</h2>
                        <span className="text-[8px] font-black text-gray-500 tracking-[0.3em] uppercase">Images with zero pull metadata for 90+ standard cycles</span>
                    </div>
                </div>
                <div className="overflow-x-auto">
                    <table className="w-full text-[10px] font-mono uppercase tracking-wider">
                        <thead className="bg-white/[0.02]">
                            <tr>
                                <th className="px-8 py-4 text-left text-gray-500 font-black">Identity</th>
                                <th className="px-6 py-4 text-left text-gray-500 font-black">Last Access</th>
                                <th className="px-6 py-4 text-left text-gray-500 font-black">Leak Rate</th>
                                <th className="px-6 py-4 text-right text-gray-500 font-black">Intervention</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-white/5">
                            {(zombies || []).slice(0, 5).map((zombie: ZombieImage, idx: number) => (
                                <tr key={idx} className="hover:bg-white/[0.01] transition-colors">
                                    <td className="px-8 py-4">
                                        <div className="text-white font-bold">{zombie.repository}</div>
                                        <div className="text-[8px] text-gray-500">{zombie.tag}</div>
                                    </td>
                                    <td className="px-6 py-4 text-yellow-500">{zombie.days_since_last_pull} Days</td>
                                    <td className="px-6 py-4 text-gray-400">{formatCurrency(zombie.storage_cost_usd)}/mo</td>
                                    <td className="px-8 py-4 text-right">
                                        <button
                                            onClick={() => setZombieToDelete(zombie)}
                                            className="px-4 py-2 border border-red-500/30 text-red-500 rounded-lg hover:bg-red-500/10 transition-all font-black"
                                        >
                                            RECLAIM
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Modals */}
            <Modal
                isOpen={showPurgeModal}
                onClose={() => setShowPurgeModal(false)}
                title="CRITICAL ACTION: SECURE PURGE"
                variant="danger"
                footer={
                    <>
                        <button onClick={() => setShowPurgeModal(false)} className="px-4 py-2 text-gray-400 hover:text-white transition font-bold uppercase text-xs tracking-wider">Cancel</button>
                        <button
                            onClick={confirmCleanupZombies}
                            className="px-6 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-black uppercase tracking-wider transition shadow-lg shadow-red-600/20"
                        >
                            Confirm Purge
                        </button>
                    </>
                }
            >
                <div>
                    Are you sure you want to delete <span className="text-white font-bold">ALL</span> detected zombie images?
                    <div className="mt-4 p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-xs">
                        This action is irreversible and will permanently remove unreferenced manifests and blobs.
                    </div>
                </div>
            </Modal>

            <Modal
                isOpen={!!zombieToDelete}
                onClose={() => setZombieToDelete(null)}
                title="Single Entity Purge"
                variant="warning"
                footer={
                    <>
                        <button onClick={() => setZombieToDelete(null)} className="px-4 py-2 text-gray-400 hover:text-white transition font-bold uppercase text-xs tracking-wider">Cancel</button>
                        <button
                            onClick={handleSingleDelete}
                            className="px-6 py-2 bg-yellow-600 hover:bg-yellow-500 text-white rounded-lg font-black uppercase tracking-wider transition shadow-lg shadow-yellow-600/20"
                        >
                            Delete Entity
                        </button>
                    </>
                }
            >
                Confirm deletion of <span className="text-white font-bold">{zombieToDelete?.repository}:{zombieToDelete?.tag}</span>?
            </Modal>

            <Modal
                isOpen={!!alertMessage}
                onClose={() => setAlertMessage(null)}
                title={alertMessage?.type === 'success' ? 'Protocol Success' : 'System Alert'}
                variant={alertMessage?.type === 'success' ? 'success' : 'danger'}
                footer={
                    <button onClick={() => setAlertMessage(null)} className="px-6 py-2 bg-white/10 hover:bg-white/20 text-white rounded-lg font-bold uppercase text-xs tracking-wider transition">
                        Acknowledge
                    </button>
                }
            >
                {alertMessage?.message}
            </Modal>
        </div>
    );
}

interface StatCardProps {
    icon: React.ElementType;
    label: string;
    value: string;
    sub?: string;
    sub2?: string;
    accent?: 'blue' | 'green' | 'red' | 'yellow';
}

const StatCard = ({ icon: Icon, label, value, sub, sub2, accent = 'blue' }: StatCardProps) => {
    const colors: Record<string, string> = {
        blue: "text-blue-500 bg-blue-500/10 border-blue-500/20",
        green: "text-green-500 bg-green-500/10 border-green-500/20",
        red: "text-red-500 bg-red-500/10 border-red-500/20",
        yellow: "text-yellow-500 bg-yellow-500/10 border-yellow-500/20"
    };

    const valueColors: Record<string, string> = {
        blue: "text-white",
        green: "text-green-400",
        red: "text-red-400",
        yellow: "text-yellow-400"
    };

    return (
        <div className="cyber-card p-8 group relative overflow-hidden">
            <div className="flex items-start justify-between mb-8">
                <div className={clsx("p-3 rounded-2xl border transition-all group-hover:scale-110", colors[accent])}>
                    <Icon size={24} />
                </div>
                <ArrowUpRight size={16} className="text-gray-600 group-hover:text-white transition-colors" />
            </div>

            <div className="space-y-2">
                <div className="text-[10px] font-black text-gray-500 uppercase tracking-[0.2em]">{label}</div>
                <div className={clsx("text-4xl font-black uppercase tracking-tighter leading-none", valueColors[accent])}>{value}</div>
                <div className="flex flex-col gap-1 pt-2">
                    {sub && <div className="text-[10px] font-mono text-gray-500 uppercase tracking-widest">{sub}</div>}
                    {sub2 && <div className="text-[10px] font-mono text-gray-500 uppercase tracking-widest">{sub2}</div>}
                </div>
            </div>
        </div>
    );
};

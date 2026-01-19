import React from 'react';
import { Activity, Server, Shield, HardDrive, Database, Trash2, AlertTriangle } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../lib/api';
import clsx from 'clsx';

interface RecentPush {
    repository: string;
    tag: string;
    digest: string;
    pushedAt: string;
}

interface DashboardStats {
    repositories: number;
    images: number;
    vulnerabilities: number;
    storageUsed: string;
    recentPushes: RecentPush[];
    severity?: {
        critical: number;
        high: number;
        medium: number;
        low: number;
    };
}

interface GCPreview {
    blobsDeleted: number;
    spaceFreedMB: string;
}

const Dashboard = () => {
    const { data: statsData, isLoading: isStatsLoading } = useQuery({
        queryKey: ['dashboard-stats'],
        queryFn: api.getDashboardStats,
        refetchInterval: 5000
    });

    const [isGCing, setIsGCing] = React.useState(false);
    const [showGCModal, setShowGCModal] = React.useState(false);
    const [gcPreview, setGCPreview] = React.useState<GCPreview | null>(null);
    const [toast, setToast] = React.useState<{ message: string, type: 'success' | 'error' } | null>(null);

    const stats: DashboardStats = statsData?.data || {
        repositories: 0,
        images: 0,
        vulnerabilities: 0,
        storageUsed: '0 GB',
        recentPushes: [],
        severity: { critical: 0, high: 0, medium: 0, low: 0 }
    };

    const handlePreviewGC = async () => {
        setIsGCing(true);
        try {
            const res = await api.runGarbageCollection(true);
            setGCPreview(res.data);
            setShowGCModal(true);
        } catch (err) {
            alert('Failed to preview garbage collection');
        } finally {
            setIsGCing(false);
        }
    };

    const handleConfirmGC = async () => {
        setShowGCModal(false);
        setIsGCing(true);
        try {
            const res = await api.runGarbageCollection(false);
            const report = res.data;
            setToast({
                message: `Garbage collection complete! Freed ${report.spaceFreedMB}`,
                type: 'success'
            });
            setTimeout(() => window.location.reload(), 2000);
        } catch (err) {
            setToast({ message: 'Garbage collection failed', type: 'error' });
        } finally {
            setIsGCing(false);
        }
    };

    return (
        <div className="space-y-8 pb-20">
            {/* Header */}
            <div>
                <h1 className="text-3xl font-bold text-white mb-2">Dashboard</h1>
                <p className="text-gray-400">Overview of your container registry</p>
            </div>

            {/* Toast Notification */}
            {toast && (
                <div className={clsx(
                    "fixed top-4 right-4 z-50 px-6 py-4 rounded-lg shadow-lg",
                    toast.type === 'success' ? 'bg-green-500/10 border border-green-500/20 text-green-400' : 'bg-red-500/10 border border-red-500/20 text-red-400'
                )}>
                    <div className="flex items-center gap-3">
                        <span className="text-sm font-medium">{toast.message}</span>
                        <button onClick={() => setToast(null)} className="text-xs underline">Dismiss</button>
                    </div>
                </div>
            )}

            {/* GC Modal */}
            {showGCModal && gcPreview && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg max-w-md w-full p-6">
                        <div className="flex items-center gap-3 mb-4">
                            <AlertTriangle className="text-yellow-500" size={24} />
                            <h3 className="text-xl font-bold text-white">Confirm Garbage Collection</h3>
                        </div>

                        <p className="text-gray-400 text-sm mb-6">
                            This will permanently delete unreferenced blobs from storage.
                        </p>

                        <div className="grid grid-cols-2 gap-4 mb-6">
                            <div className="bg-gray-800/50 p-4 rounded-lg">
                                <p className="text-gray-500 text-xs mb-1">Blobs to Delete</p>
                                <p className="text-white text-2xl font-bold">{gcPreview.blobsDeleted}</p>
                            </div>
                            <div className="bg-gray-800/50 p-4 rounded-lg">
                                <p className="text-gray-500 text-xs mb-1">Space to Free</p>
                                <p className="text-white text-2xl font-bold">{gcPreview.spaceFreedMB}</p>
                            </div>
                        </div>

                        <div className="flex gap-3">
                            <button
                                onClick={() => setShowGCModal(false)}
                                className="flex-1 px-4 py-2 bg-gray-800 text-white rounded-lg hover:bg-gray-700 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleConfirmGC}
                                className="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                            >
                                Confirm
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <StatCard
                    icon={<Server size={24} />}
                    title="Repositories"
                    value={stats.repositories}
                    color="text-blue-400"
                    bgColor="bg-blue-500/10"
                    isLoading={isStatsLoading}
                />
                <StatCard
                    icon={<Database size={24} />}
                    title="Images"
                    value={stats.images}
                    color="text-emerald-400"
                    bgColor="bg-emerald-500/10"
                    isLoading={isStatsLoading}
                />
                <StatCard
                    icon={<Shield size={24} />}
                    title="Vulnerabilities"
                    value={stats.vulnerabilities}
                    color={stats.vulnerabilities > 0 ? "text-red-400" : "text-green-400"}
                    bgColor={stats.vulnerabilities > 0 ? "bg-red-500/10" : "bg-green-500/10"}
                    isLoading={isStatsLoading}
                />
                <StatCard
                    icon={<HardDrive size={24} />}
                    title="Storage Used"
                    value={stats.storageUsed}
                    color="text-purple-400"
                    bgColor="bg-purple-500/10"
                    isLoading={isStatsLoading}
                    action={
                        <button
                            onClick={handlePreviewGC}
                            disabled={isGCing}
                            className="mt-4 w-full px-3 py-2 bg-purple-600/20 text-purple-400 rounded-lg hover:bg-purple-600/30 transition-colors text-sm flex items-center justify-center gap-2"
                        >
                            <Trash2 size={14} />
                            {isGCing ? 'Running...' : 'Run GC'}
                        </button>
                    }
                />
            </div>

            {/* Main Content Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Recent Activity */}
                <div className="lg:col-span-2 bg-gray-900/50 border border-gray-800 rounded-lg p-6">
                    <h2 className="text-xl font-bold text-white mb-6">Recent Pushes</h2>

                    {stats.recentPushes && stats.recentPushes.length > 0 ? (
                        <div className="space-y-3">
                            {stats.recentPushes.map((push: RecentPush, idx: number) => (
                                <div key={idx} className="flex items-center justify-between p-4 bg-gray-800/50 rounded-lg hover:bg-gray-800 transition-colors">
                                    <div className="flex items-center gap-4">
                                        <div className="w-10 h-10 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-400">
                                            <Database size={20} />
                                        </div>
                                        <div>
                                            <p className="text-white font-medium">{push.repository}</p>
                                            <p className="text-gray-500 text-sm">{push.tag}</p>
                                        </div>
                                    </div>
                                    <div className="text-right">
                                        <p className="text-gray-400 text-xs">
                                            {new Date(push.pushedAt).toLocaleString()}
                                        </p>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="flex flex-col items-center justify-center py-12 text-gray-600">
                            <Database size={48} className="opacity-20 mb-4" />
                            <p className="text-sm">No recent activity</p>
                        </div>
                    )}
                </div>

                {/* Security Overview */}
                <div className="bg-gray-900/50 border border-gray-800 rounded-lg p-6">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-xl font-bold text-white">Security</h2>
                        <Shield className={clsx(
                            "w-6 h-6",
                            stats.vulnerabilities > 0 ? "text-red-500" : "text-green-500"
                        )} />
                    </div>

                    {stats.vulnerabilities > 0 ? (
                        <div className="space-y-4">
                            <div className="text-center p-4 bg-red-500/10 border border-red-500/20 rounded-lg">
                                <p className="text-red-400 text-sm mb-1">Total Vulnerabilities</p>
                                <p className="text-white text-3xl font-bold">{stats.vulnerabilities}</p>
                            </div>

                            <div className="space-y-3">
                                <SeverityBar label="Critical" count={stats.severity?.critical || 0} color="bg-red-500" />
                                <SeverityBar label="High" count={stats.severity?.high || 0} color="bg-orange-500" />
                                <SeverityBar label="Medium" count={stats.severity?.medium || 0} color="bg-yellow-500" />
                                <SeverityBar label="Low" count={stats.severity?.low || 0} color="bg-blue-500" />
                            </div>
                        </div>
                    ) : (
                        <div className="text-center py-8">
                            <div className="w-16 h-16 rounded-full bg-green-500/10 flex items-center justify-center text-green-500 mx-auto mb-4">
                                <Shield size={32} />
                            </div>
                            <p className="text-white font-medium mb-1">All Clear</p>
                            <p className="text-gray-500 text-sm">No vulnerabilities detected</p>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

interface StatCardProps {
    icon: React.ReactNode;
    title: string;
    value: string | number;
    color: string;
    bgColor: string;
    isLoading?: boolean;
    action?: React.ReactNode;
}

const StatCard = ({ icon, title, value, color, bgColor, isLoading, action }: StatCardProps) => (
    <div className="bg-gray-900/50 border border-gray-800 rounded-lg p-6">
        <div className={clsx("w-12 h-12 rounded-lg flex items-center justify-center mb-4", bgColor, color)}>
            {icon}
        </div>
        <p className="text-gray-400 text-sm mb-1">{title}</p>
        {isLoading ? (
            <div className="h-8 bg-gray-800 rounded animate-pulse" />
        ) : (
            <p className="text-white text-2xl font-bold">{value}</p>
        )}
        {action}
    </div>
);

const SeverityBar = ({ label, count, color }: { label: string, count: number, color: string }) => (
    <div>
        <div className="flex justify-between text-sm mb-1">
            <span className="text-gray-400">{label}</span>
            <span className="text-white font-medium">{count}</span>
        </div>
        <div className="w-full bg-gray-800 rounded-full h-2">
            <div className={clsx("h-2 rounded-full", color)} style={{ width: `${Math.min(count * 10, 100)}%` }} />
        </div>
    </div>
);

export default Dashboard;

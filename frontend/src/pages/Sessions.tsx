import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../lib/api';
import { useAuth } from '../lib/auth-context';
import {
    Activity,
    Shield,
    Clock,
    User,
    LogOut,
    Search,
    RefreshCw,
    AlertTriangle,
    CheckCircle2,
    XCircle,
    Monitor,
    ShieldAlert,
    Trash2
} from 'lucide-react';

interface Session {
    id: string;
    user_id: string;
    username: string;
    role: string;
    login_at: string;
    last_active?: string;
}

export default function Sessions() {
    const { user } = useAuth();
    const queryClient = useQueryClient();
    const [searchQuery, setSearchQuery] = useState('');

    const { data: sessions, isLoading, error, refetch } = useQuery({
        queryKey: ['sessions'],
        queryFn: async () => {
            const res = await api.getActiveSessions();
            return res.data as Session[];
        },
        enabled: user?.role === 'admin'
    });

    const revokeMutation = useMutation({
        mutationFn: (id: string) => api.revokeSession(id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['sessions'] });
        }
    });

    if (user?.role !== 'admin') {
        return (
            <div className="flex flex-col items-center justify-center min-h-[60vh] text-center px-4">
                <div className="w-20 h-20 rounded-3xl bg-red-500/10 flex items-center justify-center text-red-500 mb-6 border border-red-500/20">
                    <ShieldAlert size={40} />
                </div>
                <h1 className="text-3xl font-black uppercase tracking-tighter mb-2">Access Denied</h1>
                <p className="text-gray-400 font-mono text-sm max-w-md">
                    You do not have administrative privileges to access the session management dashboard.
                </p>
            </div>
        );
    }

    const filteredSessions = sessions?.filter(s =>
        s.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
        s.id.toLowerCase().includes(searchQuery.toLowerCase())
    );

    return (
        <div className="space-y-8 pb-20">
            {/* Header Section */}
            <div className="flex flex-col md:flex-row md:items-end justify-between gap-6">
                <div>
                    <div className="flex items-center gap-3 mb-2">
                        <div className="p-2 rounded-xl bg-blue-500/10 text-blue-400 border border-blue-500/20">
                            <Activity size={20} />
                        </div>
                        <span className="text-[10px] font-black text-blue-500 uppercase tracking-[0.3em]">System Security</span>
                    </div>
                    <h1 className="text-4xl font-black uppercase tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-white to-gray-500">
                        Active Sessions
                    </h1>
                    <p className="text-gray-500 font-mono text-xs mt-2 uppercase tracking-widest">
                        Monitor and manage live user connections in real-time
                    </p>
                </div>

                <div className="flex items-center gap-3">
                    <div className="bg-white/5 border border-white/5 rounded-2xl px-5 py-3 flex items-center gap-4">
                        <div className="flex flex-col items-end">
                            <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">Live Sessions</span>
                            <span className="text-xl font-black font-mono text-blue-400">{sessions?.length || 0}</span>
                        </div>
                        <div className="w-[1px] h-8 bg-white/5" />
                        <button
                            onClick={() => refetch()}
                            className="p-2 rounded-xl bg-white/5 hover:bg-white/10 text-gray-400 hover:text-white transition-all flex items-center justify-center group"
                        >
                            <RefreshCw size={18} className={`group-hover:rotate-180 transition-transform duration-500 ${isLoading ? 'animate-spin' : ''}`} />
                        </button>
                    </div>
                </div>
            </div>

            {/* Stats Overlays */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <StatCard
                    icon={<Activity className="text-blue-400" />}
                    label="Real-time Connections"
                    value={String(sessions?.length || 0)}
                    desc="Active authenticated sessions"
                />
                <StatCard
                    icon={<Shield className="text-purple-400" />}
                    label="Admin Sessions"
                    value={String(sessions?.filter(s => s.role === 'admin').length || 0)}
                    desc="High privilege access points"
                />
                <StatCard
                    icon={<Clock className="text-green-400" />}
                    label="Avg. Session TTL"
                    value="24h"
                    desc="Standard automatic expiry"
                />
            </div>

            {/* Control Bar */}
            <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
                <div className="relative w-full max-w-md group">
                    <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500 group-focus-within:text-blue-400 transition-colors" />
                    <input
                        type="text"
                        placeholder="Filter sessions by username or ID..."
                        className="w-full bg-white/5 border border-white/5 rounded-2xl py-3 pl-12 pr-6 text-xs font-mono uppercase tracking-widest focus:outline-none focus:border-blue-500/30 focus:bg-white/10 transition-all shadow-inner"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>
            </div>

            {/* Session Table */}
            <div className="bg-black/40 backdrop-blur-xl border border-white/5 rounded-[2.5rem] overflow-hidden shadow-2xl relative">
                <div className="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-blue-500/40 to-transparent" />

                <div className="overflow-x-auto">
                    <table className="w-full text-left border-collapse">
                        <thead>
                            <tr className="border-b border-white/5 bg-white/[0.02]">
                                <th className="px-8 py-6 text-[10px] font-black text-gray-500 uppercase tracking-[0.2em]">Session Identity</th>
                                <th className="px-8 py-6 text-[10px] font-black text-gray-500 uppercase tracking-[0.2em]">Access Role</th>
                                <th className="px-8 py-6 text-[10px] font-black text-gray-500 uppercase tracking-[0.2em]">Initiated At</th>
                                <th className="px-8 py-6 text-[10px] font-black text-gray-500 uppercase tracking-[0.2em]">Status</th>
                                <th className="px-8 py-6 text-[10px] font-black text-gray-500 uppercase tracking-[0.2em] text-right">Emergency Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-white/5 font-mono">
                            {isLoading ? (
                                Array(5).fill(0).map((_, i) => (
                                    <tr key={i} className="animate-pulse">
                                        <td colSpan={5} className="px-8 py-8"><div className="h-8 bg-white/5 rounded-xl w-full" /></td>
                                    </tr>
                                ))
                            ) : filteredSessions?.length === 0 ? (
                                <tr>
                                    <td colSpan={5} className="px-8 py-20 text-center">
                                        <div className="flex flex-col items-center">
                                            <div className="w-16 h-16 rounded-2xl bg-white/5 flex items-center justify-center text-gray-600 mb-4">
                                                <Search size={30} />
                                            </div>
                                            <p className="text-gray-500 font-mono text-sm uppercase tracking-widest">No active sessions found matching criteria</p>
                                        </div>
                                    </td>
                                </tr>
                            ) : (
                                filteredSessions?.map((session) => (
                                    <tr key={session.id} className="group hover:bg-white/[0.02] transition-colors relative">
                                        <td className="px-8 py-6">
                                            <div className="flex items-center gap-4">
                                                <div className="w-10 h-10 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-400 group-hover:bg-blue-500/20 transition-colors">
                                                    <Monitor size={18} />
                                                </div>
                                                <div className="flex flex-col">
                                                    <span className="text-sm font-black text-white group-hover:text-blue-400 transition-colors">
                                                        {session.username}
                                                    </span>
                                                    <span className="text-[10px] text-gray-600 tracking-tighter">
                                                        ID: {session.id.substring(0, 8)}...
                                                    </span>
                                                </div>
                                            </div>
                                        </td>
                                        <td className="px-8 py-6">
                                            <span className={`px-2 py-1 rounded-md text-[9px] font-black uppercase tracking-widest ${session.role === 'admin'
                                                    ? 'bg-purple-500/10 text-purple-400 border border-purple-500/20'
                                                    : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'
                                                }`}>
                                                {session.role}
                                            </span>
                                        </td>
                                        <td className="px-8 py-6">
                                            <div className="flex flex-col">
                                                <span className="text-xs text-gray-400">
                                                    {new Date(session.login_at).toLocaleDateString()}
                                                </span>
                                                <span className="text-[10px] text-gray-600 font-bold">
                                                    {new Date(session.login_at).toLocaleTimeString()}
                                                </span>
                                            </div>
                                        </td>
                                        <td className="px-8 py-6">
                                            <div className="flex items-center gap-2">
                                                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                                                <span className="text-[10px] font-black text-green-500 uppercase tracking-widest">Active</span>
                                            </div>
                                        </td>
                                        <td className="px-8 py-6 text-right">
                                            <button
                                                disabled={revokeMutation.isPending}
                                                onClick={() => {
                                                    if (confirm(`Are you sure you want to terminate the session for ${session.username}? The user will be instantly logged out.`)) {
                                                        revokeMutation.mutate(session.id);
                                                    }
                                                }}
                                                className="inline-flex items-center gap-2 px-4 py-2 rounded-xl bg-red-500/10 hover:bg-red-500 border border-red-500/20 hover:border-red-500 text-red-500 hover:text-white transition-all duration-300 font-black uppercase text-[10px] tracking-widest disabled:opacity-50"
                                            >
                                                {revokeMutation.isPending ? <RefreshCw size={14} className="animate-spin" /> : <Trash2 size={14} />}
                                                Revoke Access
                                            </button>
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            {/* Extra Info */}
            <div className="bg-orange-500/5 border border-orange-500/10 rounded-3xl p-6 flex flex-col md:flex-row items-center gap-6">
                <div className="w-12 h-12 rounded-2xl bg-orange-500/10 flex items-center justify-center text-orange-500 shrink-0">
                    <AlertTriangle size={24} />
                </div>
                <div>
                    <h4 className="text-sm font-black text-orange-400 uppercase tracking-widest mb-1">Administrative Note</h4>
                    <p className="text-xs text-orange-500/70 font-mono leading-relaxed uppercase tracking-tight">
                        Revoking a session will delete the session token from the Redis cluster.
                        Users will receive a 401 Unauthorized error on their next request and will be
                        automatically redirected to the login flow. Active image pushes or pulls will be terminated immediately.
                    </p>
                </div>
            </div>
        </div>
    );
}

function StatCard({ icon, label, value, desc }: { icon: React.ReactNode, label: string, value: string, desc: string }) {
    return (
        <div className="bg-white/5 border border-white/5 p-6 rounded-[2rem] group hover:border-blue-500/30 transition-all relative overflow-hidden">
            <div className="absolute top-0 right-0 p-6 opacity-10 group-hover:scale-125 transition-transform duration-500">
                {React.cloneElement(icon as React.ReactElement, { size: 64 })}
            </div>
            <div className="relative space-y-4">
                <div className="flex items-center gap-3">
                    <div className="w-8 h-8 rounded-xl bg-white/5 flex items-center justify-center">
                        {React.cloneElement(icon as React.ReactElement, { size: 16 })}
                    </div>
                    <span className="text-[10px] font-black text-gray-500 uppercase tracking-widest">{label}</span>
                </div>
                <div className="flex flex-col">
                    <span className="text-3xl font-black font-mono text-white tracking-tighter">{value}</span>
                    <span className="text-[10px] text-gray-600 font-mono mt-1 uppercase tracking-tight">{desc}</span>
                </div>
            </div>
        </div>
    );
}

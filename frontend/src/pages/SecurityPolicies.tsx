import React, { useState, useEffect } from 'react';
import {
    Shield, AlertTriangle, AlertCircle, Info, Save,
    CheckCircle, Lock, RefreshCw, Layers, Plus,
    Trash2, Search, Zap, Layout, ArrowRight, ShieldCheck,
    MoreVertical, Edit2
} from 'lucide-react';
import { api, SecurityPolicyConfig as SecurityPolicy, RepositoryOverride } from '../lib/api';

const SecurityPolicies: React.FC = () => {
    const [activeTab, setActiveTab] = useState<'global' | 'overrides'>('global');
    const [policy, setPolicy] = useState<SecurityPolicy | null>(null);
    const [overrides, setOverrides] = useState<RepositoryOverride[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

    // New Override State
    const [showAddOverride, setShowAddOverride] = useState(false);
    const [newRepoPath, setNewRepoPath] = useState('');
    const [newRepoPolicy, setNewRepoPolicy] = useState<SecurityPolicy>({
        critical_threshold: 0,
        high_threshold: 5,
        medium_threshold: 10,
        low_threshold: 100,
        block_unscanned: true
    });

    useEffect(() => {
        fetchData();
    }, []);

    const fetchData = async () => {
        setLoading(true);
        try {
            const policyRes = await api.getSecurityPolicy().catch(err => {
                console.error("Failed to fetch global policy", err);
                return { data: { critical_threshold: 0, high_threshold: 0, medium_threshold: 0, low_threshold: 0, block_unscanned: false } };
            });
            const overridesRes = await api.listRepositoryOverrides().catch(err => {
                console.error("Failed to fetch overrides", err);
                return { data: [] };
            });

            setPolicy(policyRes.data);
            setOverrides(overridesRes.data || []);
        } catch (err) {
            console.error("Failed to fetch security data", err);
        } finally {
            setLoading(false);
        }
    };

    const handleSaveGlobal = async () => {
        if (!policy) return;
        setSaving(true);
        setMessage(null);
        try {
            await api.updateSecurityPolicy(policy);
            setMessage({ type: 'success', text: 'Global security standards deployed successfully!' });
            setTimeout(() => setMessage(null), 3000);
        } catch (err: any) {
            setMessage({ type: 'error', text: 'Failed to deploy global standards.' });
        } finally {
            setSaving(false);
        }
    };

    const handleAddOverride = async () => {
        if (!newRepoPath) return;
        setSaving(true);
        try {
            await api.updateRepositoryOverride(newRepoPath, newRepoPolicy);
            await fetchData();
            setShowAddOverride(false);
            setNewRepoPath('');
            setMessage({ type: 'success', text: `Override added for ${newRepoPath}` });
            setTimeout(() => setMessage(null), 3000);
        } catch (err) {
            setMessage({ type: 'error', text: 'Failed to add repository override.' });
        } finally {
            setSaving(false);
        }
    };

    const handleDeleteOverride = async (path: string) => {
        try {
            await api.deleteRepositoryOverride(path);
            await fetchData();
            setMessage({ type: 'success', text: 'Repository override removed.' });
            setTimeout(() => setMessage(null), 3000);
        } catch (err) {
            setMessage({ type: 'error', text: 'Failed to remove override.' });
        }
    };

    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center min-h-[500px] space-y-4">
                <div className="relative">
                    <RefreshCw className="w-12 h-12 animate-spin text-indigo-500" />
                    <Shield className="w-6 h-6 text-indigo-400 absolute inset-0 m-auto" />
                </div>
                <p className="text-slate-400 font-medium animate-pulse">Syncing Security Matrix...</p>
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto pb-20">
            {/* Header with Background Glow */}
            <div className="relative mb-12">
                <div className="absolute -top-20 -left-20 w-64 h-64 bg-indigo-500/10 rounded-full blur-3xl" />
                <div className="absolute -top-20 -right-20 w-64 h-64 bg-cyan-500/10 rounded-full blur-3xl" />

                <div className="relative flex flex-col md:flex-row md:items-end justify-between gap-6">
                    <div className="flex-1">
                        <div className="flex items-center gap-2 mb-3">
                            <span className="px-2 py-0.5 bg-indigo-500/10 text-indigo-400 text-[10px] font-bold uppercase tracking-widest rounded-full border border-indigo-500/20">
                                Security Core v2.5
                            </span>
                        </div>
                        <h1 className="text-4xl font-black text-white tracking-tight flex items-center gap-4">
                            <ShieldCheck className="w-10 h-10 text-indigo-500" />
                            Security Policies
                        </h1>
                        <p className="text-slate-400 mt-3 text-lg max-w-2xl leading-relaxed">
                            Define the automated defense perimeter for RegistryX. Configure global severity thresholds
                            and fine-tune enforcement for specific mission-critical repositories.
                        </p>
                    </div>

                    <div className="flex items-center gap-3">
                        <div className="hidden lg:block text-right mr-4">
                            <p className="text-xs font-bold text-slate-500 uppercase tracking-widest">Global Status</p>
                            <p className="text-sm font-bold text-emerald-400 flex items-center gap-1 justify-end">
                                <Zap className="w-3 h-3 fill-current" /> Active Defense
                            </p>
                        </div>
                        {activeTab === 'global' && (
                            <button
                                onClick={handleSaveGlobal}
                                disabled={saving}
                                className="group flex items-center gap-2 px-8 py-3.5 bg-gradient-to-r from-indigo-600 to-indigo-500 text-white rounded-xl font-bold uppercase tracking-wider text-sm hover:from-indigo-500 hover:to-indigo-400 transition-all shadow-xl shadow-indigo-500/20 disabled:opacity-50 active:scale-95"
                            >
                                {saving ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4 group-hover:scale-110 transition-transform" />}
                                Deploy Changes
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {message && (
                <div className={`mb-8 p-4 rounded-2xl border backdrop-blur-md flex items-center justify-between gap-3 animate-in fade-in zoom-in duration-300 ${message.type === 'success'
                    ? 'bg-emerald-500/5 border-emerald-500/20 text-emerald-400'
                    : 'bg-rose-500/5 border-rose-500/20 text-rose-400'
                    }`}>
                    <div className="flex items-center gap-3">
                        <div className={`p-2 rounded-lg ${message.type === 'success' ? 'bg-emerald-500/20' : 'bg-rose-500/20'}`}>
                            {message.type === 'success' ? <CheckCircle className="w-5 h-5" /> : <AlertCircle className="w-5 h-5" />}
                        </div>
                        <span className="font-semibold">{message.text}</span>
                    </div>
                    <button onClick={() => setMessage(null)} className="text-slate-500 hover:text-white transition-colors">
                        <Plus className="w-5 h-5 rotate-45" />
                    </button>
                </div>
            )}

            {/* Tabs Design */}
            <div className="flex p-1 bg-slate-900/50 border border-slate-800 rounded-2xl mb-8 max-w-md">
                <button
                    onClick={() => setActiveTab('global')}
                    className={`flex-1 flex items-center justify-center gap-2 py-3 px-6 rounded-xl font-bold transition-all ${activeTab === 'global' ? 'bg-indigo-500 text-white shadow-lg shadow-indigo-500/20' : 'text-slate-500 hover:text-slate-300'
                        }`}
                >
                    <Layout className="w-4 h-4" /> Global Standards
                </button>
                <button
                    onClick={() => setActiveTab('overrides')}
                    className={`flex-1 flex items-center justify-center gap-2 py-3 px-6 rounded-xl font-bold transition-all ${activeTab === 'overrides' ? 'bg-indigo-500 text-white shadow-lg shadow-indigo-500/20' : 'text-slate-500 hover:text-slate-300'
                        }`}
                >
                    <Layers className="w-4 h-4" /> Repo Exceptions
                </button>
            </div>

            {activeTab === 'global' ? (
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 animate-in slide-in-from-left-4 duration-500">
                    {/* Main Controls */}
                    <div className="lg:col-span-8 space-y-8">
                        <div className="bg-[#0f111a]/80 backdrop-blur-xl border border-white/5 rounded-3xl overflow-hidden shadow-2xl">
                            <div className="p-8 border-b border-white/5 bg-gradient-to-r from-indigo-500/10 to-transparent">
                                <div className="flex items-center gap-4">
                                    <div className="p-3 bg-indigo-500/20 rounded-2xl text-indigo-400">
                                        <ShieldCheck className="w-8 h-8" />
                                    </div>
                                    <div>
                                        <h2 className="text-xl font-bold text-white tracking-tight">Standard Enforcement</h2>
                                        <p className="text-slate-500 text-sm">Rules applied to all repositories without exceptions.</p>
                                    </div>
                                </div>
                            </div>

                            <div className="p-8 space-y-10">
                                {/* Block Unscanned Toggle */}
                                <div className="group relative bg-white/5 border border-white/5 p-6 rounded-2xl hover:border-indigo-500/30 transition-all">
                                    <div className="flex items-center justify-between">
                                        <div className="space-y-1">
                                            <h3 className="text-white font-bold flex items-center gap-2 uppercase tracking-wide text-xs">
                                                Global Quarantine
                                            </h3>
                                            <p className="text-slate-400 text-sm">Block Pulls for Images without recent scan data</p>
                                        </div>
                                        <label className="relative inline-flex items-center cursor-pointer scale-110">
                                            <input
                                                type="checkbox"
                                                className="sr-only peer"
                                                checked={policy?.block_unscanned}
                                                onChange={(e) => policy && setPolicy({ ...policy, block_unscanned: e.target.checked })}
                                            />
                                            <div className="w-14 h-7 bg-slate-800 border border-slate-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[4px] after:left-[4px] after:bg-slate-500 after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-indigo-600 after:peer-checked:bg-white shadow-inner"></div>
                                        </label>
                                    </div>
                                </div>

                                {/* Severity Sliders */}
                                <div className="space-y-12">
                                    <h3 className="text-slate-400 font-bold uppercase tracking-[0.2em] text-[10px]">Severity Thresholds</h3>

                                    {[
                                        { key: 'critical_threshold', label: 'Critical', color: 'rose', icon: <AlertCircle />, max: 100 },
                                        { key: 'high_threshold', label: 'High', color: 'orange', icon: <AlertTriangle />, max: 250 },
                                        { key: 'medium_threshold', label: 'Medium', color: 'yellow', icon: <Zap />, max: 500 },
                                        { key: 'low_threshold', label: 'Low', color: 'slate', icon: <Info />, max: 1000 }
                                    ].map((s) => (
                                        <div key={s.key} className="relative group">
                                            <div className="flex items-center justify-between mb-4">
                                                <label className={`flex items-center gap-3 font-bold text-${s.color}-400 uppercase tracking-widest text-xs`}>
                                                    <span className={`p-2 bg-${s.color}-500/10 rounded-lg group-hover:scale-110 transition-transform`}>{s.icon}</span>
                                                    {s.label} Severities
                                                </label>
                                                <div className="flex items-center gap-2">
                                                    <span className="text-slate-500 text-xs font-mono">Limit</span>
                                                    <input
                                                        type="number"
                                                        value={policy ? (policy as any)[s.key] : 0}
                                                        onChange={(e) => policy && setPolicy({ ...policy, [s.key]: parseInt(e.target.value) || 0 })}
                                                        className={`w-16 bg-white/5 border border-white/5 text-center text-${s.color}-400 font-black rounded-lg py-1 px-1 focus:outline-none focus:border-${s.color}-500/50`}
                                                    />
                                                </div>
                                            </div>
                                            <input
                                                type="range" min="0" max={s.max}
                                                value={policy ? (policy as any)[s.key] : 0}
                                                onChange={(e) => policy && setPolicy({ ...policy, [s.key]: parseInt(e.target.value) })}
                                                className={`w-full h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-${s.color}-500`}
                                            />
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Info Panel */}
                    <div className="lg:col-span-4 space-y-6">
                        <div className="bg-indigo-600/10 border border-indigo-500/20 p-8 rounded-3xl relative overflow-hidden group">
                            <div className="absolute top-0 right-0 p-4 opacity-10 group-hover:scale-125 transition-transform duration-700">
                                <ShieldCheck size={120} />
                            </div>
                            <h3 className="text-indigo-400 font-bold mb-4 flex items-center gap-2">
                                <Zap size={18} /> OPA Enforcer
                            </h3>
                            <p className="text-slate-400 text-sm leading-relaxed mb-6">
                                Policies are dynamically compiled into <strong>Rego</strong> and enforced at the
                                proxy layer. This prevents un-trusted images from ever touching your cluster.
                            </p>
                            <div className="space-y-3">
                                <div className="flex items-center gap-3 p-3 bg-white/5 rounded-xl border border-white/5">
                                    <div className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                                    <span className="text-xs font-bold text-slate-300 tracking-wide uppercase">Real-time Evaluation</span>
                                </div>
                                <div className="flex items-center gap-3 p-3 bg-white/5 rounded-xl border border-white/5">
                                    <div className="w-2 h-2 rounded-full bg-indigo-500" />
                                    <span className="text-xs font-bold text-slate-300 tracking-wide uppercase">DB Persisted Engine</span>
                                </div>
                            </div>
                        </div>

                        <div className="bg-[#0f111a]/50 border border-white/5 p-8 rounded-3xl">
                            <h4 className="text-white font-bold text-sm mb-4">Security Best Practices</h4>
                            <ul className="space-y-4 text-xs text-slate-500">
                                <li className="flex gap-2">
                                    <ArrowRight className="w-3 h-3 text-indigo-500 shrink-0" />
                                    <span>Set Critical Threshold to <strong>0</strong> for Production environments.</span>
                                </li>
                                <li className="flex gap-2">
                                    <ArrowRight className="w-3 h-3 text-indigo-500 shrink-0" />
                                    <span>Enable "Block Unscanned" to prevent zero-day vulnerability bypass.</span>
                                </li>
                                <li className="flex gap-2">
                                    <ArrowRight className="w-3 h-3 text-indigo-500 shrink-0" />
                                    <span>Use exceptions for legacy systems only after risk assessment.</span>
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
            ) : (
                <div className="space-y-8 animate-in slide-in-from-right-4 duration-500">
                    <div className="flex items-center justify-between">
                        <h2 className="text-2xl font-black text-white tracking-tight flex items-center gap-3">
                            <Layers className="w-6 h-6 text-indigo-400" /> Repository Exceptions
                        </h2>
                        <button
                            onClick={() => setShowAddOverride(true)}
                            className="flex items-center gap-2 px-6 py-2.5 bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-400 border border-indigo-500/30 rounded-xl font-bold uppercase tracking-widest text-xs transition-all active:scale-95"
                        >
                            <Plus className="w-4 h-4" /> Add Override
                        </button>
                    </div>

                    {showAddOverride && (
                        <div className="bg-[#0f111a] border border-indigo-500/30 rounded-3xl p-8 relative overflow-hidden animate-in zoom-in-95 duration-300">
                            <div className="absolute top-0 right-0 p-8 opacity-5">
                                <Plus size={150} />
                            </div>
                            <h3 className="text-white font-bold mb-8 uppercase tracking-widest text-sm flex items-center gap-2">
                                Configure New Exception
                            </h3>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-8 relative z-10">
                                <div className="space-y-6">
                                    <div className="space-y-2">
                                        <label className="text-slate-500 font-bold uppercase tracking-widest text-[10px]">Target Repository Path</label>
                                        <div className="relative">
                                            <Search className="absolute left-4 top-1/2 -translate-y-1/2 text-slate-500 w-4 h-4" />
                                            <input
                                                type="text"
                                                placeholder="e.g. production/orders-api"
                                                value={newRepoPath}
                                                onChange={(e) => setNewRepoPath(e.target.value)}
                                                className="w-full bg-white/5 border border-white/10 rounded-xl py-4 pl-12 pr-4 text-white font-medium focus:outline-none focus:border-indigo-500 transition-colors"
                                            />
                                        </div>
                                    </div>

                                    <div className="p-4 bg-amber-500/5 border border-amber-500/20 rounded-2xl flex gap-3">
                                        <AlertTriangle className="w-5 h-5 text-amber-500 shrink-0" />
                                        <p className="text-xs text-amber-200/60 leading-relaxed">
                                            Overrides completely replace global settings for this path. Use with caution for restricted workloads.
                                        </p>
                                    </div>
                                </div>

                                <div className="grid grid-cols-2 gap-4">
                                    {[
                                        { key: 'critical_threshold', label: 'Critical', color: 'rose' },
                                        { key: 'high_threshold', label: 'High', color: 'orange' },
                                        { key: 'medium_threshold', label: 'Medium', color: 'yellow' },
                                        { key: 'low_threshold', label: 'Low', color: 'slate' }
                                    ].map(s => (
                                        <div key={s.key} className="space-y-2">
                                            <label className="text-slate-500 font-bold uppercase tracking-widest text-[9px]">{s.label} Limit</label>
                                            <input
                                                type="number"
                                                value={(newRepoPolicy as any)[s.key]}
                                                onChange={(e) => setNewRepoPolicy({ ...newRepoPolicy, [s.key]: parseInt(e.target.value) || 0 })}
                                                className={`w-full bg-white/5 border border-white/10 rounded-xl py-3 px-4 text-white font-extrabold text-center focus:outline-none focus:border-${s.color}-500/50`}
                                            />
                                        </div>
                                    ))}
                                    <div className="col-span-2 flex items-center justify-between bg-white/5 p-3 rounded-xl border border-white/5 mt-2">
                                        <span className="text-xs font-bold text-slate-400 uppercase tracking-widest">Allow Unscanned?</span>
                                        <label className="relative inline-flex items-center cursor-pointer">
                                            <input
                                                type="checkbox"
                                                className="sr-only peer"
                                                checked={!newRepoPolicy.block_unscanned}
                                                onChange={(e) => setNewRepoPolicy({ ...newRepoPolicy, block_unscanned: !e.target.checked })}
                                            />
                                            <div className="w-10 h-5 bg-slate-800 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-slate-500 after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-emerald-600 after:peer-checked:bg-white"></div>
                                        </label>
                                    </div>
                                </div>
                            </div>

                            <div className="mt-10 flex gap-4">
                                <button
                                    onClick={handleAddOverride}
                                    className="px-10 py-4 bg-indigo-600 hover:bg-indigo-500 text-white font-black uppercase tracking-widest text-xs rounded-xl transition-all shadow-xl shadow-indigo-900/40 active:scale-95"
                                >
                                    Confirm Exception
                                </button>
                                <button
                                    onClick={() => setShowAddOverride(false)}
                                    className="px-10 py-4 bg-white/5 hover:bg-white/10 text-slate-400 font-black uppercase tracking-widest text-xs rounded-xl transition-all"
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    )}

                    <div className="bg-[#0f111a]/80 backdrop-blur-xl border border-white/5 rounded-3xl overflow-hidden shadow-2xl">
                        <table className="w-full text-left">
                            <thead>
                                <tr className="bg-white/5 border-b border-white/5">
                                    <th className="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-[0.2em]">Repository Path</th>
                                    <th className="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] text-center">Threat Matrix</th>
                                    <th className="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] text-center">Enforcement</th>
                                    <th className="px-8 py-5 text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] text-right">Actions</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-white/5">
                                {overrides.length === 0 ? (
                                    <tr>
                                        <td colSpan={4} className="px-8 py-20 text-center">
                                            <div className="flex flex-col items-center gap-4">
                                                <div className="p-4 bg-white/5 rounded-full text-slate-700">
                                                    <Layers size={48} />
                                                </div>
                                                <p className="text-slate-500 font-medium">No overrides defined. All repos following global standards.</p>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    overrides.map((o) => (
                                        <tr key={o.id} className="group hover:bg-white/[0.02] transition-colors">
                                            <td className="px-8 py-6">
                                                <div className="flex items-center gap-3">
                                                    <div className="p-2 bg-indigo-500/10 rounded-lg text-indigo-400">
                                                        <Layout size={16} />
                                                    </div>
                                                    <span className="text-white font-bold tracking-tight">{o.repository_path}</span>
                                                </div>
                                            </td>
                                            <td className="px-8 py-6">
                                                <div className="flex items-center justify-center gap-2">
                                                    <span className="w-8 h-8 flex items-center justify-center rounded-lg bg-rose-500/20 text-rose-400 text-[10px] font-black border border-rose-500/20" title="Critical">{o.config.critical_threshold}</span>
                                                    <span className="w-8 h-8 flex items-center justify-center rounded-lg bg-orange-500/20 text-orange-400 text-[10px] font-black border border-orange-500/20" title="High">{o.config.high_threshold}</span>
                                                    <span className="w-8 h-8 flex items-center justify-center rounded-lg bg-yellow-500/20 text-yellow-400 text-[10px] font-black border border-yellow-500/20" title="Medium">{o.config.medium_threshold}</span>
                                                    <span className="w-8 h-8 flex items-center justify-center rounded-lg bg-slate-700/50 text-slate-400 text-[10px] font-black border border-slate-700/50" title="Low">{o.config.low_threshold}</span>
                                                </div>
                                            </td>
                                            <td className="px-8 py-6 text-center">
                                                {o.config.block_unscanned ? (
                                                    <span className="px-3 py-1 bg-emerald-500/10 text-emerald-400 text-[10px] font-black uppercase rounded-full border border-emerald-500/20">Strict</span>
                                                ) : (
                                                    <span className="px-3 py-1 bg-amber-500/10 text-amber-400 text-[10px] font-black uppercase rounded-full border border-amber-500/20">Permissive</span>
                                                )}
                                            </td>
                                            <td className="px-8 py-6">
                                                <div className="flex items-center justify-end gap-2">
                                                    <button className="p-2.5 text-slate-600 hover:text-white transition-colors bg-white/0 hover:bg-white/5 rounded-xl opacity-0 group-hover:opacity-100">
                                                        <Edit2 size={16} />
                                                    </button>
                                                    <button
                                                        onClick={() => handleDeleteOverride(o.repository_path)}
                                                        className="p-2.5 text-slate-600 hover:text-rose-500 transition-colors bg-white/0 hover:bg-rose-500/10 rounded-xl opacity-0 group-hover:opacity-100"
                                                    >
                                                        <Trash2 size={16} />
                                                    </button>
                                                </div>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            )}
        </div>
    );
};

export default SecurityPolicies;

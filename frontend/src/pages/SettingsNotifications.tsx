import React, { useState, useEffect } from 'react';
import { Bell, Plus, Trash2, CheckCircle, AlertTriangle, Zap, Slack, MessageSquare } from 'lucide-react';
import api from '../lib/api';

interface Webhook {
    id: string;
    url: string;
    type: string;
    events: string[];
    enabled: boolean;
    created_at: string;
}

const SettingsNotifications = () => {
    const [webhooks, setWebhooks] = useState<Webhook[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showAdd, setShowAdd] = useState(false);

    // Form State
    const [newUrl, setNewUrl] = useState('');
    const [newType, setNewType] = useState('slack');
    const [newEvents, setNewEvents] = useState<string[]>(['SCAN_COMPLETED', 'CRITICAL_VULN_FOUND']);

    useEffect(() => {
        loadWebhooks();
    }, []);

    const loadWebhooks = async () => {
        try {
            setLoading(true);
            const res = await api.listWebhooks();
            setWebhooks(res.data.data || []);
        } catch (err) {
            console.error(err);
            setError("Failed to load webhooks");
        } finally {
            setLoading(false);
        }
    };

    const handleAdd = async () => {
        try {
            await api.createWebhook(newUrl, newType, newEvents);
            setNewUrl('');
            setShowAdd(false);
            loadWebhooks();
        } catch (err) {
            console.error(err);
            setError("Failed to create webhook");
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure?")) return;
        try {
            await api.deleteWebhook(id);
            loadWebhooks();
        } catch (err) {
            setError("Failed to delete webhook");
        }
    };

    const handleTest = async (id: string) => {
        try {
            await api.testWebhook(id);
            alert("Test notification sent!");
        } catch (err) {
            alert("Failed to send test notification");
        }
    };

    const toggleEvent = (event: string) => {
        if (newEvents.includes(event)) {
            setNewEvents(newEvents.filter(e => e !== event));
        } else {
            setNewEvents([...newEvents, event]);
        }
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-xl font-bold text-white flex items-center gap-2">
                        <Bell className="text-blue-400" /> Notification Channels
                    </h2>
                    <p className="text-gray-400 text-sm mt-1">
                        Configure where you want to receive RegistryX alerts.
                    </p>
                </div>
                <button
                    onClick={() => setShowAdd(!showAdd)}
                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-500 rounded-lg text-white font-bold text-xs uppercase tracking-wider transition-colors"
                >
                    <Plus size={16} /> Add Channel
                </button>
            </div>

            {error && (
                <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl flex items-center gap-2">
                    <AlertTriangle size={18} /> {error}
                </div>
            )}

            {showAdd && (
                <div className="bg-white/5 border border-white/10 rounded-xl p-6 space-y-4 animate-in fade-in slide-in-from-top-4">
                    <h3 className="font-bold text-white">New Webhook</h3>

                    <div>
                        <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Webhook URL / Email</label>
                        <input
                            type="text"
                            value={newUrl}
                            onChange={e => setNewUrl(e.target.value)}
                            placeholder={newType === 'email' ? 'admin@example.com' : 'https://hooks.slack.com/services/...'}
                            className="w-full bg-black/40 border border-white/10 rounded-lg p-3 text-white text-sm focus:border-blue-500/50 outline-none"
                        />
                    </div>

                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Platform</label>
                            <select
                                value={newType}
                                onChange={e => setNewType(e.target.value)}
                                className="w-full bg-black/40 border border-white/10 rounded-lg p-3 text-white text-sm outline-none"
                            >
                                <option value="slack">Slack</option>
                                <option value="discord">Discord</option>
                                <option value="email">Email (SMTP)</option>
                                <option value="generic">Generic JSON</option>
                            </select>
                        </div>
                        <div>
                            <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Triggers</label>
                            <div className="space-y-2">
                                {['SCAN_COMPLETED', 'CRITICAL_VULN_FOUND', 'POLICY_VIOLATION'].map(evt => (
                                    <label key={evt} className="flex items-center gap-2 cursor-pointer">
                                        <input
                                            type="checkbox"
                                            checked={newEvents.includes(evt)}
                                            onChange={() => toggleEvent(evt)}
                                            className="rounded border-gray-600 bg-gray-800 text-blue-500"
                                        />
                                        <span className="text-sm text-gray-300">{evt}</span>
                                    </label>
                                ))}
                            </div>
                        </div>
                    </div>

                    <div className="flex justify-end gap-3 pt-4">
                        <button
                            onClick={() => setShowAdd(false)}
                            className="px-4 py-2 hover:bg-white/5 rounded-lg text-gray-400 text-sm font-bold"
                        >
                            Cancel
                        </button>
                        <button
                            onClick={handleAdd}
                            disabled={!newUrl}
                            className="px-6 py-2 bg-emerald-600 hover:bg-emerald-500 disabled:opacity-50 disabled:cursor-not-allowed rounded-lg text-white font-bold text-sm shadow-lg shadow-emerald-900/20"
                        >
                            Save Channel
                        </button>
                    </div>
                </div>
            )}

            <div className="grid gap-4">
                {webhooks.length === 0 && !loading ? (
                    <div className="text-center py-12 text-gray-500 bg-white/5 rounded-xl border border-white/5 border-dashed">
                        No notifications configured.
                    </div>
                ) : (
                    webhooks.map(hook => (
                        <div key={hook.id} className="bg-white/5 border border-white/10 rounded-xl p-5 flex items-center justify-between group hover:border-blue-500/30 transition-all">
                            <div className="flex items-center gap-4">
                                <div className={`w-12 h-12 rounded-full flex items-center justify-center ${hook.type === 'slack' ? 'bg-[#4A154B]/20 text-[#E01E5A]' :
                                    hook.type === 'discord' ? 'bg-[#5865F2]/20 text-[#5865F2]' : 'bg-gray-700/20 text-gray-400'
                                    }`}>
                                    {hook.type === 'slack' ? <Slack size={24} /> :
                                        hook.type === 'discord' ? <MessageSquare size={24} /> : <Zap size={24} />}
                                </div>
                                <div>
                                    <div className="font-bold text-white flex items-center gap-2">
                                        <span className="uppercase">{hook.type}</span>
                                        <span className="text-[10px] bg-white/10 px-2 py-0.5 rounded text-gray-400 truncate max-w-[200px]">{hook.url}</span>
                                    </div>
                                    <div className="flex gap-2 mt-1">
                                        {hook.events.map(e => (
                                            <span key={e} className="text-[10px] font-mono text-emerald-400 bg-emerald-500/10 px-1.5 rounded">
                                                {e}
                                            </span>
                                        ))}
                                    </div>
                                </div>
                            </div>

                            <div className="flex items-center gap-3 opacity-50 group-hover:opacity-100 transition-opacity">
                                <button
                                    onClick={() => handleTest(hook.id)}
                                    className="p-2 hover:bg-white/10 rounded-lg text-blue-400"
                                    title="Send Test Notification"
                                >
                                    <Zap size={18} />
                                </button>
                                <button
                                    onClick={() => handleDelete(hook.id)}
                                    className="p-2 hover:bg-red-500/10 rounded-lg text-red-400"
                                    title="Delete Channel"
                                >
                                    <Trash2 size={18} />
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default SettingsNotifications;

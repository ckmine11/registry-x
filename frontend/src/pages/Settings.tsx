import React, { useState } from 'react';
import { Webhook, Key, Plus, Trash2, Check, X, AlertTriangle, User, Bell, Shield } from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api, ServiceAccount } from '../lib/api';
import SettingsNotifications from './SettingsNotifications';
import SettingsUsers from './SettingsUsers';
import LicenseSettings from './LicenseSettings';


export default function Settings() {
    const [activeTab, setActiveTab] = useState<'general' | 'notifications' | 'users' | 'license'>('license');

    const [newAccountName, setNewAccountName] = useState('');
    const [showNewKey, setShowNewKey] = useState<string | null>(null);

    const queryClient = useQueryClient();

    // Fetch Accounts
    const { data: accountsData, isLoading } = useQuery({
        queryKey: ['service-accounts'],
        queryFn: api.getServiceAccounts,
    });
    const accounts = accountsData?.data?.data || [];

    // Create Account Mutation
    const createMutation = useMutation({
        mutationFn: (name: string) => api.createServiceAccount(name, 'Generated via UI'),
        onSuccess: (data) => {
            queryClient.invalidateQueries({ queryKey: ['service-accounts'] });
            setShowNewKey(data.data.apiKey);
            setNewAccountName('');
        }
    });

    // Revoke Account Mutation
    const revokeMutation = useMutation({
        mutationFn: api.revokeServiceAccount,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['service-accounts'] });
        }
    });

    const handleCreate = () => {
        if (!newAccountName) return;
        createMutation.mutate(newAccountName);
    };

    return (
        <div className="p-6 space-y-8">
            <h1 className="text-3xl font-bold tracking-tight text-white">Settings</h1>

            {/* Tabs */}
            <div className="flex space-x-4 border-b border-gray-700 pb-1">
                <button
                    onClick={() => setActiveTab('general')}
                    className={`pb-3 px-2 flex items-center gap-2 font-medium text-sm transition-colors relative ${activeTab === 'general' ? 'text-blue-400' : 'text-gray-400 hover:text-white'}`}
                >
                    <Key size={16} /> Service Accounts
                    {activeTab === 'general' && <div className="absolute bottom-[-5px] left-0 w-full h-1 bg-blue-500 rounded-t-full" />}
                </button>
                <button
                    onClick={() => setActiveTab('notifications')}
                    className={`pb-3 px-2 flex items-center gap-2 font-medium text-sm transition-colors relative ${activeTab === 'notifications' ? 'text-blue-400' : 'text-gray-400 hover:text-white'}`}
                >
                    <Bell size={16} /> Notifications
                    {activeTab === 'notifications' && <div className="absolute bottom-[-5px] left-0 w-full h-1 bg-blue-500 rounded-t-full" />}
                </button>
                <button
                    onClick={() => setActiveTab('users')}
                    className={`pb-3 px-2 flex items-center gap-2 font-medium text-sm transition-colors relative ${activeTab === 'users' ? 'text-purple-400' : 'text-gray-400 hover:text-white'}`}
                >
                    <User size={16} /> Team & Users
                    {activeTab === 'users' && <div className="absolute bottom-[-5px] left-0 w-full h-1 bg-purple-500 rounded-t-full" />}
                </button>
                <button
                    onClick={() => setActiveTab('license')}
                    className={`pb-3 px-2 flex items-center gap-2 font-medium text-sm transition-colors relative ${activeTab === 'license' ? 'text-blue-400' : 'text-gray-400 hover:text-white'}`}
                >
                    <Shield size={16} /> License & Plan
                    {activeTab === 'license' && <div className="absolute bottom-[-5px] left-0 w-full h-1 bg-blue-500 rounded-t-full" />}
                </button>

            </div>

            {/* Content */}
            {activeTab === 'general' && (
                <section className="bg-gray-800 rounded-xl border border-gray-700 p-6 animate-in fade-in slide-in-from-left-4">
                    <div className="flex items-center justify-between mb-6">
                        <h2 className="text-xl font-semibold text-white flex items-center gap-2">
                            <Key className="w-5 h-5 text-yellow-400" /> Service Accounts
                        </h2>
                    </div>

                    <div className="mb-6 bg-gray-900/50 p-4 rounded-lg border border-gray-700">
                        <h3 className="text-sm font-medium text-gray-300 mb-3">Create New Service Account</h3>
                        <div className="flex gap-4">
                            <input
                                type="text"
                                placeholder="Account Name (e.g. ci-bot)"
                                className="flex-1 bg-gray-800 border border-gray-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                                value={newAccountName}
                                onChange={(e) => setNewAccountName(e.target.value)}
                            />
                            <button
                                onClick={handleCreate}
                                disabled={createMutation.isPending || !newAccountName}
                                className="bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 flex items-center gap-2"
                            >
                                {createMutation.isPending ? 'Creating...' : <><Plus size={16} /> Create</>}
                            </button>
                        </div>

                        {showNewKey && (
                            <div className="mt-4 p-4 bg-yellow-900/20 border border-yellow-500/30 rounded-lg flex items-start gap-3">
                                <AlertTriangle className="text-yellow-500 shrink-0" size={20} />
                                <div className="flex-1">
                                    <p className="text-yellow-200 font-medium mb-1">New API Key Generated</p>
                                    <p className="text-sm text-yellow-200/70 mb-2">Copy this key now. You won't be able to see it again.</p>
                                    <div className="flex items-center gap-2 bg-black/40 p-2 rounded border border-yellow-500/20 font-mono text-yellow-400 break-all select-all">
                                        {showNewKey}
                                    </div>
                                    <button
                                        onClick={() => setShowNewKey(null)}
                                        className="mt-3 text-sm text-yellow-500 hover:text-yellow-400 underline"
                                    >
                                        I have copied it
                                    </button>
                                </div>
                            </div>
                        )}
                    </div>

                    <div className="space-y-3">
                        {isLoading ? <div className="text-gray-400">Loading accounts...</div> : accounts.length === 0 ? <div className="text-gray-400">No service accounts found.</div> :
                            accounts.map((acc: ServiceAccount) => (
                                <div key={acc.id} className={`bg-black/20 rounded-lg p-4 flex items-center justify-between border border-white/5 ${acc.status === 'revoked' ? 'opacity-50' : ''}`}>
                                    <div>
                                        <div className="text-white font-medium flex items-center gap-2">
                                            {acc.name}
                                            {acc.status === 'revoked' && <span className="text-xs bg-red-500/20 text-red-500 px-2 py-0.5 rounded">Revoked</span>}
                                            {acc.status === 'active' && <span className="text-xs bg-green-500/20 text-green-500 px-2 py-0.5 rounded">Active</span>}
                                        </div>
                                        <div className="text-xs text-gray-500 mt-1">
                                            ID: {acc.id} • Created {new Date(acc.created).toLocaleDateString()}
                                        </div>
                                    </div>
                                    {acc.status === 'active' && (
                                        <button
                                            onClick={() => {
                                                if (confirm('Are you sure you want to revoke this key?')) {
                                                    revokeMutation.mutate(acc.id);
                                                }
                                            }}
                                            className="text-red-400 hover:text-red-300 p-2 hover:bg-red-500/10 rounded-lg transition-colors"
                                            title="Revoke Access"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    )}
                                </div>
                            ))}
                    </div>
                </section>
            )}

            {activeTab === 'notifications' && (
                <section className="bg-gray-800 rounded-xl border border-gray-700 p-6 animate-in fade-in slide-in-from-right-4">
                    <SettingsNotifications />
                </section>
            )}

            {activeTab === 'users' && (
                <section className="bg-gray-800 rounded-xl border border-gray-700 p-6 animate-in fade-in slide-in-from-right-4">
                    <SettingsUsers />
                </section>
            )}

            {activeTab === 'license' && (
                <section className="bg-black/20 rounded-xl border border-white/5 p-6 animate-in fade-in slide-in-from-right-4">
                    <LicenseSettings />
                </section>
            )}

        </div>
    );
}

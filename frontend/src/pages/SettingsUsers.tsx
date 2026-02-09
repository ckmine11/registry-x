import React, { useState, useEffect } from 'react';
import { User, Plus, Trash2, Shield, Mail, Lock } from 'lucide-react';
import api from '../lib/api';

interface UserType {
    id: string;
    username: string;
    email: string;
    role: string;
    created_at: string;
}

const SettingsUsers = () => {
    const [users, setUsers] = useState<UserType[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [showInvite, setShowInvite] = useState(false);
    const [inviteResult, setInviteResult] = useState<{ user: UserType, temporaryPassword: string } | null>(null);

    // Form State
    const [email, setEmail] = useState('');
    const [role, setRole] = useState('developer');

    useEffect(() => {
        loadUsers();
    }, []);

    const loadUsers = async () => {
        try {
            setLoading(true);
            const res = await api.listUsers();
            setUsers(res.data.data || []);
        } catch (err) {
            console.error(err);
            setError("Failed to load users");
        } finally {
            setLoading(false);
        }
    };

    const handleInvite = async () => {
        try {
            setError(null);
            const res = await api.inviteUser(email, role);
            setInviteResult(res.data);
            setEmail('');
            loadUsers();
        } catch (err: any) {
            console.error(err);
            setError(err.response?.data || "Failed to invite user");
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm("Are you sure? This will delete the user and their personal namespace.")) return;
        try {
            await api.deleteUser(id);
            loadUsers();
        } catch (err) {
            setError("Failed to delete user");
        }
    };

    const handleRoleChange = async (id: string, newRole: string) => {
        try {
            await api.updateUserRole(id, newRole);
            loadUsers();
        } catch (err) {
            setError("Failed to update role");
        }
    };

    const closeInviteModal = () => {
        setShowInvite(false);
        setInviteResult(null);
    };

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h2 className="text-xl font-bold text-white flex items-center gap-2">
                        <User className="text-purple-400" /> Team Management
                    </h2>
                    <p className="text-gray-400 text-sm mt-1">
                        Manage access and roles for your registry.
                    </p>
                </div>
                <button
                    onClick={() => setShowInvite(true)}
                    className="flex items-center gap-2 px-4 py-2 bg-purple-600 hover:bg-purple-500 rounded-lg text-white font-bold text-xs uppercase tracking-wider transition-colors"
                >
                    <Plus size={16} /> Invite User
                </button>
            </div>

            {error && (
                <div className="bg-red-500/10 border border-red-500/20 text-red-400 p-4 rounded-xl">
                    {error}
                </div>
            )}

            {/* Invite Modal */}
            {showInvite && (
                <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-2xl w-full max-w-md p-6 space-y-4">
                        <h3 className="text-lg font-bold text-white flex items-center gap-2">
                            <Mail size={18} /> Invite New Member
                        </h3>

                        {!inviteResult ? (
                            <>
                                <div>
                                    <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Email Address</label>
                                    <input
                                        type="email"
                                        value={email}
                                        onChange={e => setEmail(e.target.value)}
                                        className="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-white text-sm focus:border-purple-500/50 outline-none"
                                        placeholder="colleague@example.com"
                                    />
                                </div>
                                <div>
                                    <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Role</label>
                                    <select
                                        value={role}
                                        onChange={e => setRole(e.target.value)}
                                        className="w-full bg-black/40 border border-gray-700 rounded-lg p-3 text-white text-sm outline-none"
                                    >
                                        <option value="developer">Developer (Push/Pull)</option>
                                        <option value="admin">Admin (Full Access)</option>
                                        <option value="readonly">Read-Only (Pull Only)</option>
                                    </select>
                                </div>
                                <div className="flex justify-end gap-3 pt-4">
                                    <button
                                        onClick={closeInviteModal}
                                        className="px-4 py-2 hover:bg-white/5 rounded-lg text-gray-400 text-sm font-bold"
                                    >
                                        Cancel
                                    </button>
                                    <button
                                        onClick={handleInvite}
                                        disabled={!email}
                                        className="px-6 py-2 bg-purple-600 hover:bg-purple-500 disabled:opacity-50 rounded-lg text-white font-bold text-sm"
                                    >
                                        Send Invite
                                    </button>
                                </div>
                            </>
                        ) : (
                            <div className="space-y-4">
                                <div className="bg-green-500/10 border border-green-500/20 p-4 rounded-xl text-green-400 text-sm">
                                    User created successfully! An invitation email has been sent to the user. Credentials are also shown below for your records:
                                </div>
                                <div className="bg-black/50 p-4 rounded-xl space-y-2 border border-white/5 font-mono text-sm">
                                    <div className="flex justify-between">
                                        <span className="text-gray-500">Username:</span>
                                        <span className="text-white select-all">{inviteResult.user.username}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-gray-500">Temp Password:</span>
                                        <span className="text-yellow-400 select-all">{inviteResult.temporaryPassword}</span>
                                    </div>
                                </div>
                                <button
                                    onClick={closeInviteModal}
                                    className="w-full py-2 bg-gray-800 hover:bg-gray-700 rounded-lg text-white font-bold text-sm"
                                >
                                    Done
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            )}

            <div className="bg-gray-800/50 border border-gray-700 rounded-xl overflow-hidden">
                <table className="w-full text-left border-collapse">
                    <thead>
                        <tr className="bg-white/5 border-b border-white/5 text-gray-400 text-xs uppercase tracking-wider">
                            <th className="p-4 font-bold">User</th>
                            <th className="p-4 font-bold">Role</th>
                            <th className="p-4 font-bold">Joined</th>
                            <th className="p-4 font-bold text-right">Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-white/5">
                        {users.map(user => (
                            <tr key={user.id} className="hover:bg-white/5 transition-colors">
                                <td className="p-4">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded-full bg-gradient-to-br from-purple-500 to-indigo-600 flex items-center justify-center text-white font-bold text-xs">
                                            {user.username.slice(0, 2).toUpperCase()}
                                        </div>
                                        <div>
                                            <div className="text-white font-medium">{user.username}</div>
                                            <div className="text-gray-500 text-xs">{user.email}</div>
                                        </div>
                                    </div>
                                </td>
                                <td className="p-4">
                                    <select
                                        value={user.role}
                                        onChange={(e) => handleRoleChange(user.id, e.target.value)}
                                        className="bg-black/20 border border-white/10 rounded px-2 py-1 text-xs text-white outline-none focus:border-purple-500/50"
                                    >
                                        <option value="admin" className="bg-gray-900 text-white">Admin</option>
                                        <option value="developer" className="bg-gray-900 text-white">Developer</option>
                                        <option value="readonly" className="bg-gray-900 text-white">Read-Only</option>
                                    </select>
                                </td>
                                <td className="p-4 text-sm text-gray-400">
                                    {new Date(user.created_at).toLocaleDateString()}
                                </td>
                                <td className="p-4 text-right">
                                    <button
                                        onClick={() => handleDelete(user.id)}
                                        className="text-gray-600 hover:text-red-400 transition-colors p-2"
                                        title="Delete User"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
                {!loading && users.length === 0 && (
                    <div className="p-8 text-center text-gray-500">No users found.</div>
                )}
            </div>
        </div>
    );
};

export default SettingsUsers;

import React, { useState } from 'react';
import { useAuth } from '../lib/auth-context';
import { api } from '../lib/api';
import { User, Lock, CheckCircle, AlertTriangle } from 'lucide-react';

const UserProfile = () => {
    const { user } = useAuth();
    // Fallback if not loaded yet (though protected route handles this)
    if (!user) return null;

    const [newPass, setNewPass] = useState('');
    const [confirmPass, setConfirmPass] = useState('');
    const [status, setStatus] = useState<{ type: 'success' | 'error' | '', msg: string }>({ type: '', msg: '' });
    const [loading, setLoading] = useState(false);

    const handlePasswordReset = async (e: React.FormEvent) => {
        e.preventDefault();
        setStatus({ type: '', msg: '' });

        if (newPass !== confirmPass) {
            setStatus({ type: 'error', msg: 'Passwords do not match' });
            return;
        }

        if (newPass.length < 6) {
            setStatus({ type: 'error', msg: 'Password must be at least 6 characters' });
            return;
        }

        setLoading(true);
        try {
            await api.changePassword(newPass);
            setStatus({ type: 'success', msg: 'Password updated successfully' });
            setNewPass('');
            setConfirmPass('');
        } catch (err: any) {
            setStatus({ type: 'error', msg: err.response?.data || err.message || 'Failed to update password' });
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto space-y-6">
            <h1 className="text-2xl font-bold text-white mb-6">User Profile</h1>

            {/* Profile Info Card */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <div className="flex items-center space-x-4 mb-6">
                    <div className="w-16 h-16 bg-blue-600 rounded-full flex items-center justify-center">
                        <User size={32} className="text-white" />
                    </div>
                    <div>
                        <h2 className="text-xl font-semibold text-white">{user.username}</h2>
                        <p className="text-gray-400">{user.email}</p>
                    </div>
                </div>
            </div>

            {/* Password Reset Card */}
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <div className="flex items-center mb-4">
                    <Lock className="text-purple-400 mr-2" size={20} />
                    <h3 className="text-lg font-semibold text-white">Change Password</h3>
                </div>

                <form onSubmit={handlePasswordReset} className="space-y-4">
                    {status.msg && (
                        <div className={`p-3 rounded-lg flex items-center ${status.type === 'success' ? 'bg-green-500/10 text-green-400' : 'bg-red-500/10 text-red-400'}`}>
                            {status.type === 'success' ? <CheckCircle size={18} className="mr-2" /> : <AlertTriangle size={18} className="mr-2" />}
                            {status.msg}
                        </div>
                    )}

                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-1">New Password</label>
                        <input
                            type="password"
                            required
                            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            value={newPass}
                            onChange={(e) => setNewPass(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-400 mb-1">Confirm Password</label>
                        <input
                            type="password"
                            required
                            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-2 text-white focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                            value={confirmPass}
                            onChange={(e) => setConfirmPass(e.target.value)}
                        />
                    </div>

                    <div className="pt-2">
                        <button
                            type="submit"
                            disabled={loading}
                            className="bg-blue-600 hover:bg-blue-700 text-white px-6 py-2 rounded-lg font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {loading ? 'Updating...' : 'Update Password'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default UserProfile;

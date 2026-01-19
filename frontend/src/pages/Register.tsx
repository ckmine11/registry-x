import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { api } from '../lib/api';
import { UserPlus, Mail, KeyRound, Loader2, ArrowRight, User } from 'lucide-react';

const Register = () => {
    const [formData, setFormData] = useState({
        username: '',
        email: '',
        password: ''
    });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const navigate = useNavigate();

    const [recoveryKey, setRecoveryKey] = useState('');

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const res = await api.register(formData.username, formData.email, formData.password);
            // Check formatted response (recovery_key from Go snake_case)
            const key = res.data.recovery_key || res.data.recoveryKey;

            if (key) {
                setRecoveryKey(key);
            } else {
                navigate('/login');
            }
        } catch (err: any) {
            setError(err.response?.data || 'Registration failed');
        } finally {
            setIsLoading(false);
        }
    };

    if (recoveryKey) {
        return (
            <div className="min-h-screen bg-black text-white flex items-center justify-center p-4">
                <div className="absolute inset-0 z-0 bg-[radial-gradient(circle_at_bottom_right,_var(--tw-gradient-stops))] from-green-900/20 via-black to-black opacity-40"></div>
                <div className="relative z-10 w-full max-w-md bg-gray-900/60 backdrop-blur-xl rounded-2xl border border-gray-800 p-8 shadow-2xl text-center">
                    <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-green-500/10 text-green-400 mb-6 border border-green-500/20">
                        <KeyRound size={32} />
                    </div>
                    <h1 className="text-2xl font-bold text-white mb-2">Account Created!</h1>
                    <p className="text-gray-400 mb-6">Save this Recovery Key in a safe place. You will need it to reset your password if you lose access.</p>

                    <div className="bg-black/50 border border-green-500/30 rounded-lg p-4 mb-8">
                        <code className="text-xl font-mono text-green-400 tracking-wider select-all">{recoveryKey}</code>
                    </div>

                    <button
                        onClick={() => navigate('/login')}
                        className="w-full bg-green-600 hover:bg-green-500 text-white font-medium py-3 rounded-lg transition-all"
                    >
                        I've Saved It, Continue to Login
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-black text-white flex items-center justify-center p-4">
            {/* Background Effects */}
            <div className="absolute inset-0 z-0 bg-[radial-gradient(circle_at_bottom_right,_var(--tw-gradient-stops))] from-purple-900/20 via-black to-black opacity-40"></div>

            <div className="relative z-10 w-full max-w-md bg-gray-900/60 backdrop-blur-xl rounded-2xl border border-gray-800 p-8 shadow-2xl">
                <div className="text-center mb-8">
                    <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-purple-500/10 text-purple-400 mb-4">
                        <UserPlus size={24} />
                    </div>
                    <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-pink-300">
                        Create Account
                    </h1>
                    <p className="text-gray-400 mt-2 text-sm">Join RegistryX today</p>
                </div>

                {error && (
                    <div className="bg-red-500/10 border border-red-500/20 text-red-400 px-4 py-3 rounded-lg mb-6 text-sm">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-300">Username</label>
                        <div className="relative">
                            <input
                                type="text"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500 outline-none text-white transition-all placeholder-gray-600"
                                placeholder="jdoe"
                                value={formData.username}
                                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                                required
                            />
                            <User className="absolute left-3 top-3 text-gray-600" size={16} />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-300">Email Address</label>
                        <div className="relative">
                            <input
                                type="email"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500 outline-none text-white transition-all placeholder-gray-600"
                                placeholder="you@example.com"
                                value={formData.email}
                                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                required
                            />
                            <Mail className="absolute left-3 top-3 text-gray-600" size={16} />
                        </div>
                    </div>

                    <div className="space-y-2">
                        <label className="text-sm font-medium text-gray-300">Password</label>
                        <div className="relative">
                            <input
                                type="password"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500 outline-none text-white transition-all placeholder-gray-600"
                                placeholder="••••••••"
                                value={formData.password}
                                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                                required
                            />
                            <KeyRound className="absolute left-3 top-3 text-gray-600" size={16} />
                        </div>
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full bg-purple-600 hover:bg-purple-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
                    >
                        {isLoading ? <Loader2 className="animate-spin" size={20} /> : <span>Create Account</span>}
                        {!isLoading && <ArrowRight size={18} />}
                    </button>
                </form>

                <div className="mt-6 text-center text-sm text-gray-500">
                    Already have an account?{' '}
                    <Link to="/login" className="text-purple-400 hover:text-purple-300 font-medium transition-colors">
                        Sign in
                    </Link>
                </div>
            </div>
        </div>
    );
};

export default Register;

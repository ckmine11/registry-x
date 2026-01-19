import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../lib/api';
import { Mail, ArrowRight, Loader2, CheckCircle, ArrowLeft } from 'lucide-react';

const ForgotPassword = () => {
    const [mode, setMode] = useState<'link' | 'key'>('link');

    // Link Mode State
    const [email, setEmail] = useState('');
    const [linkSuccess, setLinkSuccess] = useState(false);
    const [debugToken, setDebugToken] = useState('');

    // Key Mode State
    const [keyData, setKeyData] = useState({ email: '', key: '', password: '' });
    const [keySuccess, setKeySuccess] = useState(false);

    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    const handleLinkSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const res = await api.forgotPassword(email);
            setLinkSuccess(true);
            if (res.data.debug_token) {
                setDebugToken(res.data.debug_token);
            }
        } catch (err: any) {
            setError(err.response?.data?.message || 'Request failed. Please try again.');
        } finally {
            setIsLoading(false);
        }
    };

    const handleKeySubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            await api.resetPasswordWithKey(keyData.email, keyData.key, keyData.password);
            setKeySuccess(true);
        } catch (err: any) {
            setError(err.response?.data || 'Reset failed. Check your key and email.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-black text-white flex items-center justify-center p-4">
            <div className="absolute inset-0 z-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-purple-900/20 via-black to-black opacity-40"></div>

            <div className="relative z-10 w-full max-w-md bg-gray-900/60 backdrop-blur-xl rounded-2xl border border-gray-800 p-8 shadow-2xl">

                {/* Header */}
                <div className="text-center mb-6">
                    {!linkSuccess && !keySuccess && (
                        <>
                            <div className="flex justify-center space-x-6 mb-6">
                                <button
                                    onClick={() => setMode('link')}
                                    className={`pb-2 text-sm font-medium transition-colors border-b-2 ${mode === 'link' ? 'border-purple-500 text-purple-400' : 'border-transparent text-gray-500 hover:text-gray-300'}`}
                                >
                                    Email Link
                                </button>
                                <button
                                    onClick={() => setMode('key')}
                                    className={`pb-2 text-sm font-medium transition-colors border-b-2 ${mode === 'key' ? 'border-purple-500 text-purple-400' : 'border-transparent text-gray-500 hover:text-gray-300'}`}
                                >
                                    Recovery Key
                                </button>
                            </div>
                            <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-purple-400 to-pink-300">
                                {mode === 'link' ? 'Forgot Password?' : 'Use Recovery Key'}
                            </h1>
                            <p className="text-gray-400 mt-2 text-sm">
                                {mode === 'link' ? "Enter your email and we'll send you a reset link." : "Reset your password instantly using your saved key."}
                            </p>
                        </>
                    )}

                    {(linkSuccess || keySuccess) && (
                        <>
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-green-500/10 text-green-400 mb-4">
                                <CheckCircle size={24} />
                            </div>
                            <h1 className="text-2xl font-bold text-white">
                                {keySuccess ? "Password Reset!" : "Check your email"}
                            </h1>
                            <p className="text-gray-400 mt-2 text-sm">
                                {keySuccess ? "Your password has been updated successfully." : <span>We sent a password reset link to <span className="text-white font-medium">{email}</span></span>}
                            </p>
                        </>
                    )}
                </div>

                {error && (
                    <div className="bg-red-500/10 border border-red-500/20 text-red-400 px-4 py-3 rounded-lg mb-6 text-sm">
                        {error}
                    </div>
                )}

                {/* Forms */}
                {!linkSuccess && !keySuccess && mode === 'link' && (
                    <form onSubmit={handleLinkSubmit} className="space-y-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">Email Address</label>
                            <div className="relative">
                                <input
                                    type="email"
                                    className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-purple-500/50 focus:border-purple-500 outline-none text-white transition-all placeholder-gray-600"
                                    placeholder="your@email.com"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    required
                                />
                                <Mail className="absolute left-3 top-3 text-gray-600" size={16} />
                            </div>
                        </div>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-purple-600 hover:bg-purple-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
                        >
                            {isLoading ? <Loader2 className="animate-spin" size={20} /> : <span>Send Reset Link</span>}
                            {!isLoading && <ArrowRight size={18} />}
                        </button>
                    </form>
                )}

                {!linkSuccess && !keySuccess && mode === 'key' && (
                    <form onSubmit={handleKeySubmit} className="space-y-4">
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">Email Address</label>
                            <input
                                type="email"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-purple-500/50 outline-none text-white"
                                placeholder="your@email.com"
                                value={keyData.email}
                                onChange={(e) => setKeyData({ ...keyData, email: e.target.value })}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">Recovery Key</label>
                            <input
                                type="text"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-purple-500/50 outline-none text-white font-mono placeholder-gray-600"
                                placeholder="RX-XXXX-XXXX"
                                value={keyData.key}
                                onChange={(e) => setKeyData({ ...keyData, key: e.target.value })}
                                required
                            />
                        </div>
                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">New Password</label>
                            <input
                                type="password"
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-purple-500/50 outline-none text-white"
                                placeholder="New strong password"
                                value={keyData.password}
                                onChange={(e) => setKeyData({ ...keyData, password: e.target.value })}
                                required
                            />
                        </div>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="w-full bg-purple-600 hover:bg-purple-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
                        >
                            {isLoading ? <Loader2 className="animate-spin" size={20} /> : <span>Reset Password</span>}
                        </button>
                    </form>
                )}

                {/* Success States */}
                {linkSuccess && (
                    <div className="space-y-6 mt-6">
                        {debugToken && (
                            <div className="bg-yellow-500/10 border border-yellow-500/20 p-4 rounded-lg">
                                <span className="text-xs text-yellow-500 font-bold uppercase tracking-wider block mb-1">Demo Mode</span>
                                <p className="text-sm text-yellow-200 mb-2">Simulated email sent. Use this link:</p>
                                <div className="mt-2">
                                    <Link to={`/reset-password?token=${debugToken}`} className="text-yellow-400 underline decoration-yellow-400/50 hover:decoration-yellow-400">
                                        Reset Link
                                    </Link>
                                </div>
                            </div>
                        )}
                        <Link to="/login" className="flex items-center justify-center text-gray-400 hover:text-white transition-colors text-sm font-medium">
                            <ArrowLeft size={16} className="mr-2" /> Back to Login
                        </Link>
                    </div>
                )}

                {keySuccess && (
                    <div className="space-y-6 mt-6">
                        <Link to="/login" className="block w-full text-center bg-gray-800 hover:bg-gray-700 text-white font-medium py-2.5 rounded-lg transition-all">
                            Back to Login
                        </Link>
                    </div>
                )}

                {!linkSuccess && !keySuccess && (
                    <div className="mt-6 text-center text-sm text-gray-500">
                        <Link to="/login" className="flex items-center justify-center hover:text-white transition-colors">
                            <ArrowLeft size={14} className="mr-2" /> Back to Login
                        </Link>
                    </div>
                )}

            </div>
        </div>
    );
};

export default ForgotPassword;

import React, { useState, useEffect } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { api } from '../lib/api';
import { KeyRound, ArrowRight, Loader2, CheckCircle, AlertTriangle } from 'lucide-react';

const ResetPassword = () => {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();

    const [token, setToken] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [success, setSuccess] = useState(false);
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);

    useEffect(() => {
        const tokenParam = searchParams.get('token');
        if (tokenParam) {
            setToken(tokenParam);
        } else {
            setError('Invalid reset link. Token is missing.');
        }
    }, [searchParams]);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');

        if (password !== confirmPassword) {
            setError("Passwords don't match");
            return;
        }

        if (password.length < 6) {
            setError("Password must be at least 6 characters");
            return;
        }

        setIsLoading(true);

        try {
            await api.resetPassword(token, password);
            setSuccess(true);
            setTimeout(() => {
                navigate('/login');
            }, 3000);
        } catch (err: any) {
            setError(err.response?.data?.message || 'Failed to reset password. Link may be expired.');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen bg-black text-white flex items-center justify-center p-4">
            {/* Background Effects */}
            <div className="absolute inset-0 z-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-blue-900/20 via-black to-black opacity-40"></div>

            <div className="relative z-10 w-full max-w-md bg-gray-900/60 backdrop-blur-xl rounded-2xl border border-gray-800 p-8 shadow-2xl">

                <div className="text-center mb-8">
                    {!success ? (
                        <>
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-blue-500/10 text-blue-400 mb-4">
                                <KeyRound size={24} />
                            </div>
                            <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-cyan-300">
                                Reset Password
                            </h1>
                            <p className="text-gray-400 mt-2 text-sm">Enter your new password below.</p>
                        </>
                    ) : (
                        <>
                            <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-green-500/10 text-green-400 mb-4">
                                <CheckCircle size={24} />
                            </div>
                            <h1 className="text-2xl font-bold text-white">Password Reset!</h1>
                            <p className="text-gray-400 mt-2 text-sm">
                                Your password has been updated successfully. Redirecting to login...
                            </p>
                        </>
                    )}
                </div>

                {error && (
                    <div className="bg-red-500/10 border border-red-500/20 text-red-400 px-4 py-3 rounded-lg mb-6 text-sm flex items-start">
                        <AlertTriangle size={18} className="mr-2 mt-0.5 flex-shrink-0" />
                        <span>{error}</span>
                    </div>
                )}

                {!success && (
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {!token && (
                            <div className="space-y-2">
                                <label className="text-sm font-medium text-gray-300">Reset Token</label>
                                <input
                                    type="text"
                                    className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 outline-none text-white transition-all placeholder-gray-600"
                                    placeholder="Paste token here"
                                    value={token}
                                    onChange={(e) => setToken(e.target.value)}
                                    required
                                />
                            </div>
                        )}

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">New Password</label>
                            <div className="relative">
                                <input
                                    type="password"
                                    className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 outline-none text-white transition-all placeholder-gray-600"
                                    placeholder="Minimum 6 characters"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    required
                                />
                                <KeyRound className="absolute left-3 top-3 text-gray-600" size={16} />
                            </div>
                        </div>

                        <div className="space-y-2">
                            <label className="text-sm font-medium text-gray-300">Confirm Password</label>
                            <div className="relative">
                                <input
                                    type="password"
                                    className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 outline-none text-white transition-all placeholder-gray-600"
                                    placeholder="Repeat password"
                                    value={confirmPassword}
                                    onChange={(e) => setConfirmPassword(e.target.value)}
                                    required
                                />
                                <KeyRound className="absolute left-3 top-3 text-gray-600" size={16} />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={isLoading || !token}
                            className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
                        >
                            {isLoading ? <Loader2 className="animate-spin" size={20} /> : <span>Reset Password</span>}
                            {!isLoading && <ArrowRight size={18} />}
                        </button>
                    </form>
                )}
            </div>
        </div>
    );
};

export default ResetPassword;

import React, { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../lib/auth-context';
import { api } from '../lib/api';
import { KeyRound, Mail, Loader2, ArrowRight } from 'lucide-react';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setError('');
        setIsLoading(true);

        try {
            const res = await api.login(username, password);
            login(res.data.token, res.data.user);
            navigate('/');
        } catch (err: any) {
            setError(err.response?.data?.message || 'Login failed');
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
                    <div className="inline-flex items-center justify-center w-12 h-12 rounded-xl bg-blue-500/10 text-blue-400 mb-4">
                        <KeyRound size={24} />
                    </div>
                    <h1 className="text-2xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-blue-400 to-cyan-300">
                        Welcome Back
                    </h1>
                    <p className="text-gray-400 mt-2 text-sm">Sign in to your RegistryX account</p>
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
                                className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 outline-none text-white transition-all placeholder-gray-600"
                                placeholder="Enter your username"
                                value={username}
                                onChange={(e) => setUsername(e.target.value)}
                                required
                            />
                            <Mail className="absolute left-3 top-3 text-gray-600" size={16} />
                        </div>
                    </div>

                    <div className="flex justify-between items-center text-sm">
                        <label className="text-sm font-medium text-gray-300">Password</label>
                        <Link to="/forgot-password" className="text-blue-400 hover:text-blue-300 transition-colors">
                            Forgot password?
                        </Link>
                    </div>
                    <div className="relative">
                        <input
                            type="password"
                            className="w-full bg-black/50 border border-gray-700 rounded-lg px-4 py-2.5 pl-10 focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 outline-none text-white transition-all placeholder-gray-600"
                            placeholder="••••••••"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />
                        <KeyRound className="absolute left-3 top-3 text-gray-600" size={16} />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading}
                        className="w-full bg-blue-600 hover:bg-blue-500 text-white font-medium py-2.5 rounded-lg transition-all flex items-center justify-center space-x-2 disabled:opacity-50 disabled:cursor-not-allowed mt-6"
                    >
                        {isLoading ? <Loader2 className="animate-spin" size={20} /> : <span>Sign In</span>}
                        {!isLoading && <ArrowRight size={18} />}
                    </button>
                </form>

                <div className="mt-6 text-center text-sm text-gray-500">
                    Don't have an account?{' '}
                    <Link to="/register" className="text-blue-400 hover:text-blue-300 font-medium transition-colors">
                        Create account
                    </Link>
                </div>
            </div >
        </div >
    );
};

export default Login;

import React from 'react';
import { useNavigate } from 'react-router-dom';
import {
    Shield,
    Zap,
    BarChart3,
    Command,
    Check,
    X,
    ArrowRight,
    Database,
    Lock,
    Globe,
    Cpu
} from 'lucide-react';
import clsx from 'clsx';

const FeatureCard = ({ icon: Icon, title, description }: { icon: any, title: string, description: string }) => (
    <div className="bg-slate-900/50 border border-slate-800 p-6 rounded-2xl hover:border-blue-500/50 transition-all group">
        <div className="w-12 h-12 bg-blue-500/10 rounded-xl flex items-center justify-center mb-4 group-hover:scale-110 transition-transform">
            <Icon className="text-blue-500" size={24} />
        </div>
        <h3 className="text-xl font-bold text-white mb-2">{title}</h3>
        <p className="text-slate-400 leading-relaxed">{description}</p>
    </div>
);

const ComparisonRow = ({ feature, rx, docker, harbor }: { feature: string, rx: boolean, docker: string | boolean, harbor: string | boolean }) => (
    <tr className="border-b border-slate-800 hover:bg-white/5 transition-colors">
        <td className="py-4 px-6 text-slate-300 font-medium">{feature}</td>
        <td className="py-4 px-6 text-center">
            <div className="flex justify-center">
                <div className="w-8 h-8 bg-blue-500/20 rounded-full flex items-center justify-center">
                    <Check className="text-blue-500" size={18} />
                </div>
            </div>
        </td>
        <td className="py-4 px-6 text-center text-slate-500">
            {typeof docker === 'string' ? docker : (docker ? <Check size={18} className="mx-auto" /> : <X size={18} className="mx-auto" />)}
        </td>
        <td className="py-4 px-6 text-center text-slate-500">
            {typeof harbor === 'string' ? harbor : (harbor ? <Check size={18} className="mx-auto" /> : <X size={18} className="mx-auto" />)}
        </td>
    </tr>
);

const Landing = () => {
    const navigate = useNavigate();

    return (
        <div className="min-h-screen bg-black text-white selection:bg-blue-500/30">
            {/* Header */}
            <header className="fixed top-0 left-0 right-0 z-50 bg-black/50 backdrop-blur-xl border-b border-slate-800">
                <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <div className="w-10 h-10 bg-gradient-to-br from-blue-600 to-indigo-600 rounded-xl flex items-center justify-center">
                            <span className="text-xl font-black italic">X</span>
                        </div>
                        <span className="text-2xl font-bold tracking-tight">REGISTRYX</span>
                    </div>
                    <div className="flex items-center gap-4">
                        <button
                            onClick={() => navigate('/login')}
                            className="px-6 py-2 text-slate-300 hover:text-white transition-colors"
                        >
                            Sign In
                        </button>
                        <button
                            onClick={() => navigate('/register')}
                            className="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg font-medium transition-all shadow-lg shadow-blue-500/20"
                        >
                            Get Started
                        </button>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <section className="pt-40 pb-24 px-6 relative overflow-hidden">
                <div className="absolute top-0 left-1/2 -translate-x-1/2 w-[1000px] h-[600px] bg-blue-600/10 blur-[120px] rounded-full -z-10" />
                <div className="max-w-4xl mx-auto text-center">
                    <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-sm font-medium mb-8">
                        <Zap size={14} />
                        Next-Gen Container Intelligence
                    </div>
                    <h1 className="text-6xl md:text-7xl font-black mb-8 leading-[1.1] tracking-tight">
                        Welcome to <span className="text-transparent bg-clip-text bg-gradient-to-r from-blue-400 to-indigo-500">RegistryX</span>
                    </h1>
                    <p className="text-xl text-slate-400 mb-12 leading-relaxed max-w-2xl mx-auto">
                        The world's first Private Registry with built-in Cost Intelligence and Security Autopilot.
                        Manage, secure, and optimize your container fleet in one place.
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
                        <button
                            onClick={() => navigate('/register')}
                            className="w-full sm:w-auto px-8 py-4 bg-white text-black rounded-xl font-bold flex items-center justify-center gap-2 hover:bg-slate-200 transition-all"
                        >
                            Start for Free <ArrowRight size={20} />
                        </button>
                        <button className="w-full sm:w-auto px-8 py-4 bg-slate-900 border border-slate-800 rounded-xl font-bold hover:bg-slate-800 transition-all">
                            View Demo
                        </button>
                    </div>
                </div>
            </section>

            {/* Feature Grid */}
            <section className="py-24 max-w-7xl mx-auto px-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
                    <FeatureCard
                        icon={Shield}
                        title="Security Autopilot"
                        description="Automated Trivy scanning with OPA-based policy enforcement. Block vulnerable images before they reach production."
                    />
                    <FeatureCard
                        icon={BarChart3}
                        title="Cost Intelligence"
                        description="Real-time storage and bandwidth cost tracking. Identify zombie images and reclaim wasted cloud spend."
                    />
                    <FeatureCard
                        icon={Zap}
                        title="Extreme Performance"
                        description="Built on a distributed architecture with MinIO backend. Blazing fast pulls regardless of your artifact size."
                    />
                </div>
            </section>

            {/* Comparison Section */}
            <section className="py-24 bg-slate-900/30">
                <div className="max-w-5xl mx-auto px-6">
                    <h2 className="text-4xl font-black text-center mb-4">Why Choose RegistryX?</h2>
                    <p className="text-slate-400 text-center mb-16 text-lg">See how we stack up against the industry standard.</p>

                    <div className="bg-black/40 border border-slate-800 rounded-3xl overflow-hidden backdrop-blur-md">
                        <table className="w-full">
                            <thead>
                                <tr className="bg-slate-900/50 border-b border-slate-800">
                                    <th className="py-6 px-6 text-left text-slate-400 font-bold uppercase text-xs tracking-widest">Feature</th>
                                    <th className="py-6 px-6 text-center text-blue-400 font-bold uppercase text-xs tracking-widest">RegistryX</th>
                                    <th className="py-6 px-6 text-center text-slate-400 font-bold uppercase text-xs tracking-widest">Docker Hub</th>
                                    <th className="py-6 px-6 text-center text-slate-400 font-bold uppercase text-xs tracking-widest">Harbor</th>
                                </tr>
                            </thead>
                            <tbody>
                                <ComparisonRow feature="Private Repositories" rx={true} docker={true} harbor={true} />
                                <ComparisonRow feature="Vuln. Scanning" rx={true} docker="Paid ONLY" harbor="Basic" />
                                <ComparisonRow feature="Cost Intelligence" rx={true} docker={false} harbor={false} />
                                <ComparisonRow feature="Policy Autopilot" rx={true} docker="Limited" harbor="Manual" />
                                <ComparisonRow feature="One-Click Deploy" rx={true} docker={false} harbor="Complex" />
                                <ComparisonRow feature="Modern Dashboard" rx={true} docker="Basic" harbor="Legacy" />
                            </tbody>
                        </table>
                    </div>
                </div>
            </section>

            {/* Pricing / Edition Choice */}
            {/* Pricing / Edition Choice */}
            <section className="py-24 max-w-7xl mx-auto px-6">
                <div className="text-center mb-16">
                    <h2 className="text-4xl font-black mb-4">Select Your Edition</h2>
                    <p className="text-slate-400 text-lg">Scale RegistryX as your team grows.</p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-5xl mx-auto">
                    {/* Community Edition */}
                    <div className="bg-slate-900/50 border border-slate-800 rounded-3xl p-10 flex flex-col hover:border-slate-700 transition-all">
                        <div className="mb-8">
                            <h3 className="text-2xl font-bold mb-2">Community Edition</h3>
                            <p className="text-slate-400">Perfect for individual developers and small teams.</p>
                        </div>
                        <div className="text-4xl font-black mb-8 flex items-baseline gap-1">
                            $0<span className="text-lg text-slate-500 font-medium">/forever</span>
                        </div>
                        <ul className="space-y-4 mb-10 flex-grow">
                            <li className="flex items-center gap-3 text-slate-300">
                                <Check size={18} className="text-emerald-500" /> Unlimited Public/Private Repos
                            </li>
                            <li className="flex items-center gap-3 text-slate-300">
                                <Check size={18} className="text-emerald-500" /> Basic Tag Management
                            </li>
                            <li className="flex items-center gap-3 text-slate-300">
                                <Check size={18} className="text-emerald-500" /> Modern Web Dashboard
                            </li>
                            <li className="flex items-center gap-3 text-slate-500 line-through decoration-slate-600">
                                <X size={18} /> Vulnerability Scanning
                            </li>
                            <li className="flex items-center gap-3 text-slate-500 line-through decoration-slate-600">
                                <X size={18} /> Cost Intelligence
                            </li>
                        </ul>
                        <button
                            onClick={() => {
                                localStorage.setItem('selected_edition', 'community');
                                navigate('/register');
                            }}
                            className="w-full py-4 bg-slate-800 hover:bg-slate-700 text-white rounded-xl font-bold transition-all"
                        >
                            Sign Up Free
                        </button>
                    </div>

                    {/* Enterprise Edition */}
                    <div className="bg-gradient-to-b from-blue-600/10 to-transparent border-2 border-blue-500 rounded-3xl p-10 flex flex-col relative overflow-hidden group">
                        <div className="absolute top-0 right-0 px-4 py-2 bg-blue-500 text-white text-xs font-black uppercase tracking-widest rounded-bl-2xl">
                            Recommended
                        </div>
                        <div className="mb-8">
                            <h3 className="text-2xl font-bold mb-2">Enterprise Plan</h3>
                            <p className="text-slate-400">Advanced security, costs, and team management.</p>
                        </div>
                        <div className="text-4xl font-black mb-8 flex items-baseline gap-1">
                            Contact<span className="text-lg text-slate-500 font-medium">/sales</span>
                        </div>
                        <ul className="space-y-4 mb-10 flex-grow">
                            <li className="flex items-center gap-3 text-white font-medium">
                                <Check size={18} className="text-blue-500" /> Everything in Community
                            </li>
                            <li className="flex items-center gap-3 text-white font-medium">
                                <Check size={18} className="text-blue-500" /> Vuln. Scanning (Trivy+)
                            </li>
                            <li className="flex items-center gap-3 text-white font-medium">
                                <Check size={18} className="text-blue-500" /> Cost Intelligence Dashboard
                            </li>
                            <li className="flex items-center gap-3 text-white font-medium">
                                <Check size={18} className="text-blue-500" /> OPA Policy Autopilot
                            </li>
                            <li className="flex items-center gap-3 text-white font-medium">
                                <Check size={18} className="text-blue-500" /> SSO & RBAC (Team Management)
                            </li>
                        </ul>
                        <button
                            onClick={() => {
                                localStorage.setItem('selected_edition', 'enterprise');
                                navigate('/register');
                            }}
                            className="w-full py-4 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-bold transition-all shadow-lg shadow-blue-500/25 group-hover:scale-[1.02]"
                        >
                            Try Enterprise
                        </button>
                    </div>
                </div>
            </section>

            {/* Footer */}
            <footer className="py-20 border-t border-slate-900 bg-black">
                <div className="max-w-7xl mx-auto px-6 flex flex-col md:row items-center justify-between gap-8">
                    <div className="flex items-center gap-2">
                        <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                            <span className="text-sm font-black italic">X</span>
                        </div>
                        <span className="text-xl font-bold">REGISTRYX</span>
                    </div>
                    <div className="text-slate-500 text-sm">
                        © 2026 RegistryX. Built for the modern cloud.
                    </div>
                </div>
            </footer>
        </div>
    );
};

export default Landing;

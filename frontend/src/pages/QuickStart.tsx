import React, { useState } from 'react';
import { Terminal, Copy, Check, Info, Box, ShieldCheck, Zap, Download, Database } from 'lucide-react';

const QuickStart = () => {
    return (
        <div className="space-y-12 pb-20">
            {/* Header Section */}
            <div className="relative overflow-hidden cyber-card p-12 bg-gradient-to-br from-blue-600/10 via-transparent to-purple-500/10">
                <div className="absolute top-0 right-0 w-64 h-64 bg-blue-500/5 blur-[100px] pointer-events-none" />
                <div className="relative z-10 max-w-3xl">
                    <div className="flex items-center gap-4 mb-6">
                        <div className="w-12 h-12 rounded-2xl bg-blue-500 flex items-center justify-center text-white shadow-lg shadow-blue-500/20">
                            <Zap size={24} />
                        </div>
                        <h1 className="text-4xl font-black uppercase tracking-tighter text-white">Quick Start Terminal</h1>
                    </div>
                    <p className="text-gray-400 text-sm font-medium leading-relaxed mb-8 max-w-2xl">
                        Welcome to the RegistryX Command Center. Follow these steps to authenticate and push your first image to the secure vault.
                    </p>

                    <div className="flex flex-wrap gap-4">
                        <div className="flex items-center gap-2 px-4 py-2 bg-green-500/10 rounded-xl border border-green-500/20 text-green-400 text-[10px] font-black uppercase tracking-widest">
                            <ShieldCheck size={14} /> CLI Authorized
                        </div>
                        <div className="flex items-center gap-2 px-4 py-2 bg-blue-500/10 rounded-xl border border-blue-500/20 text-blue-400 text-[10px] font-black uppercase tracking-widest">
                            <Database size={14} /> Port 5000 Active
                        </div>
                    </div>
                </div>
            </div>

            {/* Steps Grid */}
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-10">
                {/* Main Guide */}
                <div className="lg:col-span-8 space-y-10">
                    <GuideStep
                        number="01"
                        title="Registry Authentication"
                        desc="Login to the local OCI registry using your RegistryX credentials."
                        cmd="docker login localhost:5000"
                        label="AUTH"
                    />

                    <GuideStep
                        number="02"
                        title="Tag your Image"
                        desc="Prepare your local image for the registry by giving it a full OCI address."
                        cmd="docker tag my-app:latest localhost:5000/avinash/my-app:v1"
                        label="TAG"
                    />

                    <GuideStep
                        number="03"
                        title="Upload to Vault"
                        desc="Push the image to our secure storage. Scanning will trigger automatically."
                        cmd="docker push localhost:5000/avinash/my-app:v1"
                        label="PUSH"
                    />

                    <GuideStep
                        number="04"
                        title="Verify Scan & Score"
                        desc="Pull the image to verify security policies have been evaluated."
                        cmd="docker pull localhost:5000/avinash/my-app:v1"
                        label="PULL"
                    />
                </div>

                {/* Sidebar Tips */}
                <div className="lg:col-span-4 space-y-6">
                    <div className="cyber-card p-6 border-blue-500/20">
                        <div className="flex items-center gap-3 mb-4 text-blue-400">
                            <Info size={18} />
                            <h3 className="font-black uppercase text-[11px] tracking-widest">Expert Tip</h3>
                        </div>
                        <p className="text-[11px] text-gray-500 leading-relaxed uppercase font-bold tracking-tight">
                            Use the <span className="text-white">namespace/repo</span> format to organize images. Your personal namespace is already allocated based on your username.
                        </p>
                    </div>

                    <div className="cyber-card p-6 border-purple-500/20">
                        <div className="flex items-center gap-3 mb-4 text-purple-400">
                            <Box size={18} />
                            <h3 className="font-black uppercase text-[11px] tracking-widest">Oci Compatibility</h3>
                        </div>
                        <p className="text-[11px] text-gray-500 leading-relaxed uppercase font-bold tracking-tight">
                            RegistryX supports Docker, Podman, and Skopeo. All images are scanned via Trivy during the push process.
                        </p>
                    </div>

                    <div className="p-8 rounded-3xl bg-gradient-to-br from-blue-600 to-purple-600 relative overflow-hidden group">
                        <div className="absolute inset-0 bg-black/20 group-hover:bg-transparent transition-colors" />
                        <Download className="text-white/20 absolute -bottom-4 -right-4 w-32 h-32 group-hover:scale-110 transition-transform" />
                        <h4 className="text-lg font-black text-white uppercase tracking-tighter mb-2 relative z-10">Download CLI Client</h4>
                        <p className="text-white/70 text-[10px] font-bold uppercase tracking-widest mb-6 relative z-10">Control RegistryX from your native terminal.</p>
                        <button className="w-full py-3 bg-white text-blue-600 rounded-xl font-black uppercase text-[10px] tracking-widest shadow-xl relative z-10 active:scale-95 transition-all">
                            Coming Soon
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

const GuideStep = ({ number, title, desc, cmd, label }: { number: string; title: string; desc: string; cmd: string; label: string }) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
        navigator.clipboard.writeText(cmd);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="flex gap-8 group">
            <div className="flex flex-col items-center">
                <div className="w-10 h-10 rounded-xl bg-white/5 border border-white/5 flex items-center justify-center font-black text-blue-500 group-hover:bg-blue-500 group-hover:text-white transition-all text-sm group-hover:shadow-[0_0_15px_rgba(59,130,246,0.3)]">
                    {number}
                </div>
                <div className="w-[1px] flex-1 bg-gradient-to-b from-blue-500/20 to-transparent mt-4" />
            </div>

            <div className="flex-1 space-y-4">
                <div>
                    <h3 className="text-lg font-black text-white uppercase tracking-tight group-hover:text-blue-400 transition-colors">{title}</h3>
                    <p className="text-xs text-gray-500 font-medium">{desc}</p>
                </div>

                <div className="relative">
                    <div className="flex items-center bg-black/50 border border-white/5 rounded-2xl group-hover:border-blue-500/20 transition-all overflow-hidden pl-4 pr-16 py-4">
                        <Terminal size={14} className="text-gray-600 mr-4" />
                        <code className="text-[11px] font-mono text-gray-400 group-hover:text-white transition-colors truncate">
                            {cmd}
                        </code>
                        <div className="absolute right-2 top-1/2 -translate-y-1/2">
                            <button
                                onClick={handleCopy}
                                className="p-3 bg-white/5 hover:bg-white/10 rounded-xl text-gray-500 hover:text-white transition-all border border-transparent hover:border-white/10"
                            >
                                {copied ? <Check size={16} className="text-green-500" /> : <Copy size={16} />}
                            </button>
                        </div>
                    </div>
                    <div className="absolute -top-3 right-4 bg-black border border-white/10 px-2 py-0.5 rounded-md text-[8px] font-black text-blue-500 tracking-widest uppercase">
                        {label}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default QuickStart;

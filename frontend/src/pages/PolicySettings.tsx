import React, { useState, useEffect } from 'react';
import { Shield, Save, AlertTriangle, CheckCircle, FileText, Lock } from 'lucide-react';
import api from '../lib/api';

const PolicySettings = () => {
    const [policy, setPolicy] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [message, setMessage] = useState<{ type: 'success' | 'error', text: string } | null>(null);

    useEffect(() => {
        loadPolicy();
    }, []);

    const loadPolicy = async () => {
        try {
            setLoading(true);
            const res = await api.getPolicy();
            setPolicy(res.data.rego);
        } catch (err) {
            setMessage({ type: 'error', text: 'Failed to load policy' });
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        try {
            setSaving(true);
            await api.updatePolicy(policy);
            setMessage({ type: 'success', text: 'Policy updated successfully and is now active.' });
        } catch (err) {
            setMessage({ type: 'error', text: 'Failed to update policy. Check syntax.' });
        } finally {
            setSaving(false);
        }
    };

    const applyTemplate = (type: 'strict' | 'audit') => {
        if (type === 'strict') {
            setPolicy(`package registryx.policy

default allow = true

# STRICT MODE: Block Criticals in Prod
violations[msg] {
	input.vulnerabilities.critical > 0
	msg := sprintf("BLOCK: Image has %d critical vulnerabilities.", [input.vulnerabilities.critical])
}

violations[msg] {
	input.is_signed == false
	msg := "BLOCK: Image is not signed."
}

allow = false {
	count(violations) > 0
}`);
        } else {
            setPolicy(`package registryx.policy

default allow = true

# AUDIT MODE: Log but Allow
# violations[msg] {
# 	input.vulnerabilities.critical > 0
# 	msg := "WARN: Critical vulnerabilities found"
# }

allow = true`);
        }
    };

    return (
        <div className="p-8 max-w-6xl mx-auto text-gray-200">
            <header className="mb-8">
                <h1 className="text-3xl font-black text-white flex items-center gap-3 tracking-tight">
                    <Shield className="w-8 h-8 text-emerald-400" />
                    POLICY GATEKEEPER
                </h1>
                <p className="text-gray-400 mt-2 text-lg">
                    Define active security rules. Violations will block <code className="text-emerald-400 bg-black/30 px-2 py-1 rounded">docker pull</code> requests.
                </p>
            </header>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
                {/* Editor Column */}
                <div className="lg:col-span-2 space-y-4">
                    <div className="bg-[#0f111a] border border-white/10 rounded-xl overflow-hidden shadow-2xl">
                        <div className="flex items-center justify-between px-4 py-3 bg-white/5 border-b border-white/5">
                            <span className="flex items-center gap-2 text-xs font-bold text-gray-400 uppercase tracking-widest">
                                <FileText size={14} /> policy.rego
                            </span>
                            <span className="text-xs text-emerald-400 font-mono">OPA Engine Active</span>
                        </div>
                        <textarea
                            value={policy}
                            onChange={(e) => setPolicy(e.target.value)}
                            className="w-full h-[500px] bg-[#0f111a] text-gray-300 font-mono text-sm p-4 focus:outline-none resize-none"
                            spellCheck="false"
                        />
                    </div>
                </div>

                {/* Controls Column */}
                <div className="space-y-6">
                    {/* Status Card */}
                    <div className="bg-white/5 border border-white/10 rounded-2xl p-6">
                        <h3 className="font-bold text-white mb-4 flex items-center gap-2">
                            <Lock size={18} /> Enforcement Mode
                        </h3>
                        <p className="text-sm text-gray-400 mb-6">
                            Changes apply immediately to all incoming requests.
                        </p>

                        <div className="space-y-3">
                            <button
                                onClick={() => applyTemplate('strict')}
                                className="w-full flex items-center gap-3 px-4 py-3 bg-red-500/10 hover:bg-red-500/20 border border-red-500/20 text-red-400 rounded-xl transition-all text-sm font-bold text-left"
                            >
                                <Shield size={16} />
                                <div>
                                    <div className="uppercase text-xs tracking-wider opacity-70">Template</div>
                                    Strict Blocking
                                </div>
                            </button>

                            <button
                                onClick={() => applyTemplate('audit')}
                                className="w-full flex items-center gap-3 px-4 py-3 bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/20 text-blue-400 rounded-xl transition-all text-sm font-bold text-left"
                            >
                                <CheckCircle size={16} />
                                <div>
                                    <div className="uppercase text-xs tracking-wider opacity-70">Template</div>
                                    Audit Only (Allow All)
                                </div>
                            </button>
                        </div>

                        <div className="h-px bg-white/10 my-6" />

                        <button
                            onClick={handleSave}
                            disabled={saving}
                            className={`w-full flex items-center justify-center gap-2 px-6 py-4 rounded-xl font-bold uppercase tracking-widest transition-all ${saving
                                ? 'bg-emerald-600/50 cursor-wait'
                                : 'bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-900/20'
                                }`}
                        >
                            <Save size={18} />
                            {saving ? 'Deploying...' : 'Deploy Policy'}
                        </button>

                        {message && (
                            <div className={`mt-4 p-3 rounded-lg text-xs font-bold flex items-center gap-2 ${message.type === 'success' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'
                                }`}>
                                {message.type === 'success' ? <CheckCircle size={14} /> : <AlertTriangle size={14} />}
                                {message.text}
                            </div>
                        )}
                    </div>

                    {/* Info Card */}
                    <div className="bg-blue-500/5 border border-blue-500/10 rounded-2xl p-6">
                        <h4 className="text-blue-400 font-bold text-sm mb-2 flex items-center gap-2">
                            Documentation
                        </h4>
                        <p className="text-xs text-gray-400 leading-relaxed">
                            Policies are written in <strong>Rego</strong>. The <code>input</code> object contains:
                        </p>
                        <ul className="mt-2 space-y-1 text-xs text-gray-500 font-mono">
                            <li>input.vulnerabilities.critical</li>
                            <li>input.vulnerabilities.high</li>
                            <li>input.is_signed (bool)</li>
                            <li>input.environment (prod/dev)</li>
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default PolicySettings;

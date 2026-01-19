import React, { useState, useEffect } from 'react';
import { api } from '../lib/api';
import { Save, CheckCircle } from 'lucide-react';

export default function Policies() {
    const [repo, setRepo] = useState('');

    useEffect(() => {
        api.getPolicy().then((res: any) => setRepo(res.data.rego));
    }, []);

    const handleSave = async () => {
        try {
            await api.updatePolicy(repo);
            // Optionally add a toast here. For now, we rely on the UI state which implies "saved"
            console.log("Policy updated successfully");
        } catch (e) {
            console.error(e);
            alert("Failed to update policy");
        }
    };

    return (
        <div className="p-6 space-y-6">
            <div className="flex items-center justify-between">
                <h1 className="text-3xl font-bold tracking-tight text-white">Policy Management</h1>
                <button
                    onClick={handleSave}
                    className="flex items-center gap-2 bg-blue-600 hover:bg-blue-500 text-white px-4 py-2 rounded-lg font-medium transition-colors"
                >
                    <Save className="w-4 h-4" /> Save Policy
                </button>
            </div>

            <div className="bg-white/5 border border-white/10 rounded-xl p-6">
                <div className="flex items-center gap-2 mb-4 text-green-400">
                    <CheckCircle className="w-4 h-4" />
                    <span className="text-sm font-mono">Status: Active (Enforced)</span>
                </div>

                <p className="text-gray-400 mb-2 text-sm">Edit your OPA Rego policy below. This policy is evaluated on every 'docker pull'.</p>

                <textarea
                    value={repo}
                    onChange={(e) => setRepo(e.target.value)}
                    className="w-full h-[500px] bg-black/40 border border-white/10 rounded-lg p-4 font-mono text-sm text-blue-100 focus:outline-none focus:border-blue-500"
                    spellCheck={false}
                />
            </div>
        </div>
    );
}

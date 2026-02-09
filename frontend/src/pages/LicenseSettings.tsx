import React from 'react';
import { Shield as ShieldIcon, Key as KeyIcon, CheckCircle2, AlertCircle, ExternalLink, Copy, Zap as ZapIcon, Info as InfoIcon } from 'lucide-react';
import { api } from '../lib/api';

type StatusType = 'idle' | 'loading' | 'success' | 'error';

interface LicenseConfig {
    license_plan?: string;
    [key: string]: any;
}

const LicenseSettings = () => {
    const [licenseKey, setLicenseKey] = React.useState('');
    const [status, setStatus] = React.useState<{ type: StatusType, message?: string }>({ type: 'idle' });
    const [config, setConfig] = React.useState<LicenseConfig | null>(null);

    const fetchConfig = async () => {
        try {
            const res = await api.getSystemConfig();
            setConfig(res.data);
        } catch (err) {
            console.error(err);
        }
    };

    React.useEffect(() => {
        fetchConfig();
    }, []);

    const handleUpdate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!licenseKey.trim()) return;

        setStatus({ type: 'loading' });
        try {
            await api.post('/api/v1/system/license', { licenseKey });
            setStatus({ type: 'success', message: 'License updated successfully!' });
            setLicenseKey('');
            fetchConfig();
            // Optional: trigger a full page reload or context refresh
            setTimeout(() => window.location.reload(), 2000);
        } catch (err: any) {
            const errorMessage = err.response?.data?.message || err.response?.data || 'Failed to update license. Please check your key.';
            setStatus({
                type: 'error',
                message: typeof errorMessage === 'string' ? errorMessage : JSON.stringify(errorMessage)
            });
        }
    };

    const isEnterprise = config?.license_plan?.toUpperCase() === 'ENTERPRISE';

    return (
        <div className="space-y-8 max-w-4xl animate-in fade-in slide-in-from-bottom-4 duration-500">
            {/* Current Plan Card */}
            <div className={`relative overflow-hidden rounded-3xl border p-8 transition-all ${isEnterprise
                ? 'bg-gradient-to-br from-blue-600/20 via-blue-900/10 to-transparent border-blue-500/30'
                : 'bg-slate-900/50 border-slate-800'
                }`}>
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-6 relative z-10">
                    <div className="space-y-2">
                        <div className="flex items-center gap-3">
                            <h2 className="text-3xl font-black tracking-tight text-white uppercase italic">
                                {config?.license_plan || 'Community Edition'}
                            </h2>
                            {isEnterprise && (
                                <span className="bg-blue-500 text-white text-[10px] font-black px-2 py-0.5 rounded-full uppercase tracking-widest animate-pulse">
                                    Active
                                </span>
                            )}
                        </div>
                        <p className="text-slate-400 font-medium">
                            {isEnterprise
                                ? 'Full access to all premium security and cost features.'
                                : 'Limited features. Upgrade to Enterprise for advanced security policies.'}
                        </p>
                    </div>

                    <div className="flex items-center gap-4">
                        <div className="text-right hidden sm:block">
                            <div className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">Status</div>
                            <div className={`text-sm font-mono font-bold uppercase tracking-tighter ${isEnterprise ? 'text-blue-400' : 'text-amber-500'}`}>
                                {isEnterprise ? 'Validated & Secure' : 'Basic Mode'}
                            </div>
                        </div>
                        <div className={`w-14 h-14 rounded-2xl flex items-center justify-center ${isEnterprise ? 'bg-blue-500/20 text-blue-400' : 'bg-slate-800 text-slate-500'
                            }`}>
                            {isEnterprise ? <ZapIcon size={28} /> : <ShieldIcon size={28} />}
                        </div>
                    </div>
                </div>

                {/* Features Checklist */}
                <div className="grid grid-cols-1 sm:grid-cols-2 mt-8 gap-4 pt-8 border-t border-white/5">
                    {[
                        { name: 'Vulnerability Scanning', enabled: isEnterprise },
                        { name: 'Cost Intelligence', enabled: isEnterprise },
                        { name: 'OPA Policy Engine', enabled: isEnterprise },
                        { name: 'RBAC & Team Management', enabled: isEnterprise },
                        { name: 'Audit Logs', enabled: isEnterprise },
                        { name: 'Priority Support', enabled: isEnterprise },
                    ].map((f) => (
                        <div key={f.name} className="flex items-center gap-3">
                            <div className={`w-5 h-5 rounded-full flex items-center justify-center ${f.enabled ? 'bg-blue-500/20 text-blue-400' : 'bg-slate-800 text-slate-600'
                                }`}>
                                <CheckCircle2 size={12} />
                            </div>
                            <span className={`text-sm font-medium ${f.enabled ? 'text-slate-200' : 'text-slate-600'}`}>
                                {f.name}
                            </span>
                        </div>
                    ))}
                </div>
            </div>

            {/* License Activation Form */}
            <div className="bg-slate-900/30 border border-slate-800 rounded-3xl p-8">
                <div className="flex items-center gap-3 mb-6">
                    <KeyIcon className="text-blue-400" size={20} />
                    <h3 className="text-xl font-bold text-white">Activate Enterprise</h3>
                </div>

                <form onSubmit={handleUpdate} className="space-y-6">
                    <div>
                        <label className="block text-[10px] font-black text-slate-500 uppercase tracking-[0.2em] mb-3">
                            License Activation Key
                        </label>
                        <div className="relative group">
                            <input
                                type="text"
                                value={licenseKey}
                                onChange={(e) => setLicenseKey(e.target.value)}
                                placeholder="eyJhbGciOiJub25lIn0..."
                                className="w-full bg-black/50 border border-slate-800 rounded-2xl py-4 pl-6 pr-32 text-sm font-mono text-blue-300 placeholder:text-slate-700 focus:outline-none focus:border-blue-500/50 transition-all"
                            />
                            <div className="absolute right-2 top-2 bottom-2">
                                <button
                                    type="submit"
                                    disabled={status.type === 'loading' || !licenseKey}
                                    className="h-full px-6 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-800 disabled:text-slate-500 text-white rounded-xl font-bold text-xs uppercase tracking-widest transition-all shadow-lg shadow-blue-500/20"
                                >
                                    {status.type === 'loading' ? 'Verifying...' : 'Activate'}
                                </button>
                            </div>
                        </div>
                        <p className="mt-3 text-xs text-slate-500 flex items-center gap-2">
                            <InfoIcon size={12} />
                            Paste your Enterprise License JWT token to unlock all advanced features.
                        </p>
                    </div>

                    {status.message && (
                        <div className={`p-4 rounded-2xl flex items-center gap-3 animate-in fade-in slide-in-from-top-2 ${status.type === 'error' ? 'bg-red-500/10 border border-red-500/20 text-red-400' : 'bg-green-500/10 border border-green-500/20 text-green-400'
                            }`}>
                            {status.type === 'error' ? <AlertCircle size={18} /> : <CheckCircle2 size={18} />}
                            <span className="text-sm font-medium">{status.message}</span>
                        </div>
                    )}
                </form>
            </div>

            {/* Support / Get Key */}
            <div className="bg-blue-500/5 border border-blue-500/10 rounded-3xl p-8 flex flex-col md:row items-center justify-between gap-6">
                <div className="space-y-1">
                    <h4 className="text-lg font-bold text-white">Don't have a license key?</h4>
                    <p className="text-slate-400 text-sm">Contact our sales team or get a 30-day free trial key.</p>
                </div>
                <button className="px-8 py-3 bg-white text-black rounded-xl font-bold flex items-center gap-2 hover:bg-slate-200 transition-all whitespace-nowrap">
                    Talk to Sales <ExternalLink size={18} />
                </button>
            </div>
        </div>
    );
};

export default LicenseSettings;

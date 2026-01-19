import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api, registry, ScanStatus, ScanHistoryEntry } from '../lib/api';
import { Shield, ShieldAlert, CheckCircle, XCircle, Trash2, ArrowLeft, Download, Clock, RefreshCw, History, Eye, X, Activity, Database, Fingerprint, Zap } from 'lucide-react';
import clsx from 'clsx';
import { HealthBadge } from '../components/HealthBadge';
import { Modal } from '../components/Modal';

interface TrivyVulnerability {
    VulnerabilityID: string;
    PkgName: string;
    InstalledVersion: string;
    FixedVersion?: string;
    Severity: string;
    PrimaryURL?: string;
}

interface TrivyResult {
    Target: string;
    Vulnerabilities?: TrivyVulnerability[];
}

interface TrivyReport {
    Results?: TrivyResult[];
}

export default function RepositoryDetails() {
    const { name } = useParams<{ name: string }>();
    const [selectedTag, setSelectedTag] = useState<string | null>(null);
    const [showScanHistory, setShowScanHistory] = useState(false);
    const [showVulnModal, setShowVulnModal] = useState(false);
    const [vulnReport, setVulnReport] = useState<TrivyReport | null>(null);
    const [isLoadingReport, setIsLoadingReport] = useState(false);
    // Modal states
    const [tagToDelete, setTagToDelete] = useState<string | null>(null);
    const [alertMessage, setAlertMessage] = useState<string | null>(null);

    const queryClient = useQueryClient();

    const handleDeleteTag = async () => {
        if (!tagToDelete || !name) return;
        try {
            await registry.deleteTag(name, tagToDelete);
            await queryClient.invalidateQueries({ queryKey: ['tags', name] });
            if (selectedTag === tagToDelete) setSelectedTag(null);
            setTagToDelete(null);
        } catch (e) {
            console.error("Failed to delete tag", e);
            setTagToDelete(null);
            setAlertMessage("CRITICAL_FAILURE: ACTION_DENIED");
        }
    };

    const handleViewFixes = async () => {
        if (!name || !selectedTag) return;
        setIsLoadingReport(true);
        setShowVulnModal(true);
        try {
            const res = await api.getScanReportJSON(name, selectedTag);
            setVulnReport(res.data);
        } catch (e) {
            console.error("Failed to fetch report", e);
            setAlertMessage("TELEMETRY_ERROR: DATA_UNAVAILABLE");
            setShowVulnModal(false);
        } finally {
            setIsLoadingReport(false);
        }
    };

    const handleDownloadReport = async () => {
        if (!name || !selectedTag) return;
        try {
            await api.downloadScanReport(name, selectedTag);
        } catch (e) {
            console.error("Failed to download report", e);
            setAlertMessage("DOWNLOAD_FAILURE");
        }
    };

    const handleTriggerScan = async () => {
        if (!name || !selectedTag) return;

        queryClient.setQueryData(['scanStatus', name, selectedTag], (old: { data: ScanStatus } | undefined) => {
            if (!old) return old;
            return {
                ...old,
                data: {
                    ...old.data,
                    status: 'scanning',
                    progress_message: 'Calibrating neural scanners...',
                    scanned_at: new Date().toISOString()
                }
            };
        });

        try {
            await api.triggerManualScan(name, selectedTag);
        } catch (e: any) {
            console.error("Failed to trigger scan", e);
            queryClient.invalidateQueries({ queryKey: ['scanStatus', name, selectedTag] });
            setAlertMessage("SCAN_INIT_FAILURE");
        }
    };

    const { data: tagsData, isLoading: isTagsLoading } = useQuery({
        queryKey: ['tags', name],
        queryFn: () => api.getTags(name!),
        enabled: !!name,
    });

    useEffect(() => {
        if (tagsData && tagsData.data && tagsData.data.tags && tagsData.data.tags.length > 0 && !selectedTag) {
            setSelectedTag(tagsData.data.tags[0]);
        }
    }, [tagsData, selectedTag]);

    const { data: manifestData } = useQuery({
        queryKey: ['manifest', name, selectedTag],
        queryFn: () => api.getManifestDetails(name!, selectedTag!),
        enabled: !!selectedTag,
    });

    const { data: scanStatusData, refetch: refetchScanStatus } = useQuery({
        queryKey: ['scanStatus', name, selectedTag],
        queryFn: () => api.getScanStatus(name!, selectedTag!),
        enabled: !!selectedTag && !!name,
        refetchInterval: (query: any) => {
            const status = query.state.data?.data?.status;
            return status === 'scanning' ? 2000 : false;
        },
    });

    const { data: scanHistoryData } = useQuery({
        queryKey: ['scanHistory', name, selectedTag],
        queryFn: () => api.getScanHistory(name!, selectedTag!),
        enabled: showScanHistory && !!selectedTag && !!name,
    });

    const details = manifestData?.data;
    const scanStatus = scanStatusData?.data || { status: 'pending' };
    const scanHistory = scanHistoryData?.data?.scans || [];

    return (
        <div className="space-y-10 pb-20">
            {/* Header Readout */}
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                <div>
                    <div className="flex items-center gap-4 mb-2">
                        <Link to="/repositories" className="text-gray-500 hover:text-blue-400 transition-colors">
                            <ArrowLeft size={20} />
                        </Link>
                        <h1 className="text-4xl font-black uppercase tracking-tighter text-white">Entity Inspector</h1>
                    </div>
                    <p className="text-blue-400 font-mono text-sm tracking-[0.2em] uppercase opacity-70">Repository Context: <span className="text-white font-black">{name}</span></p>
                </div>

                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-6 px-6 py-3 bg-white/5 border border-white/5 rounded-2xl">
                        <div className="flex flex-col">
                            <span className="text-xs font-black text-gray-500 uppercase tracking-widest">Active Sequence</span>
                            <span className="text-lg font-black text-white leading-tight">{selectedTag || 'NONE'}</span>
                        </div>
                        <div className="w-px h-8 bg-white/10" />
                        <div className="flex items-center gap-2 px-3 py-1 bg-green-500/10 rounded-lg text-green-400 font-mono text-sm uppercase font-bold">
                            <Activity size={10} className="animate-pulse" />
                            Live Telemetry
                        </div>
                    </div>
                </div>
            </div>

            <div className="grid grid-cols-1 xl:grid-cols-4 gap-8">
                {/* Artifact Sequence (Tag List) */}
                <div className="xl:col-span-1 space-y-6">
                    <div className="cyber-card p-6 h-[calc(100vh-280px)] flex flex-col">
                        <div className="flex items-center justify-between mb-6">
                            <h2 className="text-sm font-black text-gray-400 uppercase tracking-widest">Artifact Index</h2>
                            <Database size={12} className="text-gray-600" />
                        </div>

                        {isTagsLoading ? (
                            <div className="space-y-3">
                                {[1, 2, 3, 4, 5].map(i => <div key={i} className="h-12 bg-white/5 animate-pulse rounded-xl" />)}
                            </div>
                        ) : (
                            <div className="flex-1 space-y-2 overflow-y-auto pr-2 custom-scrollbar">
                                {tagsData?.data?.tags?.map((tag: string) => (
                                    <button
                                        key={tag}
                                        onClick={() => setSelectedTag(tag)}
                                        className={clsx(
                                            "w-full flex items-center justify-between px-4 py-4 rounded-xl border transition-all group",
                                            selectedTag === tag
                                                ? "bg-blue-600/10 border-blue-500/30 text-white shadow-[0_0_15px_rgba(59,130,246,0.1)]"
                                                : "bg-black/20 border-white/5 text-gray-500 hover:text-white hover:border-white/10"
                                        )}
                                    >
                                        <span className="font-mono text-sm font-black uppercase tracking-widest truncate max-w-[120px]">{tag}</span>
                                        <button
                                            onClick={(e) => { e.stopPropagation(); setTagToDelete(tag); }}
                                            className="text-gray-700 hover:text-red-500 p-1 opacity-0 group-hover:opacity-100 transition-all active:scale-90"
                                        >
                                            <Trash2 size={12} />
                                        </button>
                                    </button>
                                )) || <div className="text-sm font-mono text-gray-700 italic">EMPTY_INDEX</div>}
                            </div>
                        )}
                    </div>
                </div>

                {/* Main Readout Display */}
                <div className="xl:col-span-3 space-y-8">
                    {selectedTag && details ? (
                        <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-500">
                            {/* Technical Specs */}
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                                <StatBox icon={Fingerprint} label="MANIFEST_DIGEST" value={details.digest.substring(0, 12)} fullValue={details.digest} />
                                <StatBox icon={Activity} label="TOTAL_MASS" value={`${(details.size / 1024 / 1024).toFixed(2)} MB`} />
                                <StatBox
                                    icon={details.isSigned ? CheckCircle : ShieldAlert}
                                    label="SIGNAL_INTEGRITY"
                                    value={details.isSigned ? "VERIFIED" : "UNSIGNED"}
                                    accent={details.isSigned ? "green" : "red"}
                                />
                                <StatBox icon={Zap} label="RECOVERY_STATE" value="ACTIVE" accent="blue" />
                            </div>

                            <div className="grid grid-cols-1 lg:grid-cols-5 gap-8">
                                {/* Scan Logic Console */}
                                <div className="lg:col-span-3 space-y-8">
                                    <div className="cyber-card p-1 overflow-hidden relative group">
                                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 to-transparent" />
                                        <div className="p-8 space-y-8">
                                            <div className="flex items-center justify-between">
                                                <div className="flex items-center gap-4">
                                                    <div className="w-12 h-12 rounded-2xl bg-blue-600/10 flex items-center justify-center text-blue-400 border border-blue-500/20">
                                                        <Shield size={24} />
                                                    </div>
                                                    <div>
                                                        <h2 className="text-2xl font-black uppercase tracking-tight text-white">Security Telemetry</h2>
                                                        <p className="text-sm font-mono text-gray-500 uppercase tracking-widest">Real-time Artifact Diagnostics</p>
                                                    </div>
                                                </div>
                                                <button onClick={() => refetchScanStatus()} className="p-3 text-gray-500 hover:text-white transition-colors"><RefreshCw size={20} /></button>
                                            </div>

                                            <ScanStatusPanel status={scanStatus} />

                                            <div className="grid grid-cols-2 gap-4">
                                                <button
                                                    onClick={handleViewFixes}
                                                    className="flex items-center justify-center gap-3 px-6 py-4 bg-blue-600 text-white rounded-2xl font-black uppercase text-sm tracking-widest hover:bg-blue-500 transition-all shadow-[0_0_20px_rgba(37,99,235,0.3)]"
                                                >
                                                    <Eye size={16} /> Evaluate Remedies
                                                </button>
                                                <button
                                                    onClick={handleDownloadReport}
                                                    className="flex items-center justify-center gap-3 px-6 py-4 bg-white/5 text-gray-400 rounded-2xl font-black uppercase text-[10px] tracking-widest hover:bg-white/10 transition-all border border-white/5"
                                                >
                                                    <Download size={16} /> Export JSON
                                                </button>
                                            </div>

                                            {scanStatus.status !== 'scanning' && (
                                                <button
                                                    onClick={handleTriggerScan}
                                                    className="w-full flex items-center justify-center gap-3 px-6 py-3 bg-white/5 text-green-500 rounded-2xl font-black uppercase text-[9px] tracking-[0.3em] hover:bg-green-500/10 transition-all border border-green-500/20"
                                                >
                                                    <RefreshCw size={14} /> Re-Initialize Scan Protocol
                                                </button>
                                            )}
                                        </div>
                                    </div>

                                    {/* Vulnerability Matrix */}
                                    <div className="cyber-card p-8">
                                        <h3 className="text-[10px] font-black text-gray-500 uppercase tracking-widest mb-6">Threat Exposure Matrix</h3>
                                        {details.vulnerabilities ? (
                                            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                                                <VulnBox label="Critical" count={details.vulnerabilities.critical} color="red" />
                                                <VulnBox label="High" count={details.vulnerabilities.high} color="orange" />
                                                <VulnBox label="Medium" count={details.vulnerabilities.medium} color="yellow" />
                                                <VulnBox label="Low" count={details.vulnerabilities.low} color="blue" />
                                            </div>
                                        ) : (
                                            <div className="h-24 flex items-center justify-center border border-dashed border-white/5 rounded-2xl text-[10px] font-mono text-gray-600 uppercase">NO_DATA_AVAILABLE</div>
                                        )}
                                    </div>
                                </div>

                                {/* Sidebar Diagnostics */}
                                <div className="lg:col-span-2 space-y-8">
                                    {/* Health Gauge */}
                                    <div className="cyber-card p-8 relative overflow-hidden">
                                        <div className="absolute top-0 right-0 p-4 opacity-5 text-white/20"><Activity size={120} /></div>
                                        <h3 className="text-[10px] font-black text-gray-500 uppercase tracking-widest mb-6">Entity Bio-Sign</h3>
                                        {details.healthScore && (
                                            <div className="space-y-6">
                                                <HealthBadge score={details.healthScore} showDetails={true} />
                                                <div className="pt-4 border-t border-white/5 grid grid-cols-2 gap-4">
                                                    <div className="space-y-1">
                                                        <div className="text-[8px] text-gray-600 font-black uppercase">Age Index</div>
                                                        <div className="text-xs text-white font-mono">14 Cycles</div>
                                                    </div>
                                                    <div className="space-y-1">
                                                        <div className="text-[8px] text-gray-600 font-black uppercase">Pull Frequency</div>
                                                        <div className="text-xs text-white font-mono">High Density</div>
                                                    </div>
                                                </div>
                                            </div>
                                        )}
                                    </div>

                                    {/* Iteration History */}
                                    <div className="cyber-card p-8 flex flex-col">
                                        <div className="flex items-center justify-between mb-6">
                                            <h3 className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Temporal Log</h3>
                                            <button
                                                onClick={() => setShowScanHistory(!showScanHistory)}
                                                className="text-[8px] font-black text-blue-500 hover:text-blue-400 uppercase tracking-widest"
                                            >
                                                {showScanHistory ? '[HIDE]' : '[UNFOLD]'}
                                            </button>
                                        </div>
                                        {showScanHistory ? (
                                            <div className="space-y-3 max-h-[300px] overflow-y-auto pr-2 custom-scrollbar">
                                                {scanHistory.length > 0 ? scanHistory.map((scan: ScanHistoryEntry) => (
                                                    <div key={scan.id} className="p-4 bg-black/40 rounded-xl border border-white/5 space-y-2 group hover:border-white/10 transition-colors">
                                                        <div className="flex justify-between items-center text-[8px] uppercase font-bold">
                                                            <div className="flex items-center gap-2">
                                                                <ScanStatusIconInline status={scan.status} />
                                                                <span className={clsx(
                                                                    scan.status === 'completed' ? "text-green-500" : "text-gray-500"
                                                                )}>{scan.status}</span>
                                                            </div>
                                                            <span className="text-gray-700">{new Date(scan.scanned_at || '').toLocaleDateString()}</span>
                                                        </div>
                                                        {scan.summary && (
                                                            <div className="flex gap-4 text-[9px] font-mono">
                                                                <span className="text-red-500">C:{scan.summary.critical}</span>
                                                                <span className="text-orange-500">H:{scan.summary.high}</span>
                                                                <span className="text-yellow-500">M:{scan.summary.medium}</span>
                                                            </div>
                                                        )}
                                                    </div>
                                                )) : <div className="text-center font-mono text-[10px] text-gray-700 py-10 uppercase">NO_HISTORY_LOGGED</div>}
                                            </div>
                                        ) : (
                                            <div className="h-4 p-4 border border-dashed border-white/5 rounded-xl flex items-center justify-center">
                                                <div className="w-1 h-1 bg-gray-800 rounded-full animate-pulse" />
                                            </div>
                                        )}
                                    </div>
                                </div>
                            </div>
                        </div>
                    ) : (
                        <div className="h-[60vh] flex flex-col items-center justify-center text-center space-y-6 cyber-card border-dashed">
                            <div className="w-20 h-20 rounded-3xl bg-white/5 flex items-center justify-center text-gray-700 border border-white/5">
                                <Activity size={40} />
                            </div>
                            <div className="space-y-2">
                                <h3 className="text-xl font-black text-white uppercase tracking-tighter">Diagnostic Hold</h3>
                                <p className="text-[10px] font-mono text-gray-500 uppercase tracking-widest max-w-xs leading-relaxed">Select a sequence artifact from the artifact index to begin deep-state inspection.</p>
                            </div>
                        </div>
                    )}
                </div>
            </div>

            {/* Deep Scan remediation Modal */}
            {showVulnModal && (
                <div className="fixed inset-0 z-[200] flex items-center justify-center bg-black/95 p-6 backdrop-blur-xl transition-all animate-in fade-in zoom-in-95 duration-300">
                    <div className="cyber-card max-w-6xl w-full h-[85vh] flex flex-col relative overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-red-600 via-orange-600 to-transparent" />

                        <div className="flex items-center justify-between p-8 border-b border-white/5 bg-white/[0.02]">
                            <div className="flex items-center gap-6">
                                <div className="w-14 h-14 rounded-2xl bg-red-600/10 flex items-center justify-center text-red-500 border border-red-500/20 shadow-[0_0_30px_rgba(220,38,38,0.15)]">
                                    <ShieldAlert size={32} />
                                </div>
                                <div>
                                    <h2 className="text-3xl font-black text-white uppercase tracking-tighter">Remediation Control</h2>
                                    <p className="text-sm font-mono text-gray-500 uppercase tracking-widest">Deep-Scan Analysis: <span className="text-red-500">{name}:{selectedTag}</span></p>
                                </div>
                            </div>
                            <button onClick={() => setShowVulnModal(false)} className="w-12 h-12 rounded-xl hover:bg-white/5 flex items-center justify-center text-gray-500 hover:text-white transition-all text-2xl group"><X size={28} className="group-hover:rotate-90 transition-transform duration-300" /></button>
                        </div>

                        <div className="flex-1 overflow-auto p-8 custom-scrollbar bg-black/40">
                            {isLoadingReport ? (
                                <div className="h-full flex flex-col items-center justify-center space-y-6">
                                    <div className="w-16 h-16 border-4 border-red-500/20 border-t-red-500 rounded-full animate-spin" />
                                    <span className="text-sm font-mono text-red-400 uppercase tracking-widest animate-pulse font-black">Decrypting Threat Data...</span>
                                </div>
                            ) : vulnReport ? (
                                <VulnerabilityMatrix report={vulnReport} />
                            ) : (
                                <div className="text-center font-mono text-gray-600 uppercase py-20 tracking-widest">CRITICAL_ERROR: TELEMETRY_LOST</div>
                            )}
                        </div>

                        <div className="p-6 bg-white/[0.02] border-t border-white/5 flex justify-end">
                            <button onClick={() => setShowVulnModal(false)} className="px-10 py-4 bg-white/5 text-gray-400 rounded-2xl hover:bg-white/10 transition font-black uppercase text-[10px] tracking-widest">Close Deep-Scan</button>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}

interface StatBoxProps {
    icon: React.ElementType;
    label: string;
    value: string | number;
    fullValue?: string;
    accent?: 'green' | 'red' | 'blue' | 'white';
}

const StatBox = ({ icon: Icon, label, value, fullValue, accent = "white" }: StatBoxProps) => (
    <div className="cyber-card p-6 group transition-all hover:bg-white/[0.02]">
        <div className="flex items-center gap-3 mb-4">
            <Icon size={14} className={clsx(
                accent === 'green' ? "text-green-500" : accent === 'red' ? "text-red-500" : accent === 'blue' ? "text-blue-500" : "text-gray-600"
            )} />
            <span className="text-xs font-black text-gray-600 uppercase tracking-widest group-hover:text-gray-400 transition-colors">{label}</span>
        </div>
        <div className={clsx(
            "text-lg font-black uppercase tracking-tight truncate",
            accent === 'green' ? "text-green-400" : accent === 'red' ? "text-red-400" : accent === 'blue' ? "text-blue-400" : "text-white"
        )} title={fullValue}>{value}</div>
    </div>
);

interface VulnBoxProps {
    label: string;
    count: number;
    color: 'red' | 'orange' | 'yellow' | 'blue';
}

const VulnBox = ({ label, count, color }: VulnBoxProps) => {
    const colors = {
        red: "bg-red-600/10 text-red-500 border-red-500/20",
        orange: "bg-orange-600/10 text-orange-500 border-orange-500/20",
        yellow: "bg-yellow-600/10 text-yellow-500 border-yellow-500/20",
        blue: "bg-blue-600/10 text-blue-500 border-blue-500/20"
    };
    return (
        <div className={clsx("p-4 rounded-xl border flex flex-col items-center justify-center transition-all hover:scale-105", (colors as any)[color])}>
            <span className="text-xl font-black leading-none mb-1">{count}</span>
            <span className="text-[8px] font-black uppercase tracking-widest opacity-70">{label}</span>
        </div>
    );
};

function ScanStatusPanel({ status }: { status: ScanStatus }) {
    const configs = {
        pending: { color: 'text-gray-500', icon: Clock, label: 'STANDBY_MODE', bg: 'bg-white/5 border-white/10' },
        scanning: { color: 'text-blue-400', icon: RefreshCw, label: 'SCAN_IN_PROGRESS', bg: 'bg-blue-400/5 border-blue-500/20' },
        completed: { color: 'text-green-500', icon: CheckCircle, label: 'PROTOCOL_COMPLETE', bg: 'bg-green-500/5 border-green-500/20' },
        failed: { color: 'text-red-500', icon: XCircle, label: 'INTEGRITY_FAILURE', bg: 'bg-red-500/5 border-red-500/20' },
    };

    const config = configs[status.status] || configs.pending;
    const Icon = config.icon;

    return (
        <div className={clsx("p-6 rounded-2xl border transition-all duration-500", config.bg)}>
            <div className="flex items-center gap-4 mb-4">
                <Icon size={24} className={clsx(config.color, status.status === 'scanning' && "animate-spin")} />
                <div className="flex-1">
                    <div className={clsx("text-xs font-black uppercase tracking-[0.2em]", config.color)}>{config.label}</div>
                    {status.scanned_at && status.status !== 'scanning' && (
                        <div className="text-[10px] text-gray-500 font-mono mt-1">LAST_SYNC: {new Date(status.scanned_at).toLocaleString().toUpperCase()}</div>
                    )}
                </div>
            </div>

            {status.status === 'scanning' && (
                <ScanningProgressUI message={status.progress_message} />
            )}

            {status.error && (
                <div className="pt-4 mt-4 border-t border-red-500/10 text-[10px] font-mono text-red-500 uppercase tracking-widest">
                    CRITICAL_ERROR: {status.error}
                </div>
            )}
        </div>
    );
}

function ScanningProgressUI({ message }: { message?: string }) {
    const [progress, setProgress] = useState(10);
    const [displayMsg, setDisplayMsg] = useState(message || 'Initializing Protocol...');

    useEffect(() => {
        const interval = setInterval(() => {
            setProgress(prev => (prev < 90 ? prev + Math.random() * 15 : prev));
        }, 800);
        return () => clearInterval(interval);
    }, []);

    return (
        <div className="space-y-3 mt-2 animate-in fade-in duration-300">
            <div className="flex justify-between text-[10px] font-mono font-black uppercase tracking-widest">
                <span className="text-blue-400 animate-pulse">{displayMsg}</span>
                <span className="text-white">{Math.round(progress)}%</span>
            </div>
            <div className="w-full bg-blue-900/30 rounded-full h-1 border border-blue-500/20 overflow-hidden">
                <div
                    className="bg-blue-500 h-full transition-all duration-500 ease-out shadow-[0_0_10px_rgba(59,130,246,0.6)]"
                    style={{ width: `${progress}%` }}
                />
            </div>
        </div>
    );
}

function ScanStatusIconInline({ status }: { status: string }) {
    switch (status) {
        case 'completed': return <CheckCircle size={10} className="text-green-500" />;
        case 'scanning': return <RefreshCw size={10} className="text-blue-400 animate-spin" />;
        case 'failed': return <XCircle size={10} className="text-red-500" />;
        default: return <Clock size={10} className="text-gray-600" />;
    }
}

function VulnerabilityMatrix({ report }: { report: TrivyReport }) {
    const vulnerabilities = report.Results?.flatMap((res: TrivyResult) => res.Vulnerabilities || []) || [];

    if (vulnerabilities.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-full py-20 animate-in zoom-in-95 duration-700">
                <div className="w-24 h-24 rounded-full bg-green-500/10 flex items-center justify-center text-green-500 mb-6 border border-green-500/20 shadow-[0_0_50px_rgba(34,197,94,0.15)]">
                    <CheckCircle size={48} />
                </div>
                <h3 className="text-2xl font-black text-white uppercase tracking-tighter mb-2">Integrity Verified</h3>
                <p className="text-[10px] font-mono text-gray-500 uppercase tracking-widest">Zero threat vectors detected in current artifact sequence.</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            <div className="overflow-x-auto rounded-2xl border border-white/5 bg-black/50">
                <table className="w-full text-left font-mono">
                    <thead>
                        <tr className="bg-white/[0.03] text-gray-500 text-sm font-black uppercase tracking-[0.1em]">
                            <th className="px-8 py-6">Threat_Level</th>
                            <th className="px-6 py-6">Entity_Package</th>
                            <th className="px-6 py-6">Current_State</th>
                            <th className="px-6 py-6">Remediation_Target</th>
                            <th className="px-8 py-6 text-right">Ident_Hash</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-white/5">
                        {vulnerabilities.map((vuln: TrivyVulnerability, idx: number) => {
                            const colors: Record<string, string> = {
                                'CRITICAL': 'text-red-500 font-black',
                                'HIGH': 'text-orange-500 font-black',
                                'MEDIUM': 'text-yellow-500 font-bold',
                                'LOW': 'text-blue-500',
                            };

                            return (
                                <tr key={idx} className="hover:bg-white/[0.02] transition-colors group">
                                    <td className="px-8 py-6">
                                        <div className={clsx("flex items-center gap-2 text-sm font-bold", colors[vuln.Severity] || 'text-gray-500')}>
                                            <div className={clsx("w-2 h-2 rounded-full", (colors[vuln.Severity] || '').includes('red') ? "bg-red-500 shadow-[0_0_8px_rgba(239,68,68,0.5)]" : "bg-current")} />
                                            {vuln.Severity}
                                        </div>
                                    </td>
                                    <td className="px-6 py-6 text-white text-base font-bold">{vuln.PkgName}</td>
                                    <td className="px-6 py-6 text-gray-400 text-sm">{vuln.InstalledVersion}</td>
                                    <td className="px-6 py-6">
                                        {vuln.FixedVersion ? (
                                            <span className="text-green-400 text-sm bg-green-500/10 px-3 py-1.5 rounded-lg border border-green-500/20 font-black">
                                                FIX_AVAIL: {vuln.FixedVersion}
                                            </span>
                                        ) : (
                                            <span className="text-gray-700 text-xs uppercase font-black">UNSTABLE_STATE</span>
                                        )}
                                    </td>
                                    <td className="px-8 py-6 text-right">
                                        <a
                                            href={vuln.PrimaryURL || `https://avd.aquasec.com/nvd/${vuln.VulnerabilityID}`}
                                            target="_blank"
                                            rel="noreferrer"
                                            className="text-blue-500 hover:text-blue-400 transition-colors text-sm font-black underline decoration-blue-500/30 underline-offset-4"
                                        >
                                            {vuln.VulnerabilityID}
                                        </a>
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </div>
        </div>
    );
}

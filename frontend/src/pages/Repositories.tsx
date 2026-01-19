import React, { useState, useEffect } from 'react';
import { Search, Filter, Box, Tag, Plus, Terminal, Copy, Check, Trash2, ArrowRight, ShieldCheck, Database, LayoutGrid, List as ListIcon } from 'lucide-react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { registry } from '../lib/api';
import { Link, useSearchParams } from 'react-router-dom';
import clsx from 'clsx';

import { Modal } from '../components/Modal';

const Repositories = () => {
    const [searchParams] = useSearchParams();
    const [searchTerm, setSearchTerm] = useState('');
    const [createRepoName, setCreateRepoName] = useState('');
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
    const [showSearchDropdown, setShowSearchDropdown] = useState(false);
    const searchInputRef = React.useRef<HTMLInputElement>(null);
    const [repoToDelete, setRepoToDelete] = useState<string | null>(null);
    const [alertMessage, setAlertMessage] = useState<string | null>(null);

    const queryClient = useQueryClient();

    // Initialize search term from URL params
    useEffect(() => {
        const query = searchParams.get('search');
        if (query) {
            setSearchTerm(query);
        }
    }, [searchParams]);

    // Close dropdown when clicking outside
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (searchInputRef.current && !searchInputRef.current.contains(event.target as Node)) {
                setShowSearchDropdown(false);
            }
        };

        if (showSearchDropdown) {
            document.addEventListener('mousedown', handleClickOutside);
            return () => document.removeEventListener('mousedown', handleClickOutside);
        }
    }, [showSearchDropdown]);

    const { data: catalogData, isLoading: isCatalogLoading } = useQuery({
        queryKey: ['catalog'],
        queryFn: registry.getCatalog
    });

    const repoList = catalogData?.data?.repositories || [];
    const filteredRepos = repoList.filter((name: string) => name.toLowerCase().includes(searchTerm.toLowerCase()));

    const handleCopy = (text: string) => {
        navigator.clipboard.writeText(text);
    };

    const confirmDeleteRepo = async () => {
        if (!repoToDelete) return;
        try {
            await registry.deleteRepository(repoToDelete);
            await queryClient.invalidateQueries({ queryKey: ['catalog'] });
            setRepoToDelete(null);
        } catch (err: any) {
            setRepoToDelete(null);
            setAlertMessage(`PURGE_FAILURE: ${err.message}`);
        }
    };

    return (
        <div className="space-y-10 pb-20">
            {/* Header / Command Bar - No Changes needed, just context */}
            <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
                {/* ... (Kept Header Code same as existing, just collapsed for brevity in tool input if strict check not passed, but I must preserve context if I don't use multi_replace. I will use multi_replace for specific blocks to be safe, but here I am rewriting the top and passing props.) */}
                {/* Actually, I will just insert the state and Modals, and update the map function */}
                <div>
                    <h1 className="text-4xl font-black uppercase tracking-tighter text-white">Registry Vault</h1>
                    <p className="text-blue-400 font-mono text-xs tracking-[0.3em] uppercase opacity-70">Secured OCI Image Repository Catalog</p>
                </div>

                <div className="flex flex-wrap items-center gap-4">
                    {/* ... Search Bar ... */}
                    <div className="relative group min-w-[300px]">
                        <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500 group-focus-within:text-blue-400 transition-colors z-10" />
                        <input
                            ref={searchInputRef}
                            type="text"
                            placeholder="SEARCH VAULT..."
                            className="w-full bg-white/5 border border-white/5 rounded-2xl py-3 pl-12 pr-6 text-xs font-mono uppercase tracking-widest placeholder:text-gray-700 focus:outline-none focus:border-blue-500/30 focus:bg-white/10 transition-all font-bold"
                            value={searchTerm}
                            onChange={(e) => {
                                setSearchTerm(e.target.value);
                                setShowSearchDropdown(e.target.value.length > 0);
                            }}
                            onFocus={() => setShowSearchDropdown(searchTerm.length > 0)}
                            onKeyDown={(e) => {
                                if (e.key === 'Escape') {
                                    setShowSearchDropdown(false);
                                    searchInputRef.current?.blur();
                                }
                            }}
                        />
                        {/* ... */}
                        {showSearchDropdown && searchTerm && filteredRepos.length > 0 && (
                            <div className="absolute top-full mt-2 w-full bg-black/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-2xl shadow-blue-500/10 overflow-hidden z-50 max-h-96 overflow-y-auto">
                                <div className="px-4 py-3 border-b border-white/5 bg-white/5">
                                    <span className="text-xs font-black text-gray-400 uppercase tracking-widest">
                                        {filteredRepos.length} {filteredRepos.length === 1 ? 'Repository' : 'Repositories'} Found
                                    </span>
                                </div>
                                {filteredRepos.slice(0, 10).map((repo: string) => (
                                    <Link
                                        key={repo}
                                        to={`/repositories/${repo}`}
                                        onClick={() => {
                                            setShowSearchDropdown(false);
                                            setSearchTerm('');
                                        }}
                                        className="flex items-center gap-3 px-4 py-3 hover:bg-white/5 transition-colors border-b border-white/5 last:border-0 group"
                                    >
                                        <Box size={16} className="text-blue-400 group-hover:text-blue-300" />
                                        <span className="font-mono text-sm text-white group-hover:text-blue-300 transition-colors">
                                            {repo}
                                        </span>
                                    </Link>
                                ))}
                                {filteredRepos.length > 10 && (
                                    <div className="px-4 py-3 bg-white/5 border-t border-white/10">
                                        <span className="text-xs font-bold text-gray-400 uppercase tracking-widest">
                                            Scroll down to see all {filteredRepos.length} results
                                        </span>
                                    </div>
                                )}
                            </div>
                        )}
                    </div>

                    <div className="flex items-center gap-1 bg-white/5 p-1 rounded-xl border border-white/5">
                        <button
                            onClick={() => setViewMode('grid')}
                            className={clsx("p-2 rounded-lg transition-all", viewMode === 'grid' ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "text-gray-500 hover:text-white")}
                        >
                            <LayoutGrid size={16} />
                        </button>
                        <button
                            onClick={() => setViewMode('list')}
                            className={clsx("p-2 rounded-lg transition-all", viewMode === 'list' ? "bg-blue-600 text-white shadow-lg shadow-blue-500/20" : "text-gray-500 hover:text-white")}
                        >
                            <ListIcon size={16} />
                        </button>
                    </div>

                    <button
                        onClick={() => setIsCreateOpen(true)}
                        className="flex items-center gap-3 bg-blue-600 hover:bg-blue-500 text-white px-6 py-3 rounded-2xl font-black uppercase text-[10px] tracking-widest transition-all shadow-[0_0_25px_rgba(37,99,235,0.3)] active:scale-95"
                    >
                        <Plus size={16} />
                        Initiate Repository
                    </button>
                </div>
            </div>

            {/* Content Area */}
            {isCatalogLoading ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
                    {[1, 2, 3].map(i => <div key={i} className="h-48 cyber-card animate-pulse" />)}
                </div>
            ) : filteredRepos.length === 0 ? (
                <div className="cyber-card p-20 flex flex-col items-center justify-center text-center">
                    <Database size={64} className="text-gray-700 mb-6" />
                    <h2 className="text-2xl font-black uppercase text-white mb-2 tracking-tighter">Vault Segment Empty</h2>
                    <p className="text-xs font-mono text-gray-500 uppercase tracking-widest">No matching repository entities found in index.</p>
                </div>
            ) : (
                <div className={clsx(
                    viewMode === 'grid' ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-8" : "flex flex-col gap-4"
                )}>
                    {filteredRepos.map((name: string, idx: number) => (
                        <RepositoryComponent
                            key={name}
                            name={name}
                            onCopy={handleCopy}
                            viewMode={viewMode}
                            onRequestDelete={setRepoToDelete}
                        />
                    ))}
                </div>
            )}

            {/* Create Repositories Modal */}
            {isCreateOpen && (
                <div className="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-6 backdrop-blur-md">
                    <div className="cyber-card max-w-2xl w-full p-8 relative overflow-hidden">
                        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-blue-600 to-transparent" />

                        <div className="flex items-center justify-between mb-8">
                            <div className="flex items-center gap-4">
                                <div className="w-12 h-12 rounded-xl bg-blue-600/10 flex items-center justify-center text-blue-400">
                                    <Terminal size={24} />
                                </div>
                                <h2 className="text-2xl font-black text-white uppercase tracking-tighter">Create Repository</h2>
                            </div>
                            <button onClick={() => setIsCreateOpen(false)} className="w-10 h-10 rounded-xl hover:bg-white/5 flex items-center justify-center text-gray-500 hover:text-white transition-all text-xl">✕</button>
                        </div>

                        <div className="space-y-8">
                            <div>
                                <label className="text-[10px] font-black text-gray-500 uppercase tracking-widest mb-3 block">Logical Identity</label>
                                <div className="relative group">
                                    <div className="absolute left-5 top-1/2 -translate-y-1/2 text-gray-600 font-mono text-xs uppercase select-none group-focus-within:text-blue-500 transition-colors tracking-tighter">registry.local:5000 /</div>
                                    <input
                                        type="text"
                                        placeholder="NAMESPACE / APP_NAME"
                                        className="w-full bg-black/40 border border-white/5 rounded-2xl py-4 pl-40 pr-6 text-white text-xs font-mono font-bold uppercase tracking-widest focus:outline-none focus:border-blue-500/30 focus:bg-white/[0.02] transition-all shadow-inner"
                                        value={createRepoName}
                                        onChange={(e) => setCreateRepoName(e.target.value)}
                                    />
                                </div>
                            </div>

                            <div className="space-y-4">
                                <p className="text-[10px] font-black text-gray-500 uppercase tracking-widest">Initialization Instructions</p>
                                <div className="space-y-2 font-mono text-[10px] uppercase">
                                    <CommandBlock cmd="docker login localhost:5000" label="AUTH" />
                                    <CommandBlock cmd={`docker tag app:latest localhost:5000/${createRepoName || 'project/image'}:latest`} label="TAG" />
                                    <CommandBlock cmd={`docker push localhost:5000/${createRepoName || 'project/image'}:latest`} label="PUSH" />
                                </div>
                            </div>
                        </div>

                        <div className="mt-10 flex gap-4">
                            <button
                                onClick={() => setIsCreateOpen(false)}
                                className="flex-1 px-8 py-4 bg-white/5 text-gray-400 rounded-2xl hover:bg-white/10 transition font-black uppercase text-[10px] tracking-widest"
                            >
                                Discard
                            </button>
                            <button
                                onClick={async () => {
                                    if (!createRepoName) return;
                                    try {
                                        await registry.createRepository(createRepoName);
                                        await queryClient.invalidateQueries({ queryKey: ['catalog'] });
                                        setIsCreateOpen(false);
                                        setCreateRepoName('');
                                    } catch (e) {
                                        console.error("Failed to create repo", e);
                                        setAlertMessage("CRITICAL ERROR: REPOSITORY_CREATION_FAILED");
                                    }
                                }}
                                className="flex-1 px-8 py-4 bg-blue-600 text-white rounded-2xl hover:bg-blue-500 transition font-black uppercase text-[10px] tracking-widest shadow-[0_0_20px_rgba(37,99,235,0.3)]"
                            >
                                Confirm Allocation
                            </button>
                        </div>
                    </div>
                </div>
            )}

            {/* Delete Confirmation Modal */}
            <Modal
                isOpen={!!repoToDelete}
                onClose={() => setRepoToDelete(null)}
                title="SECURE PURGE: DESTROY REPOSITORY?"
                variant="danger"
                footer={
                    <>
                        <button onClick={() => setRepoToDelete(null)} className="px-4 py-2 text-gray-400 hover:text-white transition font-bold uppercase text-xs tracking-wider">Cancel</button>
                        <button
                            onClick={confirmDeleteRepo}
                            className="px-6 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-black uppercase tracking-wider transition shadow-lg shadow-red-600/20"
                        >
                            Destroy Repository
                        </button>
                    </>
                }
            >
                Confirm deletion of <span className="text-white font-bold">{repoToDelete}</span>? This action is irreversible.
            </Modal>

            {/* Alert Modal */}
            <Modal
                isOpen={!!alertMessage}
                onClose={() => setAlertMessage(null)}
                title="System Alert"
                variant="danger"
                footer={
                    <button onClick={() => setAlertMessage(null)} className="px-6 py-2 bg-white/10 hover:bg-white/20 text-white rounded-lg font-bold uppercase text-xs tracking-wider transition">
                        Acknowledge
                    </button>
                }
            >
                {alertMessage}
            </Modal>
        </div>
    );
};

// ... RepositoryComponent changes ....
const RepositoryComponent = ({ name, onCopy, viewMode, onRequestDelete }: { name: string, onCopy: (txt: string) => void, viewMode: 'grid' | 'list', onRequestDelete: (name: string) => void }) => {
    // ... hooks same ...
    const { data: tagData } = useQuery({
        queryKey: ['tags', name],
        queryFn: () => registry.getTags(name)
    });

    const tags = tagData?.data?.tags || [];
    const displayTags = tags.filter((t: string) => !t.endsWith('.sig'));
    const latestTag = displayTags.length > 0 ? displayTags[0] : 'latest';
    const pullCommand = `docker pull localhost:5000/${name}:${latestTag}`;
    const [copied, setCopied] = useState(false);

    // Check if I need queryClient here? No, deletions handled by parent.
    // Actually handleCopyCmd needs logic
    const handleCopyCmd = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        onCopy(pullCommand);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    const handleDeleteRepo = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        onRequestDelete(name);
    };

    if (viewMode === 'list') {
        return (
            <div className="cyber-card p-6 flex items-center justify-between group hover:border-blue-500/30 transition-all cursor-pointer">
                <Link to={`/repositories/${encodeURIComponent(name)}`} className="flex flex-1 items-center gap-6">
                    <div className="w-12 h-12 rounded-xl bg-blue-500/10 flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform">
                        <Box size={24} />
                    </div>
                    <div className="flex-1">
                        <h3 className="text-lg font-black uppercase tracking-tight text-white mb-1 group-hover:text-blue-400 transition-colors">{name}</h3>
                        <div className="flex items-center gap-4 text-[10px] font-mono uppercase text-gray-500">
                            <span className="flex items-center gap-1"><Tag size={10} /> {displayTags.length} TAGS</span>
                            <span className="flex items-center gap-1 text-green-500"><ShieldCheck size={10} /> VERIFIED</span>
                        </div>
                    </div>
                </Link>
                <div className="flex items-center gap-3">
                    <div className="relative group/cmd">
                        <button
                            onClick={handleCopyCmd}
                            className="p-3 rounded-xl bg-white/5 text-gray-500 hover:text-white transition-all hover:bg-blue-500/10 hover:border-blue-500/30 border border-transparent"
                        >
                            {copied ? <Check size={16} className="text-green-500" /> : <Copy size={16} />}
                        </button>
                        <span className="absolute bottom-full right-0 mb-2 whitespace-nowrap bg-black text-blue-400 text-[10px] font-mono px-2 py-1 rounded border border-blue-500/20 opacity-0 group-hover/cmd:opacity-100 transition-opacity">
                            {pullCommand}
                        </span>
                    </div>
                    <button
                        onClick={handleDeleteRepo}
                        className="p-3 rounded-xl bg-white/5 text-gray-500 hover:text-red-500 transition-all hover:bg-red-500/10 border border-transparent"
                    >
                        <Trash2 size={16} />
                    </button>
                    <Link to={`/repositories/${encodeURIComponent(name)}`} className="p-3 rounded-xl bg-blue-600 text-white hover:bg-blue-500 transition-all shadow-lg shadow-blue-500/20">
                        <ArrowRight size={16} />
                    </Link>
                </div>
            </div>
        );
    }

    return (
        <div className="cyber-card p-8 group flex flex-col relative overflow-hidden h-full">
            <div className="absolute top-0 right-0 w-32 h-32 bg-blue-600/5 blur-3xl pointer-events-none group-hover:bg-blue-600/10 transition-colors" />

            <Link to={`/repositories/${encodeURIComponent(name)}`} className="flex-1">
                <div className="flex items-start justify-between mb-6">
                    <div className="w-14 h-14 rounded-2xl bg-white/[0.03] border border-white/5 flex items-center justify-center text-blue-400 group-hover:scale-110 group-hover:shadow-[0_0_20px_rgba(59,130,246,0.3)] transition-all">
                        <Box size={28} />
                    </div>
                    <div className="flex flex-col items-end gap-2">
                        <span className="text-[10px] font-mono text-green-500 flex items-center gap-1 animate-pulse">
                            <span className="w-1.5 h-1.5 rounded-full bg-green-500" /> ACTIVE
                        </span>
                        <div className="flex items-center gap-1 px-2 py-1 bg-white/5 rounded-lg text-[10px] font-mono text-gray-500">
                            <Tag size={10} /> {displayTags.length}
                        </div>
                    </div>
                </div>

                <h3 className="text-xl font-black uppercase tracking-tight text-white mb-4 group-hover:text-blue-400 transition-colors line-clamp-1">
                    {name}
                </h3>
            </Link>

            <div className="mt-2 flex flex-col gap-3 pt-6 border-t border-white/5">
                <div className="flex items-center gap-2">
                    <button
                        onClick={handleCopyCmd}
                        className="flex-1 flex items-center justify-center gap-2 py-3 bg-white/5 text-gray-400 rounded-xl hover:bg-blue-500/10 hover:text-blue-400 border border-transparent hover:border-blue-500/30 transition-all text-[10px] font-black uppercase tracking-widest"
                    >
                        {copied ? <Check size={14} className="text-green-500" /> : <Copy size={14} />}
                        {copied ? "ALLOCATED" : "PULL_HASH"}
                    </button>
                    <button
                        onClick={handleDeleteRepo}
                        className="p-3 bg-white/5 text-gray-600 rounded-xl hover:bg-red-500/10 hover:text-red-500 transition-all"
                    >
                        <Trash2 size={16} />
                    </button>
                </div>
                <Link
                    to={`/repositories/${encodeURIComponent(name)}`}
                    className="flex items-center justify-center gap-2 py-3.5 bg-blue-600 text-white rounded-xl hover:bg-blue-500 transition-all text-[10px] font-black uppercase tracking-widest shadow-lg shadow-blue-500/10 group-hover:shadow-blue-500/20 active:scale-95"
                >
                    View Repository Details
                    <ArrowRight size={14} />
                </Link>
            </div>
        </div>
    );
};

const CommandBlock = ({ cmd, label }: { cmd: string, label: string }) => {
    const [copied, setCopied] = useState(false);
    const handleCopy = () => {
        navigator.clipboard.writeText(cmd);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
    };

    return (
        <div className="group flex items-center bg-black/50 border border-white/5 rounded-xl hover:border-blue-500/30 transition-all px-4 py-3 gap-4">
            <span className="text-[9px] font-black text-blue-500/50 group-hover:text-blue-500 w-8">{label}</span>
            <code className="flex-1 text-gray-400 group-hover:text-white transition-colors truncate">{cmd}</code>
            <button
                onClick={handleCopy}
                className="text-gray-600 hover:text-white transition-colors ml-auto"
            >
                {copied ? <Check size={14} className="text-green-500" /> : <Copy size={14} />}
            </button>
        </div>
    );
};

export default Repositories;

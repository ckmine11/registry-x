import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Library, LogOut, Search, User, Shield, Settings, GitBranch, DollarSign, Activity, Bell } from 'lucide-react';
import { useAuth } from '../lib/auth-context';
import { Outlet } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { registry } from '../lib/api';

export default function Layout() {
    const { logout, user } = useAuth();
    const location = useLocation();
    const navigate = useNavigate();
    const [enableCostIntelligence, setEnableCostIntelligence] = React.useState(true);
    const [isSidebarCollapsed, setIsSidebarCollapsed] = React.useState(false);
    const [searchQuery, setSearchQuery] = React.useState('');
    const [showSearchResults, setShowSearchResults] = React.useState(false);
    const [showNotifications, setShowNotifications] = React.useState(false);
    const [showActivity, setShowActivity] = React.useState(false);
    const [notifications, setNotifications] = React.useState<{ title: string; desc: string; time: string; type: string }[]>([]);
    const searchInputRef = React.useRef<HTMLInputElement>(null);
    const notificationsRef = React.useRef<HTMLDivElement>(null);
    const activityRef = React.useRef<HTMLDivElement>(null);

    // Fetch repositories for search
    const { data: catalogData } = useQuery({
        queryKey: ['catalog'],
        queryFn: registry.getCatalog,
        enabled: true
    });

    React.useEffect(() => {
        import('../lib/api').then(({ api }) => {
            api.getSystemConfig().then(res => {
                setEnableCostIntelligence(res.data.enableCostIntelligence);
            }).catch(err => console.error("Failed to fetch system config", err));
        });
    }, []);

    // Close dropdowns when clicking outside
    React.useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            const target = event.target as Node;

            // Search Results
            if (searchInputRef.current && !searchInputRef.current.contains(target)) {
                setShowSearchResults(false);
            }

            // Notifications
            if (notificationsRef.current && !notificationsRef.current.contains(target)) {
                setShowNotifications(false);
            }

            // Activity
            if (activityRef.current && !activityRef.current.contains(target)) {
                setShowActivity(false);
            }
        };

        if (showSearchResults || showNotifications || showActivity) {
            document.addEventListener('click', handleClickOutside);
            return () => document.removeEventListener('click', handleClickOutside);
        }
    }, [showSearchResults, showNotifications, showActivity]);

    return (
        <div className="flex min-h-screen text-white font-sans antialiased selection:bg-blue-500/30 overflow-hidden relative">
            {/* Modern Gradient Background */}
            <div className="fixed inset-0 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 -z-10" />

            {/* Animated Mesh Pattern */}
            <div className="fixed inset-0 opacity-[0.03] -z-10"
                style={{
                    backgroundImage: `radial-gradient(circle at 1px 1px, rgb(148, 163, 184) 1px, transparent 0)`,
                    backgroundSize: '40px 40px'
                }}
            />

            {/* Subtle Gradient Orbs */}
            <div className="fixed top-0 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-[100px] -z-10 animate-pulse" style={{ animationDuration: '8s' }} />
            <div className="fixed bottom-0 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-[100px] -z-10 animate-pulse" style={{ animationDuration: '10s' }} />
            <div className="fixed top-1/2 left-1/2 w-96 h-96 bg-cyan-500/5 rounded-full blur-[120px] -z-10 animate-pulse" style={{ animationDuration: '12s' }} />

            {/* Sidebar */}
            <aside className={`border-r border-white/5 flex flex-col fixed h-full transition-all duration-500 ease-in-out z-20 ${isSidebarCollapsed ? "w-20" : "w-72"
                } bg-black/40 backdrop-blur-2xl`}>
                <div className="p-8 pb-4 flex items-center justify-between">
                    <div className="flex items-center gap-4 transition-all duration-300">
                        <div className="w-10 h-10 rounded-2xl bg-gradient-to-tr from-blue-600 via-blue-400 to-cyan-300 flex items-center justify-center font-black text-2xl shadow-[0_0_20px_rgba(59,130,246,0.5)] animate-pulse-glow">
                            X
                        </div>
                        {!isSidebarCollapsed && (
                            <div className="flex flex-col">
                                <span className="font-black text-xl tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-white to-gray-400 uppercase">RegistryX</span>
                                <span className="text-[10px] text-blue-400 font-mono tracking-widest uppercase opacity-70">Container Registry</span>
                            </div>
                        )}
                    </div>
                </div>

                <div className="px-4 py-8 flex-1 space-y-6 overflow-y-auto overflow-x-hidden">
                    <div>
                        {!isSidebarCollapsed && (
                            <h3 className="px-4 text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em] mb-4">Main</h3>
                        )}
                        <nav className="space-y-1">
                            <NavLink to="/dashboard" icon={<LayoutDashboard />} label="Dashboard" collapsed={isSidebarCollapsed} />
                            <NavLink to="/repositories" icon={<Library />} label="Repositories" collapsed={isSidebarCollapsed} />
                            <NavLink to="/lineage" icon={<GitBranch />} label="Lineage" collapsed={isSidebarCollapsed} />
                            {enableCostIntelligence && (
                                <NavLink to="/costs" icon={<DollarSign />} label="Cost Intelligence" collapsed={isSidebarCollapsed} />
                            )}
                        </nav>
                    </div>

                    <div>
                        {!isSidebarCollapsed && (
                            <h3 className="px-4 text-[10px] font-bold text-gray-500 uppercase tracking-[0.2em] mb-4">Settings</h3>
                        )}
                        <nav className="space-y-1">
                            <NavLink to="/policies" icon={<Shield />} label="Policies" collapsed={isSidebarCollapsed} />
                            {user?.role === 'admin' && (
                                <NavLink to="/sessions" icon={<Activity />} label="Sessions" collapsed={isSidebarCollapsed} />
                            )}
                            <NavLink to="/settings" icon={<Settings />} label="Settings" collapsed={isSidebarCollapsed} />
                        </nav>
                    </div>
                </div>

                <div className="p-4 border-t border-white/5 space-y-4">
                    {!isSidebarCollapsed && (
                        <div className="bg-white/5 rounded-2xl p-4 border border-white/5 flex items-center gap-3 group hover:border-blue-500/30 transition-all cursor-pointer overflow-hidden">
                            <div className="w-10 h-10 rounded-xl bg-blue-500/20 flex items-center justify-center text-blue-400 group-hover:scale-110 transition-transform">
                                <User size={20} />
                            </div>
                            <div className="flex-1 min-w-0">
                                <div className="text-sm font-bold truncate text-white uppercase tracking-tight">{user?.username || 'User'}</div>
                                <div className="text-[10px] text-green-400 font-mono items-center flex gap-1 animate-pulse">
                                    <span className="w-1.5 h-1.5 rounded-full bg-green-400" />
                                    ONLINE
                                </div>
                            </div>
                        </div>
                    )}

                    <button
                        onClick={logout}
                        className={`flex items-center gap-4 px-4 py-3 text-gray-400 hover:text-red-400 hover:bg-red-500/10 rounded-2xl w-full transition-all duration-300 group font-bold tracking-tight uppercase text-xs ${isSidebarCollapsed ? "justify-center" : ""
                            }`}
                    >
                        <LogOut size={20} />
                        {!isSidebarCollapsed && <span>Logout</span>}
                    </button>

                    <button
                        onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
                        className="w-full flex justify-center py-2 text-gray-600 hover:text-white transition-colors"
                    >
                        <div className="w-8 h-1 rounded-full bg-gray-800" />
                    </button>
                </div>
            </aside>

            {/* Main Content Area */}
            <div className={`flex-1 flex flex-col min-h-screen transition-all duration-500 ${isSidebarCollapsed ? "ml-20" : "ml-72"
                }`}>
                {/* Header */}
                <header className="h-20 flex items-center justify-between px-10 bg-black/40 backdrop-blur-xl border-b border-white/5 sticky top-0 z-10">
                    <div className="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-blue-500/50 to-transparent" />

                    <div className="flex items-center gap-8 flex-1">
                        <div className="relative group w-full max-w-xl">
                            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500 group-focus-within:text-blue-400 transition-colors z-10" />
                            <input
                                ref={searchInputRef}
                                type="text"
                                placeholder="Search repositories, images, CVEs..."
                                className="w-full bg-white/5 border border-white/5 rounded-2xl py-3 pl-12 pr-6 text-xs font-mono uppercase tracking-widest placeholder:text-gray-700 focus:outline-none focus:border-blue-500/30 focus:bg-white/10 transition-all shadow-inner"
                                value={searchQuery}
                                onChange={(e) => {
                                    setSearchQuery(e.target.value);
                                    setShowSearchResults(e.target.value.length > 0);
                                }}
                                onFocus={() => setShowSearchResults(searchQuery.length > 0)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && searchQuery.trim()) {
                                        navigate(`/repositories?search=${encodeURIComponent(searchQuery)}`);
                                        setShowSearchResults(false);
                                        searchInputRef.current?.blur();
                                    }
                                    if (e.key === 'Escape') {
                                        setShowSearchResults(false);
                                        searchInputRef.current?.blur();
                                    }
                                }}
                            />

                            {/* Search Results Dropdown */}
                            {showSearchResults && searchQuery && (
                                <div className="absolute top-full mt-2 w-full bg-black/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-2xl shadow-blue-500/10 overflow-hidden z-50 max-h-96 overflow-y-auto">
                                    {(() => {
                                        const repos = catalogData?.data?.repositories || [];
                                        const filtered = repos.filter((name: string) =>
                                            name.toLowerCase().includes(searchQuery.toLowerCase())
                                        );

                                        if (filtered.length === 0) {
                                            return (
                                                <div className="p-6 text-center text-gray-500 font-mono text-sm">
                                                    No repositories found
                                                </div>
                                            );
                                        }

                                        return (
                                            <>
                                                <div className="px-4 py-3 border-b border-white/5 bg-white/5">
                                                    <span className="text-xs font-black text-gray-400 uppercase tracking-widest">
                                                        {filtered.length} {filtered.length === 1 ? 'Repository' : 'Repositories'} Found
                                                    </span>
                                                </div>
                                                {filtered.slice(0, 10).map((repo: string) => (
                                                    <Link
                                                        key={repo}
                                                        to={`/repositories/${repo}`}
                                                        onClick={() => {
                                                            setShowSearchResults(false);
                                                            setSearchQuery('');
                                                        }}
                                                        className="flex items-center gap-3 px-4 py-3 hover:bg-white/5 transition-colors border-b border-white/5 last:border-0 group"
                                                    >
                                                        <Library size={16} className="text-blue-400 group-hover:text-blue-300" />
                                                        <span className="font-mono text-sm text-white group-hover:text-blue-300 transition-colors">
                                                            {repo}
                                                        </span>
                                                    </Link>
                                                ))}
                                                {filtered.length > 10 && (
                                                    <div className="px-4 py-3 bg-white/5 border-t border-white/10">
                                                        <button
                                                            onClick={() => {
                                                                navigate(`/repositories?search=${encodeURIComponent(searchQuery)}`);
                                                                setShowSearchResults(false);
                                                                setSearchQuery('');
                                                            }}
                                                            className="text-xs font-bold text-blue-400 hover:text-blue-300 uppercase tracking-widest"
                                                        >
                                                            View all {filtered.length} results →
                                                        </button>
                                                    </div>
                                                )}
                                            </>
                                        );
                                    })()}
                                </div>
                            )}
                        </div>
                    </div>

                    <div className="flex items-center gap-6">
                        <div className="hidden lg:flex flex-col items-end mr-4">
                            <div className="text-[10px] text-gray-500 uppercase font-bold tracking-widest">Status</div>
                            <div className="flex items-center gap-2 text-xs font-mono">
                                <span className="text-green-400">Online</span>
                            </div>
                        </div>

                        <div className="flex items-center gap-2 relative z-50 pointer-events-auto">
                            {/* Notifications */}
                            <div className="relative" ref={notificationsRef}>
                                <HeaderIcon
                                    icon={<Bell />}
                                    count={notifications.length}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowNotifications(prev => !prev);
                                        setShowActivity(false);
                                    }}
                                />
                                {showNotifications && (
                                    <div className="absolute top-full right-0 mt-4 w-80 bg-[#0a0a0a] border border-white/10 rounded-2xl shadow-2xl shadow-blue-500/10 overflow-hidden z-[100]">
                                        <div className="px-5 py-4 border-b border-white/5 flex justify-between items-center bg-white/5">
                                            <span className="text-xs font-black text-white uppercase tracking-widest">Notifications</span>
                                            <span className="text-[10px] font-bold text-blue-400 bg-blue-500/10 px-2 py-1 rounded-full">{notifications.length} New</span>
                                        </div>
                                        <div className="max-h-80 overflow-y-auto">
                                            {notifications.length === 0 ? (
                                                <div className="p-8 text-center text-gray-500 font-mono text-xs">
                                                    No new notifications
                                                </div>
                                            ) : (
                                                notifications.map((notif, i) => (
                                                    <div key={i} className="px-5 py-4 border-b border-white/5 hover:bg-white/5 transition-colors group cursor-pointer">
                                                        <div className="flex justify-between items-start mb-1">
                                                            <span className={`text-xs font-bold uppercase tracking-wider ${notif.type === 'critical' ? 'text-red-400' :
                                                                notif.type === 'success' ? 'text-green-400' : 'text-blue-400'
                                                                }`}>{notif.title}</span>
                                                            <span className="text-[10px] text-gray-600 font-mono">{notif.time}</span>
                                                        </div>
                                                        <p className="text-xs text-gray-400 leading-relaxed group-hover:text-gray-300 transition-colors">
                                                            {notif.desc}
                                                        </p>
                                                    </div>
                                                ))
                                            )}
                                        </div>
                                        <div className="p-3 bg-white/5 border-t border-white/10 text-center">
                                            <button
                                                onClick={() => setNotifications([])}
                                                className="text-[10px] font-bold text-gray-500 hover:text-white uppercase tracking-widest transition-colors"
                                            >
                                                Mark all as read
                                            </button>
                                        </div>
                                    </div>
                                )}
                            </div>

                            {/* Activity */}
                            <div className="relative" ref={activityRef}>
                                <HeaderIcon
                                    icon={<Activity />}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowActivity(prev => !prev);
                                        setShowNotifications(false);
                                    }}
                                />
                                {showActivity && (
                                    <div className="absolute top-full right-0 mt-4 w-72 bg-[#0a0a0a] border border-white/10 rounded-2xl shadow-2xl shadow-blue-500/10 overflow-hidden z-[100]">
                                        <div className="px-5 py-4 border-b border-white/5 bg-white/5">
                                            <span className="text-xs font-black text-white uppercase tracking-widest">System Pulse</span>
                                        </div>
                                        <div className="p-2">
                                            {[
                                                { label: 'CPU Usage', value: '45%', color: 'bg-blue-500' },
                                                { label: 'Memory', value: '2.1GB / 8GB', color: 'bg-purple-500' },
                                                { label: 'Active Scans', value: '2 Running', color: 'bg-green-500' },
                                                { label: 'Repositories', value: String(catalogData?.data?.repositories?.length || 0), color: 'bg-orange-500' }
                                            ].map((stat, i) => (
                                                <div key={i} className="p-3 hover:bg-white/5 rounded-xl transition-colors">
                                                    <div className="flex justify-between items-center mb-2">
                                                        <span className="text-[10px] font-bold text-gray-500 uppercase tracking-widest">{stat.label}</span>
                                                        <span className="text-xs font-mono font-bold text-white">{stat.value}</span>
                                                    </div>
                                                    <div className="h-1 w-full bg-white/10 rounded-full overflow-hidden">
                                                        <div className={`h-full ${stat.color} w-2/3 rounded-full opacity-80`} />
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                        <div className="px-5 py-3 bg-white/5 border-t border-white/10 flex justify-between items-center">
                                            <div className="flex items-center gap-2">
                                                <div className="w-2 h-2 rounded-full bg-green-500 animate-pulse" />
                                                <span className="text-[10px] font-bold text-gray-400 uppercase tracking-widest">All Systems Operational</span>
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        </div>

                        <div className="h-8 w-[1px] bg-white/5 ml-2" />

                        <Link to="/profile" className="flex items-center gap-3 pl-2 group">
                            <div className="w-10 h-10 rounded-xl bg-gradient-to-tr from-gray-800 to-gray-700 p-[1px] group-hover:from-blue-500 group-hover:to-cyan-400 transition-all">
                                <div className="w-full h-full rounded-xl bg-[#020204] flex items-center justify-center overflow-hidden">
                                    <User size={18} className="text-gray-400 group-hover:text-blue-400 transition-colors" />
                                </div>
                            </div>
                        </Link>
                    </div>
                </header>

                <main className="flex-1 p-8 overflow-y-auto">
                    <div className="max-w-[1600px] mx-auto animate-page-fade">
                        <Outlet />
                    </div>
                </main>
            </div>
        </div>
    );
}

function NavLink({ to, icon, label, collapsed }: { to: string, icon: React.ReactElement, label: string, collapsed: boolean }) {
    const location = useLocation();
    const active = location.pathname.startsWith(to);

    return (
        <Link
            to={to}
            className={`flex items-center gap-4 px-4 py-3.5 rounded-2xl text-xs font-bold uppercase tracking-widest transition-all duration-300 relative group overflow-hidden ${active
                ? "text-blue-400 bg-blue-500/10 border border-blue-500/20 shadow-[0_0_30px_rgba(59,130,246,0.1)]"
                : "text-gray-500 hover:text-white hover:bg-white/5 border border-transparent"
                }`}
        >
            {active && (
                <div className="absolute left-0 top-0 w-1 h-full bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]" />
            )}

            <div className={`transition-transform duration-300 group-hover:scale-110 ${active ? "animate-pulse-glow" : ""
                }`}>
                {React.cloneElement(icon, { size: 20 })}
            </div>

            {!collapsed && (
                <span className="flex-1 truncate">{label}</span>
            )}

            {active && !collapsed && (
                <div className="w-1.5 h-1.5 rounded-full bg-blue-400 animate-ping" />
            )}
        </Link>
    );
}

function HeaderIcon({ icon, count, onClick }: { icon: React.ReactNode, count?: number, onClick?: React.MouseEventHandler }) {
    return (
        <button
            type="button"
            onClick={onClick}
            className="relative p-3 rounded-xl bg-white/5 border border-white/5 hover:border-blue-500/30 hover:bg-white/10 transition-all group text-gray-500 hover:text-white pointer-events-auto"
        >
            {icon}
            {count && (
                <span className="absolute top-2 right-2 w-4 h-4 bg-blue-500 rounded-full text-[8px] font-black text-white flex items-center justify-center border-2 border-[#020204]">
                    {count}
                </span>
            )}
        </button>
    );
}


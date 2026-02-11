import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Repositories from './pages/Repositories';
import RepositoryDetails from './pages/RepositoryDetails';
import SecurityPolicies from './pages/SecurityPolicies';
import Settings from './pages/Settings';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider, useAuth } from './lib/auth-context';
import Login from './pages/Login';
import Register from './pages/Register';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import UserProfile from './pages/UserProfile';
import DependencyGraph from './pages/DependencyGraph';
import CostIntelligence from './pages/CostIntelligence';
import Sessions from './pages/Sessions';
import Landing from './pages/Landing';
import QuickStart from './pages/QuickStart';

import { Loader2 } from 'lucide-react';

const queryClient = new QueryClient();

const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
    const { token, isLoading } = useAuth();

    if (isLoading) {
        return (
            <div className="h-screen w-screen flex items-center justify-center bg-black text-white">
                <Loader2 className="animate-spin text-blue-500" size={40} />
            </div>
        );
    }

    if (!token) {
        return <Navigate to="/login" replace />;
    }

    return <>{children}</>;
};

function App() {
    return (
        <QueryClientProvider client={queryClient}>
            <AuthProvider>
                <BrowserRouter>
                    <Routes>
                        {/* Static Redirects for Legacy Paths */}
                        <Route path="/dashboard" element={<Navigate to="/app/dashboard" replace />} />
                        <Route path="/repositories" element={<Navigate to="/app/repositories" replace />} />
                        <Route path="/lineage" element={<Navigate to="/app/lineage" replace />} />
                        <Route path="/costs" element={<Navigate to="/app/costs" replace />} />
                        <Route path="/sessions" element={<Navigate to="/app/sessions" replace />} />
                        <Route path="/policies" element={<Navigate to="/app/policies" replace />} />
                        <Route path="/settings" element={<Navigate to="/app/settings" replace />} />
                        <Route path="/profile" element={<Navigate to="/app/profile" replace />} />

                        {/* Auth Routes */}
                        <Route path="/login" element={<Login />} />
                        <Route path="/register" element={<Register />} />
                        <Route path="/forgot-password" element={<ForgotPassword />} />
                        <Route path="/reset-password" element={<ResetPassword />} />

                        {/* Public Marketing Routes */}
                        <Route path="/" element={<Landing />} />


                        {/* Protected App Workspace */}
                        <Route path="/app" element={
                            <ProtectedRoute>
                                <Layout />
                            </ProtectedRoute>
                        }>
                            <Route index element={<Navigate to="/app/dashboard" replace />} />
                            <Route path="dashboard" element={<Dashboard />} />
                            <Route path="repositories" element={<Repositories />} />
                            <Route path="repositories/:name" element={<RepositoryDetails />} />
                            <Route path="lineage" element={<DependencyGraph />} />
                            <Route path="costs" element={<CostIntelligence />} />
                            <Route path="sessions" element={<Sessions />} />
                            <Route path="policies" element={<SecurityPolicies />} />
                            <Route path="settings" element={<Settings />} />
                            <Route path="profile" element={<UserProfile />} />
                            <Route path="quickstart" element={<QuickStart />} />

                            {/* Inner Catch-all */}
                            <Route path="*" element={<Navigate to="/app/dashboard" replace />} />
                        </Route>

                        {/* Global Catch-all */}
                        <Route path="*" element={<Navigate to="/" replace />} />
                    </Routes>
                </BrowserRouter>
            </AuthProvider>
        </QueryClientProvider>
    );
}

export default App;

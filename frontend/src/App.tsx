import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Repositories from './pages/Repositories';
import RepositoryDetails from './pages/RepositoryDetails';
import Policies from './pages/Policies';
import Settings from './pages/Settings';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

import { AuthProvider, useAuth } from './lib/auth-context';
import Login from './pages/Login';
import Register from './pages/Register';
import ForgotPassword from './pages/ForgotPassword';
import ResetPassword from './pages/ResetPassword';
import UserProfile from './pages/UserProfile';
import DependencyGraph from './pages/DependencyGraph';
import CostIntelligence from './pages/CostIntelligence';
import Sessions from './pages/Sessions';
import { Loader2 } from 'lucide-react';

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
                        <Route path="/login" element={<Login />} />
                        <Route path="/login" element={<Login />} />
                        <Route path="/register" element={<Register />} />
                        <Route path="/forgot-password" element={<ForgotPassword />} />
                        <Route path="/reset-password" element={<ResetPassword />} />

                        <Route path="/" element={
                            <ProtectedRoute>
                                <Layout />
                            </ProtectedRoute>
                        }>
                            <Route index element={<Navigate to="/dashboard" replace />} />
                            <Route path="dashboard" element={<Dashboard />} />
                            <Route path="repositories" element={<Repositories />} />
                            <Route path="repositories/:name" element={<RepositoryDetails />} />
                            <Route path="lineage" element={<DependencyGraph />} />
                            <Route path="costs" element={<CostIntelligence />} />
                            <Route path="sessions" element={<Sessions />} />
                            <Route path="policies" element={<Policies />} />
                            <Route path="settings" element={<Settings />} />
                            <Route path="profile" element={<UserProfile />} />

                            <Route path="*" element={<Navigate to="/dashboard" replace />} />
                        </Route>
                    </Routes>
                </BrowserRouter>
            </AuthProvider>
        </QueryClientProvider>
    );
}

export default App;

import axios from 'axios';

// Use relative URL to leverage Nginx proxy
const API_URL = ''; // Relative to current origin

const axiosInstance = axios.create({
    baseURL: API_URL,
});

// Interceptor to add Token
axiosInstance.interceptors.request.use((config) => {
    const token = sessionStorage.getItem('registryx_token') || sessionStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// ... existing interfaces ...

export interface VulnerabilitySummary {
    critical: number;
    high: number;
    medium: number;
    low: number;
    high_priority: number;
}

export interface HealthScore {
    overall: number;        // 0-100
    security: number;       // 0-100
    freshness: number;      // 0-100
    efficiency: number;     // 0-100
    maintenance: number;    // 0-100
    grade: string;          // A+, A, B, C, D, F
    trend: string;          // improving, stable, declining
    lastUpdated: string;
}

export interface ManifestDetails {
    digest: string;
    size: number;
    mediaType: string;
    vulnerabilities?: VulnerabilitySummary;
    isSigned?: boolean;
    healthScore?: HealthScore;
}

export interface ScanStatus {
    status: 'pending' | 'scanning' | 'completed' | 'failed';
    scanned_at?: string;
    summary?: VulnerabilitySummary;
    error?: string;
    progress_message?: string;
}

export interface ScanHistoryEntry {
    id: string;
    status: string;
    scanned_at?: string;
    summary?: VulnerabilitySummary;
}

export interface ServiceAccount {
    id: string;
    name: string;
    created: string;
    lastUsed: string;
    status: 'active' | 'revoked';
}

// OCI Registry Auth Helpers (For internal use if needed)
// ... existing code ...


export const registry = {
    getCatalog: async () => {
        return axiosInstance.get('/v2/_catalog');
    },

    getTags: async (repoName: string) => {
        return axiosInstance.get(`/v2/${repoName}/tags/list`);
    },

    getManifest: async (repoName: string, reference: string) => {
        return axiosInstance.get(`/v2/${repoName}/manifests/${reference}`);
    },

    createRepository: async (repoName: string) => {
        return axiosInstance.post('/api/v1/repositories', { name: repoName });
    },

    deleteRepository: async (repoName: string) => {
        const url = `/api/v1/repositories/${encodeURIComponent(repoName)}`;
        console.log(`[API] calling DELETE ${url}`);
        return axiosInstance.delete(url);
    },

    deleteTag: async (repoName: string, tagName: string) => {
        return axiosInstance.delete(`/api/v1/repositories/${encodeURIComponent(repoName)}/tags/${encodeURIComponent(tagName)}`);
    },
};

const customApi = {
    // Expose Axios methods
    get: axiosInstance.get.bind(axiosInstance),
    post: axiosInstance.post.bind(axiosInstance),
    put: axiosInstance.put.bind(axiosInstance),
    delete: axiosInstance.delete.bind(axiosInstance),

    // Custom Domain Methods

    // Auth
    login: async (username: string, password: string) => {
        return axiosInstance.post('/api/v1/auth/login', { username, password });
    },
    register: async (username: string, email: string, password: string) => {
        return axiosInstance.post('/api/v1/auth/register', { username, email, password });
    },
    forgotPassword: async (email: string) => {
        return axiosInstance.post('/api/v1/auth/forgot-password', { email });
    },
    resetPassword: async (token: string, newPassword: string) => {
        return axiosInstance.post('/api/v1/auth/reset-password', { token, newPassword });
    },
    resetPasswordWithKey: async (email: string, recoveryKey: string, newPassword: string) => {
        return axiosInstance.post('/api/v1/auth/reset-with-key', { email, recoveryKey, newPassword });
    },
    changePassword: async (newPassword: string) => {
        return axiosInstance.post('/api/v1/auth/change-password', { newPassword });
    },
    logout: async () => {
        return axiosInstance.post('/api/v1/auth/logout');
    },

    getRepositories: async () => {
        // @ts-ignore
        return axiosInstance.get('/v2/_catalog');
    },
    getTags: async (repo: string) => {
        // @ts-ignore
        return axiosInstance.get(`/v2/${repo}/tags/list`);
    },
    getManifestDetails: async (repo: string, reference: string) => {
        // Real Endpoint: /api/v1/repositories/{name}/manifests/{reference}
        return axiosInstance.get<ManifestDetails>(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}`);
    },

    getPolicy: async () => {
        // Real Endpoint: /api/v1/policy
        return axiosInstance.get<{ rego: string }>('/api/v1/policy');
    },

    updatePolicy: async (rego: string) => {
        // Real Endpoint: /api/v1/policy (PUT)
        // Send raw string or JSON? Handler reads Body as string.
        return axiosInstance.put('/api/v1/policy', rego, { headers: { 'Content-Type': 'text/plain' } });
    },

    // Service Accounts
    getServiceAccounts: async () => {
        return axiosInstance.get<{ data: ServiceAccount[] }>('/api/v1/service-accounts');
    },

    createServiceAccount: async (name: string, description: string) => {
        return axiosInstance.post<{ account: ServiceAccount, apiKey: string }>('/api/v1/service-accounts', { name, description });
    },

    revokeServiceAccount: async (id: string) => {
        return axiosInstance.delete(`/api/v1/service-accounts/${id}`);
    },

    // Dashboard Stats
    getDashboardStats: async () => {
        return axiosInstance.get<{
            repositories: number,
            images: number,
            vulnerabilities: number,
            storageUsed: string,
            recentPushes: { repository: string, tag: string, digest: string, pushedAt: string }[],
            severity: { critical: number, level: number, high: number, medium: number, low: number }
        }>('/api/v1/stats');
    },

    // System
    runGarbageCollection: async (dryRun: boolean = false) => {
        const url = dryRun ? '/api/v1/system/gc?dryRun=true' : '/api/v1/system/gc';
        return axiosInstance.post<{
            blobsDeleted: number,
            spaceFreedBytes: number,
            spaceFreedMB: string,
            duration: string,
            errors?: string[]
        }>(url);
    },

    // Dependencies
    getDependencyGraph: async (repository?: string) => {
        const url = repository ? `/api/v1/dependencies?repository=${encodeURIComponent(repository)}` : '/api/v1/dependencies';
        return axiosInstance.get<{ nodes: any[], edges: any[] }>(url);
    },

    // Scan Features
    getScanStatus: async (repo: string, reference: string) => {
        return axiosInstance.get<ScanStatus>(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}/scan/status`);
    },

    downloadScanReport: async (repo: string, reference: string) => {
        const response = await axiosInstance.get(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}/scan/report`, {
            responseType: 'blob'
        });
        // Create download link
        const url = window.URL.createObjectURL(new Blob([response.data]));
        const link = document.createElement('a');
        link.href = url;
        link.setAttribute('download', `trivy-report-${repo}-${reference}.json`);
        document.body.appendChild(link);
        link.click();
        link.remove();
        window.URL.revokeObjectURL(url);
        return response;
    },

    getScanReportJSON: async (repo: string, reference: string) => {
        return axiosInstance.get<any>(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}/scan/report`);
    },

    getScanHistory: async (repo: string, reference: string) => {
        return axiosInstance.get<{ scans: ScanHistoryEntry[] }>(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}/scan/history`);
    },

    triggerManualScan: async (repo: string, reference: string) => {
        return axiosInstance.post<{ message: string, status: string }>(`/api/v1/repositories/${encodeURIComponent(repo)}/manifests/${reference}/scan/trigger`);
    },

    // Config
    getSystemConfig: async () => {
        return axiosInstance.get<{ enableCostIntelligence: boolean }>('/api/v1/system/config');
    },

    // Sessions (Admin)
    getActiveSessions: async () => {
        return axiosInstance.get<any[]>('/api/v1/system/sessions');
    },
    revokeSession: async (id: string) => {
        return axiosInstance.delete(`/api/v1/system/sessions/${id}`);
    }
};

export const api = customApi;
export default customApi;

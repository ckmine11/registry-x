import React from 'react';

interface PriorityBadgeProps {
    score: number;
    action: string;
    epss?: number;
    className?: string;
}

export default function PriorityBadge({ score, action, epss, className = '' }: PriorityBadgeProps) {
    const getColor = () => {
        if (score >= 80) return 'bg-red-900 text-red-200 border-red-700';
        if (score >= 60) return 'bg-orange-900 text-orange-200 border-orange-700';
        if (score >= 40) return 'bg-yellow-900 text-yellow-200 border-yellow-700';
        if (score >= 20) return 'bg-blue-900 text-blue-200 border-blue-700';
        return 'bg-gray-900 text-gray-200 border-gray-700';
    };

    const getIcon = () => {
        if (score >= 80) return '🔴';
        if (score >= 60) return '🟠';
        if (score >= 40) return '🟡';
        if (score >= 20) return '🔵';
        return '⚪';
    };

    const getLabel = () => {
        switch (action) {
            case 'urgent': return 'URGENT';
            case 'high': return 'HIGH';
            case 'medium': return 'MEDIUM';
            case 'low': return 'LOW';
            case 'monitor': return 'MONITOR';
            default: return action.toUpperCase();
        }
    };

    return (
        <div className={`inline-flex items-center gap-2 px-3 py-1.5 rounded-lg border ${getColor()} ${className}`}>
            <span className="text-lg">{getIcon()}</span>
            <div className="flex flex-col">
                <span className="text-xs font-bold">{getLabel()}</span>
                <div className="flex items-center gap-2 text-xs opacity-80">
                    <span>Score: {score}</span>
                    {epss !== undefined && (
                        <>
                            <span>•</span>
                            <span>EPSS: {(epss * 100).toFixed(1)}%</span>
                        </>
                    )}
                </div>
            </div>
        </div>
    );
}

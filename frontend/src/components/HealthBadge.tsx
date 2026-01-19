import React from 'react';
import { HealthScore } from '../lib/api';

interface HealthBadgeProps {
    score: HealthScore;
    size?: 'small' | 'medium' | 'large';
    showDetails?: boolean;
}

export const HealthBadge: React.FC<HealthBadgeProps> = ({ score, size = 'medium', showDetails = false }) => {
    const getColorClass = (overall: number): string => {
        if (overall >= 80) return 'bg-green-500';
        if (overall >= 60) return 'bg-yellow-500';
        if (overall >= 40) return 'bg-orange-500';
        return 'bg-red-500';
    };

    const getTextColorClass = (overall: number): string => {
        if (overall >= 80) return 'text-green-600';
        if (overall >= 60) return 'text-yellow-600';
        if (overall >= 40) return 'text-orange-600';
        return 'text-red-600';
    };

    const getBorderColorClass = (overall: number): string => {
        if (overall >= 80) return 'border-green-500';
        if (overall >= 60) return 'border-yellow-500';
        if (overall >= 40) return 'border-orange-500';
        return 'border-red-500';
    };

    const getDescription = (overall: number): string => {
        if (overall >= 90) return 'Excellent';
        if (overall >= 75) return 'Good';
        if (overall >= 60) return 'Fair';
        if (overall >= 40) return 'Poor';
        return 'Critical';
    };

    const sizeClasses = {
        small: 'text-xs px-2 py-1',
        medium: 'text-sm px-3 py-1.5',
        large: 'text-base px-4 py-2'
    };

    if (!showDetails) {
        // Compact badge view
        return (
            <div className={`inline-flex items-center gap-2 rounded-full ${sizeClasses[size]} ${getColorClass(score.overall)} text-white font-semibold`}>
                <span>{score.overall}</span>
                <span className="text-xs opacity-90">{score.grade}</span>
            </div>
        );
    }

    // Detailed view with breakdown
    return (
        <div className={`border-2 ${getBorderColorClass(score.overall)} rounded-lg p-4 bg-white`}>
            <div className="flex items-center justify-between mb-3">
                <div>
                    <div className="flex items-center gap-2">
                        <span className={`text-3xl font-bold ${getTextColorClass(score.overall)}`}>
                            {score.overall}
                        </span>
                        <div>
                            <div className={`text-lg font-semibold ${getTextColorClass(score.overall)}`}>
                                Grade {score.grade}
                            </div>
                            <div className="text-sm text-gray-600">
                                {getDescription(score.overall)}
                            </div>
                        </div>
                    </div>
                </div>
                <div className={`${getColorClass(score.overall)} text-white px-3 py-1 rounded-full text-xs font-medium`}>
                    {score.trend === 'improving' && '📈 Improving'}
                    {score.trend === 'stable' && '➡️ Stable'}
                    {score.trend === 'declining' && '📉 Declining'}
                </div>
            </div>

            {/* Score Breakdown */}
            <div className="space-y-2">
                <ScoreBar label="Security" score={score.security} icon="🛡️" />
                <ScoreBar label="Freshness" score={score.freshness} icon="🕐" />
                <ScoreBar label="Efficiency" score={score.efficiency} icon="⚡" />
                <ScoreBar label="Maintenance" score={score.maintenance} icon="🔧" />
            </div>

            <div className="mt-3 text-xs text-gray-500">
                Last updated: {new Date(score.lastUpdated).toLocaleString()}
            </div>
        </div>
    );
};

interface ScoreBarProps {
    label: string;
    score: number;
    icon: string;
}

const ScoreBar: React.FC<ScoreBarProps> = ({ label, score, icon }) => {
    const getBarColor = (score: number): string => {
        if (score >= 80) return 'bg-green-500';
        if (score >= 60) return 'bg-yellow-500';
        if (score >= 40) return 'bg-orange-500';
        return 'bg-red-500';
    };

    return (
        <div>
            <div className="flex items-center justify-between text-sm mb-1">
                <span className="text-gray-700 flex items-center gap-1">
                    <span>{icon}</span>
                    <span>{label}</span>
                </span>
                <span className="font-semibold text-gray-900">{score}/100</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
                <div
                    className={`h-2 rounded-full ${getBarColor(score)} transition-all duration-300`}
                    style={{ width: `${score}%` }}
                />
            </div>
        </div>
    );
};

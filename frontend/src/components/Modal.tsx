import React, { useEffect, useRef } from 'react';
import clsx from 'clsx';
import { X } from 'lucide-react';

export interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: React.ReactNode;
    children: React.ReactNode;
    footer?: React.ReactNode;
    variant?: 'default' | 'danger' | 'warning' | 'success';
}

export function Modal({ isOpen, onClose, title, children, footer, variant = 'default' }: ModalProps) {
    const ref = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (e.key === 'Escape') onClose();
        };
        if (isOpen) {
            document.addEventListener('keydown', handleEscape);
            document.body.style.overflow = 'hidden';
        }
        return () => {
            document.removeEventListener('keydown', handleEscape);
            document.body.style.overflow = 'unset';
        };
    }, [isOpen, onClose]);

    if (!isOpen) return null;

    const variantStyles = {
        default: "border-blue-500/20 shadow-[0_0_50px_-12px_rgba(59,130,246,0.5)]",
        danger: "border-red-500/20 shadow-[0_0_50px_-12px_rgba(239,68,68,0.5)]",
        warning: "border-yellow-500/20 shadow-[0_0_50px_-12px_rgba(234,179,8,0.5)]",
        success: "border-green-500/20 shadow-[0_0_50px_-12px_rgba(34,197,94,0.5)]",
    };

    const headerStyles = {
        default: "text-blue-400",
        danger: "text-red-500",
        warning: "text-yellow-500",
        success: "text-green-500",
    };

    return (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-200">
            <div
                ref={ref}
                className={clsx(
                    "relative w-full max-w-lg bg-[#0a0a0a] rounded-2xl border shadow-2xl overflow-hidden animate-in zoom-in-95 duration-200",
                    variantStyles[variant]
                )}
            >
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b border-white/5 bg-white/[0.02]">
                    <h2 className={clsx("text-xl font-black uppercase tracking-tight flex items-center gap-3", headerStyles[variant])}>
                        {title}
                    </h2>
                    <button
                        onClick={onClose}
                        className="p-2 -mr-2 text-gray-500 hover:text-white transition-colors rounded-lg hover:bg-white/5"
                    >
                        <X size={20} />
                    </button>
                </div>

                {/* Body */}
                <div className="p-8 text-gray-300 font-mono text-sm leading-relaxed tracking-wide">
                    {children}
                </div>

                {/* Footer */}
                {footer && (
                    <div className="flex items-center justify-end gap-3 p-6 pt-0">
                        {footer}
                    </div>
                )}
            </div>
        </div>
    );
}

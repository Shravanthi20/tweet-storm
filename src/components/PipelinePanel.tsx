import React from 'react';

interface PipelinePanelProps {
    logs: string[];
    assignedWorkerId?: string;
}

export const PipelinePanel: React.FC<PipelinePanelProps> = ({ logs, assignedWorkerId }) => {
    if (logs.length === 0) return null;

    // Limit logs to last 5 for compactness
    const displayLogs = logs.slice(-5);

    return (
        <div className="mt-4 border-t border-gray-800 p-4 bg-black/50 backdrop-blur-sm sticky bottom-0 z-10">
            <div className="max-w-3xl mx-auto">
                {/* COMPACT HORIZONTAL DIAGRAM */}
                <div className="flex items-center justify-between mb-6 px-4 py-3 bg-gray-900/50 rounded-xl border border-gray-800">
                    <div className="flex flex-col items-center gap-1">
                        <div className="w-8 h-8 rounded-lg bg-blue-500/20 border border-blue-500/50 flex items-center justify-center text-blue-400">
                            <span className="text-[10px] font-bold">CLI</span>
                        </div>
                        <span className="text-[9px] text-gray-500 uppercase font-bold">Client</span>
                    </div>

                    <div className="h-px flex-1 bg-gradient-to-r from-blue-500/50 to-purple-500/50 mx-2 opacity-30" />

                    <div className="flex flex-col items-center gap-1">
                        <div className="w-8 h-8 rounded-lg bg-purple-500/20 border border-purple-500/50 flex items-center justify-center text-purple-400">
                            <span className="text-[10px] font-bold">LDR</span>
                        </div>
                        <span className="text-[9px] text-gray-500 uppercase font-bold">Leader</span>
                    </div>

                    <div className="h-px flex-1 bg-gradient-to-r from-purple-500/50 to-green-500/50 mx-2 opacity-30" />

                    <div className="flex flex-col items-center gap-1">
                        <div className={`w-10 h-10 rounded-lg flex items-center justify-center transition-all duration-500 ${assignedWorkerId && assignedWorkerId !== 'none' ? 'bg-green-500/30 border-green-500 ring-4 ring-green-500/10' : 'bg-gray-800 border-gray-700'} border`}>
                            <span className={`text-[10px] font-bold ${assignedWorkerId && assignedWorkerId !== 'none' ? 'text-green-400' : 'text-gray-500'}`}>
                                {assignedWorkerId?.split('-')[1] || 'W?'}
                            </span>
                        </div>
                        <span className={`text-[9px] uppercase font-bold ${assignedWorkerId && assignedWorkerId !== 'none' ? 'text-green-500' : 'text-gray-500'}`}>
                            Worker
                        </span>
                    </div>

                    <div className="h-px flex-1 bg-gradient-to-r from-green-500/50 to-amber-500/50 mx-2 opacity-30" />

                    <div className="flex flex-col items-center gap-1">
                        <div className="w-8 h-8 rounded-lg bg-amber-500/20 border border-amber-500/50 flex items-center justify-center text-amber-400">
                            <span className="text-[10px] font-bold">DB</span>
                        </div>
                        <span className="text-[9px] text-gray-500 uppercase font-bold">Storage</span>
                    </div>
                </div>

                {/* RECENT STEPS (LAST 5) */}
                <div className="space-y-1.5 max-h-32 overflow-y-auto pr-2 custom-scrollbar">
                    {displayLogs.map((log, index) => {
                        const isLast = index === displayLogs.length - 1;
                        return (
                            <div key={index} className={`text-[11px] py-1.5 px-3 rounded border flex items-center gap-3 transition-opacity duration-300 ${isLast ? 'bg-blue-900/10 border-blue-900/40 text-blue-200' : 'bg-gray-900/30 border-gray-800/50 text-gray-500 opacity-60'}`}>
                                <span className={`w-1.5 h-1.5 rounded-full ${isLast ? 'bg-blue-400 animate-pulse' : 'bg-gray-700'}`} />
                                {log}
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

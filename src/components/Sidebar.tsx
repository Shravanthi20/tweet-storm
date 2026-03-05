import React from 'react';
import { SimulationState } from '../lib/types';

interface SidebarProps {
    state: SimulationState;
    onSimulate: (text: string) => void;
    onReset: () => void;
}

export const LeftSidebar: React.FC<SidebarProps> = ({ onSimulate, onReset }) => {
    const [tweet, setTweet] = React.useState('');

    const handleSimulate = () => {
        if (!tweet.trim()) return;
        onSimulate(tweet);
        setTweet('');
    };

    return (
        <div className="w-64 p-6 border-r border-gray-800 h-screen sticky top-0 bg-black text-white flex flex-col gap-6">
            <div className="text-2xl font-bold text-blue-400">StormSim</div>

            <div className="flex flex-col gap-4">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Controls</h3>
                <textarea
                    value={tweet}
                    onChange={(e) => setTweet(e.target.value)}
                    placeholder="What's happening in the cluster?"
                    className="bg-gray-900 border border-gray-700 rounded-xl p-3 text-sm focus:ring-2 focus:ring-blue-500 outline-none resize-none h-24"
                />
                <button
                    onClick={handleSimulate}
                    className="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded-full transition-colors"
                >
                    Simulate Event
                </button>
                <button
                    onClick={onReset}
                    className="border border-gray-700 hover:bg-gray-900 text-white font-bold py-2 px-4 rounded-full transition-colors"
                >
                    Reset Simulation
                </button>
            </div>

            <div className="mt-auto text-xs text-gray-500">
                <p>Person-1: Lamport Clocks</p>
                <p>Distributed Systems Project</p>
            </div>
        </div>
    );
};

import { sharedStorage } from '../lib/shared/storage';

export const RightSidebar: React.FC<{ state: SimulationState }> = ({ state }) => {
    return (
        <div className="w-80 p-6 border-l border-gray-800 h-screen sticky top-0 bg-black text-white flex flex-col gap-6 overflow-y-auto">
            <div>
                <h3 className="text-lg font-bold mb-4">Cluster Status</h3>
                <div className={`p-3 rounded-lg flex items-center gap-3 ${state.leader.status === 'ACTIVE' ? 'bg-green-900/20 text-green-400' : 'bg-red-900/20 text-red-400'}`}>
                    <div className={`w-2 h-2 rounded-full ${state.leader.status === 'ACTIVE' ? 'bg-green-400' : 'bg-red-400'}`} />
                    <span>Leader: {state.leader.id}</span>
                    <span className="ml-auto text-xs font-mono">Clock: {state.leader.clock}</span>
                </div>
            </div>

            <div className="space-y-4">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider">Worker Nodes</h3>
                {state.workers.map((worker) => (
                    <div key={worker.id} className="p-4 bg-gray-900 rounded-xl border border-gray-800">
                        <div className="flex justify-between items-center mb-2">
                            <span className="font-bold text-sm">{worker.id}</span>
                            <span className="px-2 py-0.5 rounded text-[10px] bg-blue-900/30 text-blue-400 border border-blue-800">ACTIVE</span>
                        </div>
                        <div className="flex justify-between text-xs text-gray-400">
                            <span>Logical Clock:</span>
                            <span className="font-mono text-white">{worker.clock}</span>
                        </div>
                        <div className="flex justify-between text-xs text-gray-400 mt-1">
                            <span>Enqueued:</span>
                            <span className="font-mono text-white">{worker.queueSize}</span>
                        </div>
                    </div>
                ))}
            </div>

            <div className="p-4 bg-gray-900 rounded-xl border border-gray-800">
                <h3 className="text-sm font-semibold text-gray-400 uppercase tracking-wider mb-3">Global Word Count</h3>
                <div className="flex flex-wrap gap-2">
                    {Object.entries(sharedStorage.getWordCounts())
                        .sort(([, a], [, b]) => b - a)
                        .slice(0, 10)
                        .map(([word, count]) => (
                            <div key={word} className="flex items-center gap-2 bg-black/40 border border-gray-800 px-3 py-1 rounded-full text-[11px]">
                                <span className="text-blue-400 font-medium">#{word}</span>
                                <span className="text-gray-500 font-mono">{count}</span>
                            </div>
                        ))}
                    {Object.keys(sharedStorage.getWordCounts()).length === 0 && (
                        <span className="text-xs text-gray-600 italic">No words processed...</span>
                    )}
                </div>
            </div>

            <div className="p-4 bg-gray-900/50 rounded-xl border border-dashed border-gray-700">
                <h4 className="text-xs font-bold text-gray-500 mb-2 uppercase">Placeholders</h4>
                <ul className="text-[11px] text-gray-500 space-y-1">
                    <li>• Bully Algorithm (Election)</li>
                    <li>• Ricart-Agrawala (Mutex)</li>
                    <li>• Banker's (Deadlock)</li>
                </ul>
            </div>
        </div>
    );
};

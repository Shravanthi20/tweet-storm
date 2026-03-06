import React, { useEffect, useState } from 'react';
import './Dashboard.css';

type ClusterState = {
    activeWorkers: string[] | null;
    taskCounts: Record<string, number>;
    recentTasks: Record<string, string>;
};

export const Dashboard = () => {
    const [state, setState] = useState<ClusterState>({
        activeWorkers: [],
        taskCounts: {},
        recentTasks: {},
    });

    const [connected, setConnected] = useState(false);

    useEffect(() => {
        const fetchState = async () => {
            try {
                const leaderIP = import.meta.env.VITE_LEADER_IP || 'localhost';
                const res = await fetch(`http://${leaderIP}:8000/api/state`);
                if (!res.ok) throw new Error('Backend not ok');
                const data = (await res.json()) as ClusterState;
                setState(data);
                setConnected(true);
            } catch (e) {
                setConnected(false);
            }
        };

        fetchState();
        const timer = setInterval(fetchState, 1000);
        return () => clearInterval(timer);
    }, []);

    // Derive workers dynamically from active list and historical counts
    const knownWorkers = Array.from(new Set([
        ...(state.activeWorkers || []),
        ...Object.keys(state.taskCounts || {})
    ])).sort();

    return (
        <div className="dashboard">
            <h2>Task Distribution Tracker</h2>

            {!connected && (
                <div className="alert-disconnected">
                    ⚠️ Disconnected from Leader. Please start the Leader node on port 8000.
                </div>
            )}

            {connected && knownWorkers.length === 0 && (
                <div className="alert-disconnected" style={{ backgroundColor: '#2a2a2a', color: '#aaa', borderColor: '#444' }}>
                    Leader is online, but no workers have been detected yet. Keep this window open and start some workers on separate terminals.
                </div>
            )}

            <div className="worker-grid">
                {knownWorkers.map((worker) => {
                    const isAlive = state.activeWorkers?.includes(worker);
                    const tasksProcessed = state.taskCounts[worker] || 0;
                    const lastTask = state.recentTasks[worker] || 'None';

                    return (
                        <div key={worker} className={`worker-card ${isAlive ? 'alive' : 'dead'}`}>
                            <div className="worker-header">
                                <h3>{worker.replace(/http:\/\/[^:]+:/, 'Node ')}</h3>
                                <span className={`status-badge ${isAlive ? 'online' : 'offline'}`}>
                                    {isAlive ? 'ONLINE' : 'OFFLINE'}
                                </span>
                            </div>

                            <div className="worker-stats">
                                <div className="stat-box">
                                    <span className="stat-label">Tasks Processed</span>
                                    <span className="stat-value">{tasksProcessed}</span>
                                </div>

                                <div className="recent-task">
                                    <span className="stat-label">Latest Tweet</span>
                                    <p className="tweet-text">"{lastTask}"</p>
                                </div>
                            </div>
                        </div>
                    );
                })}
            </div>
        </div>
    );
};

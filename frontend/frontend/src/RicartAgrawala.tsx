import React, { useEffect, useState } from 'react';
import './RicartAgrawala.css';

type RAStatus = {
    nodeId: number;
    state: string;
    clock: number;
    deferredReplies: number[] | null;
    replyCount: number;
    expectedReplies: number;
};

export const RicartAgrawala = () => {
    const [statuses, setStatuses] = useState<Record<number, RAStatus>>({});
    const [connected, setConnected] = useState(false);

    const workerPorts = [8001, 8002, 8003, 8004];

    useEffect(() => {
        const fetchStatuses = async () => {
            let anyConnected = false;
            const newStatuses: Record<number, RAStatus> = {};

            for (const port of workerPorts) {
                try {
                    const res = await fetch(`http://localhost:${port}/ra/status`);
                    if (res.ok) {
                        const data = (await res.json()) as RAStatus;
                        newStatuses[data.nodeId] = data;
                        anyConnected = true;
                    }
                } catch (e) {
                    // Worker might be offline, ignore
                }
            }

            setStatuses(newStatuses);
            setConnected(anyConnected);
            console.log("Polled Statuses:", newStatuses);
        };

        fetchStatuses();
        const timer = setInterval(fetchStatuses, 500); // Poll frequently for smooth UI
        return () => clearInterval(timer);
    }, []);

    const activeNodes = Object.values(statuses);

    // Determine the global state of the shared resource
    const nodeInCritical = activeNodes.find(n => n.state === 'CRITICAL_SECTION');

    return (
        <div className="ra-dashboard">
            <h2>Ricart-Agrawala Mutual Exclusion</h2>
            <p className="subtitle">Visualizing Distributed Locking for <code>globalWordCount</code></p>

            {!connected && (
                <div className="alert-disconnected">
                    ⚠️ Waiting for workers to come online...
                </div>
            )}

            <div className="ra-layout">
                {/* Visual Representation of the Shared Resource */}
                <div className={`shared-resource ${nodeInCritical ? 'locked' : 'unlocked'}`}>
                    <div className="resource-icon">
                        {nodeInCritical ? '🔒' : '🔓'}
                    </div>
                    <h3>Global Word Count</h3>
                    {nodeInCritical ? (
                        <p className="owner-text">Locked by Node {nodeInCritical.nodeId}</p>
                    ) : (
                        <p className="owner-text">Available</p>
                    )}
                </div>

                {/* Worker Nodes Visualizer */}
                <div className="nodes-container">
                    {workerPorts.map(port => {
                        // Assuming nodes are 1 to 4 mapping to 8001 to 8004
                        const nodeId = port - 8000;
                        const status = statuses[nodeId];
                        const isOffline = !status;

                        return (
                            <div key={port} className={`ra-node ${isOffline ? 'offline' : status.state.toLowerCase()}`}>
                                <div className="node-header">
                                    <h4>Node {nodeId}</h4>
                                    <span className="port-label">:{port}</span>
                                </div>

                                {isOffline ? (
                                    <div className="node-body offline-body">
                                        Offline
                                    </div>
                                ) : (
                                    <div className="node-body">
                                        <div className="state-badge">{status.state.replace('_', ' ')}</div>
                                        <div className="clock-metric">
                                            <span className="metric-label">Lamport Time:</span>
                                            <span className="metric-val">{status.clock}</span>
                                        </div>

                                        {status.state === 'WAITING' && (
                                            <div className="waiting-metrics">
                                                <div className="metric">
                                                    Replies: {status.replyCount} / {status.expectedReplies}
                                                </div>
                                            </div>
                                        )}

                                        <div className="deferred-list">
                                            <span className="metric-label">Deferred Replies:</span>
                                            <div className="deferred-pills">
                                                {status.deferredReplies && status.deferredReplies.length > 0 ? (
                                                    status.deferredReplies.map(rid => (
                                                        <span key={rid} className="pill">Node {rid}</span>
                                                    ))
                                                ) : (
                                                    <span className="none">None</span>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};

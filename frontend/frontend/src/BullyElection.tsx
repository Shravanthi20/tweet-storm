import React, { useEffect, useState } from 'react';
import './BullyElection.css';

interface NodeStatus {
    nodeId: number;
    leaderId: number;
    inElection: boolean;
}

export const BullyElection = () => {
    // We expect nodes 1-4 (workers) and 5 (initial leader)
    const allExpectedNodes = [1, 2, 3, 4, 5];

    const [nodes, setNodes] = useState<Record<number, NodeStatus | null>>({});
    const leaderIP = import.meta.env.VITE_LEADER_IP || 'localhost';
    const workerIP = import.meta.env.VITE_WORKER_IP || 'localhost';

    const getIPForNode = (id: number) => {
        if (id === 5) return leaderIP;
        return workerIP;
    };

    const getPortForNode = (id: number) => {
        if (id === 5) return '8000';
        return `800${id}`;
    };

    useEffect(() => {
        const fetchStatuses = async () => {
            const newNodes: Record<number, NodeStatus | null> = {};

            for (const id of allExpectedNodes) {
                const ip = getIPForNode(id);
                const port = getPortForNode(id);
                try {
                    const res = await fetch(`http://${ip}:${port}/bully/status`);
                    if (res.ok) {
                        const data = await res.json() as NodeStatus;
                        newNodes[id] = data;
                    } else {
                        newNodes[id] = null; // Node is likely dead
                    }
                } catch (err) {
                    newNodes[id] = null; // Connection failed
                }
            }
            setNodes(newNodes);
        };

        fetchStatuses();
        const interval = setInterval(fetchStatuses, 1000);
        return () => clearInterval(interval);
    }, []);

    // Determine network consensus leader (the ID most alive nodes think is the leader)
    const aliveNodes = Object.values(nodes).filter(n => n !== null) as NodeStatus[];
    const leaderCounts: Record<number, number> = {};
    aliveNodes.forEach(n => {
        leaderCounts[n.leaderId] = (leaderCounts[n.leaderId] || 0) + 1;
    });

    let consensusLeader = -1;
    let maxVotes = 0;
    Object.entries(leaderCounts).forEach(([id, count]) => {
        if (count > maxVotes) {
            consensusLeader = parseInt(id);
            maxVotes = count;
        }
    });

    return (
        <div className="bully-container">
            <div className="bully-header">
                <h2>Bully Election Algorithm</h2>
                <p>Nodes ping the leader. If the leader fails, the node with the highest ID becomes the new leader.</p>
                <div className="consensus-badge">
                    Consensus Leader: {consensusLeader === -1 ? 'None/Voting' : `Node ${consensusLeader}`}
                </div>
            </div>

            <div className="nodes-grid">
                {allExpectedNodes.map(id => {
                    const status = nodes[id];
                    const isAlive = status !== null && status !== undefined;
                    const isLeader = isAlive && status.leaderId === id;
                    const isElection = isAlive && status.inElection;

                    let cardClass = 'node-card';
                    if (!isAlive) cardClass += ' dead';
                    else if (isElection) cardClass += ' election';
                    else if (isLeader) cardClass += ' leader';
                    else cardClass += ' follower';

                    return (
                        <div key={id} className={cardClass}>
                            <div className="node-id-circle">
                                {id}
                            </div>
                            <h3>Node {id}</h3>
                            <div className="node-details">
                                <p><strong>Status:</strong> {isAlive ? 'Online' : 'Offline'}</p>
                                {isAlive && (
                                    <>
                                        <p><strong>Thinks Leader is:</strong> Node {status.leaderId}</p>
                                        <p className="state-text">
                                            {isElection ? '🚨 In Election!' : (isLeader ? '👑 Leader' : '👀 Follower')}
                                        </p>
                                    </>
                                )}
                            </div>
                            {isAlive && !isLeader && !isElection && status.leaderId !== 0 && (
                                <div className="ping-indicator">
                                    Pinging Leader {status.leaderId}
                                </div>
                            )}
                        </div>
                    );
                })}
            </div>

            <div className="instructions-panel">
                <h3>How to Test</h3>
                <ol>
                    <li>Start the cluster (<code>go run main.go --role=leader --port=8000</code> and 4 workers).</li>
                    <li>Wait for all nodes to agree on <strong>Node 5</strong>.</li>
                    <li>Kill Node 5 in your terminal (<code>Ctrl+C</code>).</li>
                    <li>Watch the nodes enter 🚨 <strong>Election</strong> state and elect <strong>Node 4</strong>.</li>
                </ol>
            </div>
        </div>
    );
};

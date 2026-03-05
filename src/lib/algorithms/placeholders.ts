import { Message, NodeState } from '../types';

/**
 * Placeholder for Ricart-Agrawala algorithm for mutual exclusion.
 * Purpose: Ensure only one process accesses a shared resource at a time.
 */
export interface RicartAgrawala {
    requestAccess(nodeId: string): Promise<boolean>;
    releaseAccess(nodeId: string): void;
}

/**
 * Placeholder for Bully algorithm for leader election.
 * Purpose: Elect a new leader when the current leader fails.
 */
export interface BullyAlgorithm {
    onLeaderFailed(): void;
    electLeader(nodes: NodeState[]): string;
}

/**
 * Placeholder for Banker's algorithm for resource allocation.
 * Purpose: Avoid deadlock by checking if resource allocation is safe.
 */
export interface BankerAlgorithm {
    isSafeState(available: number[], max: number[][], allocation: number[][]): boolean;
}

/**
 * Placeholder for Consensus algorithm (e.g., Paxos or Raft-like voting).
 * Purpose: Reach agreement among nodes on a single value.
 */
export interface Consensus {
    proposeValue(value: any): Promise<boolean>;
    vote(senderId: string, value: any): boolean;
}

// Integration Hook Constants (to be used later)
export const ALGORITHM_INTEGRATION_POINTS = {
    MUTUAL_EXCLUSION: 'RICART_AGRAWALA',
    LEADER_ELECTION: 'BULLY',
    DEADLOCK_AVOIDANCE: 'BANKER',
    AGREEMENT: 'CONSENSUS',
};

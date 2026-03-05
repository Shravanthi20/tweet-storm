import { Event } from '../types';
import { LeaderNode } from '../leader/leader';
import { WorkerNode } from '../worker/worker';

export class Client {
    readonly id: string;

    constructor(id: string) {
        this.id = id;
    }

    /**
     * Simulate a user sending an event.
     * The event is sent directly via the Leader node as per the distributed architecture.
     */
    sendEvent(content: string, type: Event['type'], leader: LeaderNode, workers: WorkerNode[]) {
        const event: Event = {
            id: Math.random().toString(36).substring(7),
            type,
            content,
            timestamp: Date.now(),
        };

        // Forward the event to the leader
        leader.receiveClientEvent(event, workers);
    }
}

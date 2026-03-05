import { LamportClock } from '../algorithms/lamport_clock';
import { Event, NodeState } from '../types';
import { WorkerNode } from '../worker/worker';

export class LeaderNode {
    readonly id: string = 'Leader-Node';
    private clock: LamportClock;

    constructor() {
        this.clock = new LamportClock();
    }

    /**
     * Coordinator receives event from client.
     */
    receiveClientEvent(event: Event, workers: WorkerNode[]): { log: string; assignedWorkerId: string } {
        // Lamport Rule: Leader increments clock before coordinating assignment
        this.clock.send();

        const leaderTime = this.clock.time;

        // Assign the task to a worker
        const assignedWorker = this.assignEvent(event, workers);

        return {
            log: `Leader → received event, incremented clock to ${leaderTime}`,
            assignedWorkerId: assignedWorker?.id || 'none'
        };
    }

    private assignEvent(event: Event, workers: WorkerNode[]): WorkerNode | null {
        if (workers.length === 0) return null;

        // Current leader clock value to be attached to the message
        const leaderTime = this.clock.time;

        // Simple Random assignment for simulation
        const workerIndex = Math.floor(Math.random() * workers.length);
        const selectedWorker = workers[workerIndex];

        selectedWorker.receiveTask(event, leaderTime);
        return selectedWorker;
    }

    getState(): NodeState {
        return {
            id: this.id,
            type: 'LEADER',
            status: 'ACTIVE',
            clock: this.clock.time
        };
    }
}

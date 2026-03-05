import { LeaderNode } from './leader/leader';
import { WorkerNode } from './worker/worker';
import { Client } from './client/client';
import { Event, SimulationState } from './types';
import { sharedStorage, SharedStorage } from './shared/storage';

export class SimulationManager {
    private leader: LeaderNode;
    private workers: WorkerNode[];
    private client: Client;
    private storage: SharedStorage;
    private pipelineLogs: string[] = [];
    private isProcessing: boolean = false;
    private assignedWorkerId: string = 'none';

    constructor(workerCount: number = 3) {
        this.storage = sharedStorage;
        this.leader = new LeaderNode();
        this.client = new Client('Simulation-Client');
        this.workers = Array.from(
            { length: workerCount },
            (_, i) => new WorkerNode(`Worker-${i + 1}`, this.storage)
        );
    }

    /**
     * Entry point for the distributed simulation:
     * Client -> Leader -> Worker -> SharedStorage
     */
    simulateClientEvent(content: string, type: Event['type'] = 'tweet'): void {
        this.isProcessing = true;
        this.pipelineLogs = []; // Reset for new event

        // Step 1: Client generates and sends event to Leader
        const clientLog = `Client → sending event "${content}"`;
        this.pipelineLogs.push(clientLog);

        // Step 2-3: Coordination (Leader receive & assign)
        const coordination = this.leader.receiveClientEvent({
            id: Math.random().toString(36).substring(7),
            type,
            content,
            timestamp: Date.now(),
        }, this.workers);

        this.pipelineLogs.push(coordination.log);
        this.assignedWorkerId = coordination.assignedWorkerId;

        // For simulation purposes, we trigger processing immediately
        // In a real system, this would be asynchronous
        this.processPendingTasks();

        this.isProcessing = false;
    }

    private processPendingTasks() {
        this.workers.forEach(worker => {
            const logs = worker.processNextEvent();
            if (logs) {
                this.pipelineLogs.push(...logs);
            }
        });
    }

    getState(): SimulationState {
        return {
            leader: this.leader.getState(),
            workers: this.workers.map(w => w.getState()),
            events: this.storage.getEvents(),
            isRunning: true,
            pipelineLogs: this.pipelineLogs,
            isProcessing: this.isProcessing,
            assignedWorkerId: this.assignedWorkerId
        };
    }

    reset() {
        this.storage.clear();
        this.pipelineLogs = [];
        this.assignedWorkerId = 'none';
        this.leader = new LeaderNode();
        this.workers = this.workers.map(w => new WorkerNode(w.id, this.storage));
    }
}

export const simulationManager = new SimulationManager();

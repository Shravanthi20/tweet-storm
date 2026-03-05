import { LamportClock } from '../algorithms/lamport_clock';
import { Event, WorkerState } from '../types';
import { SharedStorage } from '../shared/storage';

export class WorkerNode {
    readonly id: string;
    private clock: LamportClock;
    private queue: Event[];
    private processedCount: number;
    private storage: SharedStorage;

    constructor(id: string, storage: SharedStorage) {
        this.id = id;
        this.clock = new LamportClock();
        this.queue = [];
        this.processedCount = 0;
        this.storage = storage;
    }

    // Receive a task from the leader
    receiveTask(event: Event, leaderLamportTime: number): void {
        // Lamport Rule: Receive message → merge clocks
        this.clock.receive(leaderLamportTime);

        // Lamport Rule: Processing the event is a local event → increment clock
        this.clock.tick();

        // Attach worker ID and updated Lamport timestamp
        const eventWithTimestamp: Event = {
            ...event,
            workerId: this.id,
            lamportTime: this.clock.time
        };

        // Add event to queue
        this.queue.push(eventWithTimestamp);

        // Ensure events stay ordered by Lamport timestamp
        this.queue.sort((a, b) => (a.lamportTime || 0) - (b.lamportTime || 0));
    }

    // Process the next event in the queue
    processNextEvent(): string[] | null {
        if (this.queue.length === 0) return null;

        const event = this.queue.shift() || null;

        if (event) {
            this.processedCount++;

            // Lamport Rule: processing is a local event
            this.clock.tick();

            // Push to shared storage for UI feed
            const storageLog = this.storage.addEvent(event);

            const workerLog = `${this.id} → processing event at Lamport time ${this.clock.time}`;
            return [workerLog, storageLog];
        }

        return null;
    }

    // Get worker state for UI display
    getState(): WorkerState {
        return {
            id: this.id,
            type: 'WORKER',
            status: 'ACTIVE',
            clock: this.clock.time,
            queueSize: this.queue.length,
            processedCount: this.processedCount
        };
    }
}
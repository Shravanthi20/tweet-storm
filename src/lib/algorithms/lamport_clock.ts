export class LamportClock {
    private count: number;

    constructor(initialValue: number = 0) {
        this.count = initialValue;
    }

    // Local event: increment clock
    tick(): number {
        this.count += 1;
        return this.count;
    }

    // Send message: tick and return current count
    send(): number {
        return this.tick();
    }

    // Receive message: clock = max(local_clock, received_timestamp) + 1
    receive(receivedTimestamp: number): number {
        this.count = Math.max(this.count, receivedTimestamp) + 1;
        return this.count;
    }

    get time(): number {
        return this.count;
    }
}

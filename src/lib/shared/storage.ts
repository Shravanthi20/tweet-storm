import { Event } from '../types';

/**
 * Simulated shared storage for processed events.
 * In a real Storm cluster, this would be a database or global state.
 */
export class SharedStorage {
    private processedEvents: Event[] = [];
    private wordCounts: Record<string, number> = {};

    addEvent(event: Event): string {
        this.processedEvents.push(event);

        // Sort all events by Lamport timestamp for the ordered feed
        this.processedEvents.sort((a, b) => (a.lamportTime || 0) - (b.lamportTime || 0));

        // Global word frequency map
        const words = event.content.toLowerCase().split(/\s+/);
        const processedWords: string[] = [];

        words.forEach(word => {
            // Clean up punctuation
            const cleanWord = word.replace(/[^a-z]/g, '');
            if (cleanWord.length > 3) {
                this.wordCounts[cleanWord] = (this.wordCounts[cleanWord] || 0) + 1;
                processedWords.push(cleanWord);
            }
        });

        return `SharedStorage → storing result, updated word counts for: ${processedWords.join(', ')}`;
    }

    getEvents(): Event[] {
        return [...this.processedEvents];
    }

    getWordCounts(): Record<string, number> {
        return { ...this.wordCounts };
    }

    clear() {
        this.processedEvents = [];
        this.wordCounts = {};
    }
}

export const sharedStorage = new SharedStorage();

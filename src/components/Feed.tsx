import React from 'react';
import { Event } from '../lib/types';
import { sharedStorage } from '../lib/shared/storage';

interface FeedProps {
    events?: Event[]; // Optional now that we read from storage
}

export const Feed: React.FC<FeedProps> = ({ events: propEvents }) => {
    const events = propEvents || sharedStorage.getEvents();
    return (
        <div className="flex-1 max-w-2xl border-x border-gray-800 min-h-screen bg-black text-white">
            <div className="sticky top-0 bg-black/80 backdrop-blur-md p-4 border-b border-gray-800 z-10">
                <h2 className="text-xl font-bold">Event Stream</h2>
                <p className="text-xs text-gray-500">Ordered by Lamport Timestamp</p>
            </div>

            <div className="divide-y divide-gray-800">
                {events.length === 0 ? (
                    <div className="p-10 text-center text-gray-500">
                        No events processed yet. Simulate one from the sidebar.
                    </div>
                ) : (
                    events.map((event) => (
                        <div key={event.id} className="p-4 hover:bg-gray-900/50 transition-colors">
                            <div className="flex gap-4">
                                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center font-bold">
                                    {event.workerId?.[event.workerId.length - 1] || 'W'}
                                </div>
                                <div className="flex-1">
                                    <div className="flex items-center gap-2 mb-1">
                                        <span className="font-bold">{event.workerId || 'Pending'}</span>
                                        <span className="text-gray-500 text-sm">@{event.type}</span>
                                        <span className="text-gray-500 text-sm">·</span>
                                        <span className="px-2 py-0.5 rounded-full bg-blue-900/20 text-blue-400 text-[10px] font-mono border border-blue-900/50">
                                            Lamport: {event.lamportTime}
                                        </span>
                                    </div>
                                    <p className="text-[15px] leading-relaxed mb-3">
                                        {event.content}
                                    </p>
                                    <div className="flex gap-6 text-gray-500 text-xs">
                                        <span>{new Date(event.timestamp).toLocaleTimeString()}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

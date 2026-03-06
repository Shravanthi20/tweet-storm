import React, { useState } from 'react';
import './TweetComposer.css';

export const TweetComposer = () => {
    const [tweetText, setTweetText] = useState('');
    const [status, setStatus] = useState<{ type: 'idle' | 'success' | 'error', message: string }>({
        type: 'idle',
        message: ''
    });

    const isBurstMode = tweetText.trim() === ''; // Temporary toggle if we want to send bursts or single
    const leaderIP = import.meta.env.VITE_LEADER_IP || 'localhost';

    const handleSend = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!tweetText.trim()) return;

        setStatus({ type: 'idle', message: 'Sending...' });

        try {
            const res = await fetch(`http://${leaderIP}:8000/tweet`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    Content: tweetText,
                }),
            });

            if (!res.ok) throw new Error('Failed to send tweet');

            setStatus({ type: 'success', message: 'Tweet sent successfully! Check timeline or MongoDB.' });
            setTweetText('');

            // Clear success message after 3 seconds
            setTimeout(() => setStatus({ type: 'idle', message: '' }), 3000);

        } catch (err) {
            console.error(err);
            setStatus({ type: 'error', message: 'Error sending tweet. Is the Leader running?' });
        }
    };

    // Keep the "Burst" button for the presentation if they want to demonstrate Ricart-Agrawala
    const handleSendBurst = async () => {
        setStatus({ type: 'idle', message: 'Sending Burst (4 Tweets)...' });

        const burstTweets = [
            "distributed systems are powerful " + Date.now(),
            "apache storm processes streams " + Date.now(),
            "go is great for concurrency " + Date.now(),
            "real time processing is fun " + Date.now(),
        ];

        try {
            await Promise.all(burstTweets.map(content =>
                fetch(`http://${leaderIP}:8000/tweet`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ Content: content }),
                })
            ));
            setStatus({ type: 'success', message: 'Burst sent successfully!' });
            setTimeout(() => setStatus({ type: 'idle', message: '' }), 3000);
        } catch (err) {
            console.error(err);
            setStatus({ type: 'error', message: 'Error sending burst.' });
        }
    }


    return (
        <div className="composer-container">
            <div className="composer-card">
                <h2>Compose Tweet</h2>
                <p className="subtitle">Broadcast a message directly to the TweetStorm cluster.</p>

                <form onSubmit={handleSend} className="composer-form">
                    <textarea
                        value={tweetText}
                        onChange={(e) => setTweetText(e.target.value)}
                        placeholder="What's happening in your distributed system?"
                        rows={4}
                        maxLength={280}
                        className="tweet-input"
                        autoFocus
                    />

                    <div className="char-count">
                        {tweetText.length}/280
                    </div>

                    <div className="composer-actions">
                        <button
                            type="button"
                            className="btn-burst"
                            onClick={handleSendBurst}
                            title="Sends 4 automated tweets instantly to demonstrate Mutual Exclusion"
                        >
                            🚀 Send Demo Burst
                        </button>

                        <button
                            type="submit"
                            className="btn-send"
                            disabled={!tweetText.trim() || status.message === 'Sending...'}
                        >
                            Send Tweet
                        </button>
                    </div>
                </form>

                {status.message && (
                    <div className={`status-message ${status.type}`}>
                        {status.message}
                    </div>
                )}
            </div>

            <div className="info-panel">
                <h3>Architecture Flow</h3>
                <ol>
                    <li>This client sends a POST request to <code>Leader ({leaderIP}:8000)</code></li>
                    <li>Leader saves the tweet to <strong>MongoDB</strong></li>
                    <li>Leader hashes the tweet and assigns it to a <strong>Worker Node</strong></li>
                    <li>Worker updates the <strong>Global Word Count</strong> using <em>Ricart-Agrawala Mutual Exclusion</em></li>
                </ol>
            </div>
        </div>
    );
};

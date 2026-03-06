import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { Dashboard } from './Dashboard';
import { RicartAgrawala } from './RicartAgrawala';
import './App.css';

type EventLog = {
  node: string;
  message: string;
  timestamp: number;
};

function Timeline() {
  const [events, setEvents] = useState<EventLog[]>([]);

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const leaderIP = import.meta.env.VITE_LEADER_IP || 'localhost';
        const res = await fetch(`http://${leaderIP}:8000/events`);
        if (!res.ok) return;
        const data = (await res.json()) as EventLog[] | null;
        setEvents(data || []);
      } catch {
        // Keep UI running even if backend is temporarily down.
      }
    };

    fetchEvents();
    const timer = setInterval(fetchEvents, 1000);
    return () => clearInterval(timer);
  }, []);

  return (
    <main style={{ maxWidth: 800, margin: '2rem auto', padding: '0 1rem' }}>
      <h1>TweetStorm Lamport Timeline</h1>
      {events.length === 0 ? (
        <p>No events yet. Start leader, workers, and client.</p>
      ) : (
        <ul>
          {events.map((event, idx) => (
            <li key={`${event.timestamp}-${idx}`}>
              [Clock {event.timestamp}] {event.node} -&gt; {event.message}
            </li>
          ))}
        </ul>
      )}
    </main>
  );
}

function App() {
  return (
    <BrowserRouter>
      <div className="app-container">
        <nav className="navbar">
          <div className="nav-brand">TweetStorm ⚡🐦</div>
          <div className="nav-links">
            <Link to="/" className="nav-link">Event Timeline</Link>
            <Link to="/dashboard" className="nav-link highlight">Hash Ring Simulation</Link>
            <Link to="/mutual-exclusion" className="nav-link" style={{ color: '#ffb300' }}>Mutual Exclusion</Link>
          </div>
        </nav>

        <Routes>
          <Route path="/" element={<Timeline />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/mutual-exclusion" element={<RicartAgrawala />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;

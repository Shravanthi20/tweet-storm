import { useEffect, useState } from 'react'
import './App.css'

type EventLog = {
  node: string
  message: string
  timestamp: number
}

function App() {
  const [events, setEvents] = useState<EventLog[]>([])

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const res = await fetch('http://localhost:8000/events')
        if (!res.ok) return
        const data = (await res.json()) as EventLog[]
        setEvents(data)
      } catch {
        // Keep UI running even if backend is temporarily down.
      }
    }

    fetchEvents()
    const timer = setInterval(fetchEvents, 1000)
    return () => clearInterval(timer)
  }, [])

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
  )
}

export default App

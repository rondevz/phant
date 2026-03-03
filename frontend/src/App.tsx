import {useEffect, useState} from 'react';
import './App.css';
import {EventsOff, EventsOn} from "../wailsjs/runtime/runtime";
import {
    DumpEventChannelName,
    GetCollectorStatus,
    GetRecentEvents,
    GetSetupDiagnostics,
} from "../wailsjs/go/main/App";

type DumpEvent = {
    id: string;
    timestamp: string;
    sourceType: string;
    projectRoot: string;
    isDd: boolean;
    payload: unknown;
};

type CollectorStatus = {
    running: boolean;
    socketPath: string;
    lastError: string;
    dropped: number;
};

type SetupDiagnostics = {
    generatedAt: string;
    phpFound: boolean;
    phpVersion: string;
    phpIniOutput: string;
    serviceManager: string;
    lastError: string;
};

const MAX_RENDERED_EVENTS = 500;

function App() {
    const [events, setEvents] = useState<DumpEvent[]>([]);
    const [channelName, setChannelName] = useState<string>('');
    const [status, setStatus] = useState<CollectorStatus | null>(null);
    const [diagnostics, setDiagnostics] = useState<SetupDiagnostics | null>(null);

    useEffect(() => {
        let activeChannel = '';

        const appendEvent = (event: DumpEvent) => {
            setEvents((prev) => {
                const next = [...prev, event];
                if (next.length <= MAX_RENDERED_EVENTS) {
                    return next;
                }
                return next.slice(next.length - MAX_RENDERED_EVENTS);
            });
        };

        const load = async () => {
            const [resolvedChannel, collectorStatus, recentEvents, setupDiagnostics] = await Promise.all([
                DumpEventChannelName(),
                GetCollectorStatus(),
                GetRecentEvents(MAX_RENDERED_EVENTS),
                GetSetupDiagnostics(),
            ]);

            activeChannel = resolvedChannel;
            setChannelName(resolvedChannel);
            setStatus(collectorStatus);
            setEvents(recentEvents);
            setDiagnostics(setupDiagnostics);

            EventsOn(resolvedChannel, (event: DumpEvent) => {
                appendEvent(event);
            });
        };

        void load();

        const interval = setInterval(() => {
            void GetCollectorStatus().then(setStatus);
        }, 2000);

        return () => {
            clearInterval(interval);
            if (activeChannel !== '') {
                EventsOff(activeChannel);
            }
        };
    }, []);

    const clearEvents = () => setEvents([]);
    const refreshDiagnostics = () => {
        void GetSetupDiagnostics().then(setDiagnostics);
    };

    return (
        <div id="app" className="app-shell">
            <header className="app-header">
                <h1>Phant Live Dumps</h1>
                <button className="btn" onClick={clearEvents}>Clear</button>
            </header>

            <section className="status-grid">
                <div><strong>Runtime channel:</strong> {channelName || 'loading...'}</div>
                <div><strong>Collector running:</strong> {status?.running ? 'yes' : 'no'}</div>
                <div><strong>Dropped events:</strong> {status?.dropped ?? 0}</div>
                <div><strong>Socket:</strong> {status?.socketPath || 'n/a'}</div>
                {status?.lastError ? <div><strong>Last error:</strong> {status.lastError}</div> : null}
            </section>

            <section className="status-grid diagnostics-grid">
                <div className="diagnostics-header">
                    <strong>Setup diagnostics</strong>
                    <button className="btn" onClick={refreshDiagnostics}>Refresh</button>
                </div>
                <div><strong>Generated:</strong> {diagnostics?.generatedAt || 'n/a'}</div>
                <div><strong>PHP found:</strong> {diagnostics?.phpFound ? 'yes' : 'no'}</div>
                <div><strong>PHP version:</strong> {diagnostics?.phpVersion || 'n/a'}</div>
                <div><strong>Service manager:</strong> {diagnostics?.serviceManager || 'unknown'}</div>
                {diagnostics?.lastError ? <div><strong>Diagnostics error:</strong> {diagnostics.lastError}</div> : null}
            </section>

            <section className="events-section">
                <div className="events-title">Events ({events.length})</div>
                <div className="events-list">
                    {events.length === 0 ? (
                        <div className="empty-state">No dumps yet. Trigger dump() or dd() in Laravel.</div>
                    ) : (
                        events.map((event) => (
                            <pre key={`${event.id}-${event.timestamp}`} className="event-item">
                                {JSON.stringify(event, null, 2)}
                            </pre>
                        ))
                    )}
                </div>
            </section>
        </div>
    );
}

export default App

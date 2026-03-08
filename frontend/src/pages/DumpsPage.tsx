import React from 'react';
import { ActionButton } from '@/components/ui/action-button';
import { JsonBox } from '@/components/ui/json-box';
import { PageCard } from '@/components/ui/page-card';
import { ValueRow } from '@/components/ui/value-row';
import type { CollectorStatus, DumpEvent } from '@/types';

const getCallsiteLabel = (event: DumpEvent): string | null => {
    const frame = event.trace?.[0];
    if (!frame?.file || !frame?.line) {
        return null;
    }

    if (frame.func) {
        return `${frame.file}:${frame.line} (${frame.func})`;
    }

    return `${frame.file}:${frame.line}`;
};

const DumpRow = React.memo(({ event }: { event: DumpEvent }) => {
    const callsiteLabel = getCallsiteLabel(event);

    return (
        <div className="space-y-2">
            {callsiteLabel ? (
                <div className="rounded-md border border-slate-800 bg-slate-950 px-3 py-2 text-xs text-slate-300">
                    <span className="text-slate-500">Callsite:</span> {callsiteLabel}
                </div>
            ) : null}
            <JsonBox value={event} />
        </div>
    );
});

const DumpList = React.memo(({ events }: { events: DumpEvent[] }) => {
    if (events.length === 0) {
        return <p className="text-sm text-slate-400">No dumps yet. Trigger `dump()` or `dd()` in Laravel.</p>;
    }

    return (
        <div className="space-y-3">
            {events.map((event) => (
                <DumpRow key={event.id} event={event} />
            ))}
        </div>
    );
});

export function DumpsPage({
    channelName,
    status,
    events,
    onClear,
}: {
    channelName: string;
    status: CollectorStatus | null;
    events: DumpEvent[];
    onClear: () => void;
}) {
    const getCallsiteLabel = (event: DumpEvent): string | null => {
        const frame = event.trace?.[0];
        if (!frame?.file || !frame?.line) {
            return null;
        }

        if (frame.func) {
            return `${frame.file}:${frame.line} (${frame.func})`;
        }

        return `${frame.file}:${frame.line}`;
    };

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">Live Dumps</h1>
                <p className="text-muted-foreground mt-2">
                    Monitor dump() and dd() calls from your Laravel application in real-time.
                </p>
            </div>
            
            <PageCard
                title="Runtime"
                actions={<ActionButton onClick={onClear}>Clear Events</ActionButton>}
            >
                <div className="grid gap-2 md:grid-cols-2">
                    <ValueRow label="Channel" value={channelName || 'loading...'} />
                    <ValueRow label="Collector" value={status?.running ? 'running' : 'stopped'} />
                    <ValueRow label="Dropped" value={String(status?.dropped ?? 0)} />
                    <ValueRow label="Socket" value={status?.socketPath || 'n/a'} copyable />
                    {status?.lastError ? <ValueRow label="Last error" value={status.lastError} /> : null}
                </div>
            </PageCard>

            <PageCard title={`Dumps (${events.length})`}>
                <DumpList events={events} />
            </PageCard>
        </div>
    );
}

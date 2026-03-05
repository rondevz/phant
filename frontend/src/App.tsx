import { useEffect, useState } from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import { BaseLayout } from './components/layout/BaseLayout';
import { EventsOffAll, EventsOn } from '../wailsjs/runtime/runtime';
import {
    ApplyValetLinuxRemediation,
    DumpEventChannelName,
    GetCollectorStatus,
    EnableCLIHook,
    GetRecentEvents,
    GetSetupDiagnostics,
    GetValetLinuxVerification,
} from '../wailsjs/go/main/App';

import type {
    CollectorStatus,
    DumpEvent,
    HookInstallResult,
    SetupDiagnostics,
    ValetLinuxRemediationResult,
    ValetLinuxVerification,
} from './types';

import { ThemeProvider } from './components/theme-provider';
import { PhpManagerPage } from './pages/PhpManagerPage';
import { ValetSitesPage } from './pages/ValetSitesPage';
import { ServicesPage } from './pages/ServicesPage';
import { SettingsPage } from './pages/SettingsPage';
import { DumpsPage } from './pages/DumpsPage';
import { SetupPage } from './pages/SetupPage';
import { ValetPage } from './pages/ValetPage';

const MAX_RENDERED_EVENTS = 500;

function App() {
    const [events, setEvents] = useState<DumpEvent[]>([]);
    const [channelName, setChannelName] = useState<string>('');
    const [status, setStatus] = useState<CollectorStatus | null>(null);
    const [diagnostics, setDiagnostics] = useState<SetupDiagnostics | null>(null);
    const [valetVerification, setValetVerification] = useState<ValetLinuxVerification | null>(null);
    const [hookResult, setHookResult] = useState<HookInstallResult | null>(null);
    const [installingHook, setInstallingHook] = useState(false);
    const [refreshingValet, setRefreshingValet] = useState(false);
    const [valetRemediationResult, setValetRemediationResult] = useState<ValetLinuxRemediationResult | null>(null);
    const [applyingValetRemediation, setApplyingValetRemediation] = useState(false);
    const [confirmValetRemediation, setConfirmValetRemediation] = useState(false);

    useEffect(() => {
        let disposed = false;
        let unsubscribe: (() => void) | null = null;

        EventsOffAll();

        const appendEvent = (event: DumpEvent) => {
            setEvents((prev) => {
                if (prev.some((existing) => existing.id === event.id)) {
                    return prev;
                }

                const next = [...prev, event];
                if (next.length <= MAX_RENDERED_EVENTS) {
                    return next;
                }

                return next.slice(next.length - MAX_RENDERED_EVENTS);
            });
        };

        const load = async () => {
            const [resolvedChannel, collectorStatus, recentEvents, setupDiagnostics, valetStatus] = await Promise.all([
                DumpEventChannelName(),
                GetCollectorStatus(),
                GetRecentEvents(MAX_RENDERED_EVENTS),
                GetSetupDiagnostics(),
                GetValetLinuxVerification(),
            ]);

            if (disposed) {
                return;
            }

            setChannelName(resolvedChannel);
            setStatus(collectorStatus);
            setEvents(recentEvents);
            setDiagnostics(setupDiagnostics);
            setValetVerification(valetStatus);

            unsubscribe = EventsOn(resolvedChannel, (event: DumpEvent) => {
                appendEvent(event);
            });
        };

        void load();

        const interval = setInterval(() => {
            if (disposed) {
                return;
            }

            void GetCollectorStatus().then((nextStatus) => {
                if (!disposed) {
                    setStatus(nextStatus);
                }
            });
        }, 2000);

        return () => {
            disposed = true;
            clearInterval(interval);

            if (unsubscribe !== null) {
                unsubscribe();
                unsubscribe = null;
            }

            EventsOffAll();
        };
    }, []);

    const clearEvents = () => setEvents([]);
    const refreshDiagnostics = () => {
        void GetSetupDiagnostics().then(setDiagnostics);
    };

    const refreshValetVerification = async () => {
        setRefreshingValet(true);
        try {
            const result = await GetValetLinuxVerification();
            setValetVerification(result);
        } finally {
            setRefreshingValet(false);
        }
    };

    const enableCLIHook = async () => {
        setInstallingHook(true);
        try {
            const result = await EnableCLIHook();
            setHookResult(result);
            const [latestDiagnostics, latestValetStatus] = await Promise.all([
                GetSetupDiagnostics(),
                GetValetLinuxVerification(),
            ]);
            setDiagnostics(latestDiagnostics);
            setValetVerification(latestValetStatus);
        } finally {
            setInstallingHook(false);
        }
    };

    const applyValetRemediation = async () => {
        setApplyingValetRemediation(true);
        try {
            const result = await ApplyValetLinuxRemediation(confirmValetRemediation);
            setValetRemediationResult(result);
            const latestValetStatus = await GetValetLinuxVerification();
            setValetVerification(latestValetStatus);
        } finally {
            setApplyingValetRemediation(false);
        }
    };

    return (
        <ThemeProvider defaultTheme="dark" storageKey="phant-ui-theme">
            <BaseLayout>
                <Routes>
                    <Route path="/" element={<Navigate to="/dumps" replace />} />
                    <Route
                        path="/php"
                    element={
                        <div className="space-y-8">
                            <PhpManagerPage />
                            <SetupPage
                                diagnostics={diagnostics}
                                hookResult={hookResult}
                                installingHook={installingHook}
                                onRefresh={refreshDiagnostics}
                                onEnable={enableCLIHook}
                            />
                        </div>
                    }
                />
                <Route path="/sites" element={<ValetSitesPage />} />
                <Route
                    path="/valet"
                    element={
                        <ValetPage
                            valetVerification={valetVerification}
                            refreshingValet={refreshingValet}
                            onRefresh={refreshValetVerification}
                            confirmValetRemediation={confirmValetRemediation}
                            onConfirm={setConfirmValetRemediation}
                            applyingValetRemediation={applyingValetRemediation}
                            onApply={applyValetRemediation}
                            valetRemediationResult={valetRemediationResult}
                        />
                    }
                />
                <Route path="/services" element={<ServicesPage />} />
                <Route
                    path="/dumps"
                    element={<DumpsPage channelName={channelName} status={status} events={events} onClear={clearEvents} />}
                />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="*" element={<Navigate to="/dumps" replace />} />
            </Routes>
        </BaseLayout>
        </ThemeProvider>
    );
}

export default App

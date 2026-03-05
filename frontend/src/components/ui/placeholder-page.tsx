export function PlaceholderPage({
    title,
    subtitle,
    nextSteps,
}: {
    title: string;
    subtitle: string;
    nextSteps: string[];
}) {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
                <p className="text-muted-foreground mt-2">{subtitle}</p>
            </div>
            <div className="rounded-md border border-slate-800 bg-slate-950 p-3">
                <p className="mb-2 text-xs uppercase tracking-wide text-slate-500">Planned MVP scope</p>
                <ul className="list-inside list-disc space-y-1 text-sm text-slate-300">
                    {nextSteps.map((item) => (
                        <li key={item}>{item}</li>
                    ))}
                </ul>
            </div>
        </div>
    );
}
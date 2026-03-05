export function JsonBox({ value }: { value: unknown }) {
    return (
        <pre className="overflow-x-auto rounded-md border border-border bg-muted/50 p-3 text-xs text-muted-foreground font-mono">
            {JSON.stringify(value, null, 2)}
        </pre>
    );
}

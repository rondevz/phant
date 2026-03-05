import { toast } from 'sonner';
import { cn } from '@/lib/utils';

export function ValueRow({ label, value, copyable = false }: { label: string; value: string; copyable?: boolean }) {
    const handleCopy = () => {
        if (!copyable || !value || value === 'n/a') return;
        navigator.clipboard.writeText(value);
        toast.success('Copied to clipboard', {
            description: value,
        });
    };

    return (
        <div className="flex items-center justify-between gap-4 rounded-md border border-border bg-muted/40 px-3 py-2 text-sm overflow-hidden">
            <span className="text-muted-foreground whitespace-nowrap shrink-0">{label}</span>
            <span 
                className={cn(
                    "text-right font-medium text-foreground truncate",
                    copyable && value !== 'n/a' && "cursor-pointer hover:text-primary hover:underline transition-colors"
                )} 
                title={value}
                onClick={handleCopy}
            >
                {value}
            </span>
        </div>
    );
}

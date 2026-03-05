import { type ReactNode } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function PageCard({ title, actions, children }: { title: string; actions?: ReactNode; children: ReactNode }) {
    return (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-base font-semibold">{title}</CardTitle>
                {actions && <div className="flex items-center gap-2">{actions}</div>}
            </CardHeader>
            <CardContent className="space-y-4 pt-4">
                {children}
            </CardContent>
        </Card>
    );
}

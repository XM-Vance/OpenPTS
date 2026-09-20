import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Rocket, CheckCircle2 } from 'lucide-react';

interface ModulePlaceholderProps {
  title: string;
  description: string;
  phase: string;
  features: string[];
}

// 业务模块占位页：导航已接通，页面与后端接口按阶段 2 计划逐个迁移。
export function ModulePlaceholder({
  title,
  description,
  phase,
  features,
}: ModulePlaceholderProps) {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{title}</h1>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>

      <Card className="border-dashed">
        <CardContent className="flex flex-col items-center gap-5 py-12 text-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-full bg-primary/10 text-primary">
            <Rocket className="h-7 w-7" aria-hidden />
          </div>
          <div className="space-y-1.5">
            <div className="flex items-center justify-center gap-2">
              <p className="font-semibold">模块建设中</p>
              <Badge variant="secondary">{phase}</Badge>
            </div>
            <p className="text-sm text-muted-foreground">
              导航已接通；本模块的页面与后端接口将按阶段 2 计划迁移。
            </p>
          </div>
          {features.length > 0 && (
            <div className="grid w-full max-w-2xl gap-2 sm:grid-cols-2">
              {features.map((f) => (
                <div
                  key={f}
                  className="flex items-center gap-2 rounded-md border bg-muted/40 px-3 py-2 text-left text-sm text-muted-foreground"
                >
                  <CheckCircle2 className="h-4 w-4 shrink-0 text-primary/70" aria-hidden />
                  {f}
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

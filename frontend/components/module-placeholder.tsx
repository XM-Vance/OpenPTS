import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

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

      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <CardTitle>模块开发中</CardTitle>
            <Badge variant="secondary">{phase}</Badge>
          </div>
          <CardDescription>
            导航已接通；本模块的页面与后端接口将按阶段 2 计划迁移。
          </CardDescription>
        </CardHeader>
        <CardContent>
          <p className="mb-2 text-sm font-medium">规划功能</p>
          <ul className="space-y-1.5 text-sm text-muted-foreground">
            {features.map((f) => (
              <li key={f} className="flex gap-2">
                <span className="text-foreground">·</span>
                {f}
              </li>
            ))}
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}

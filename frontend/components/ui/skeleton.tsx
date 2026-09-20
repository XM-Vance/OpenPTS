import * as React from 'react';
import { cn } from '@/lib/utils';

/** 骨架屏占位：呼吸动画在用户开启「减弱动态效果」时停用 */
const Skeleton = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn('animate-pulse rounded-md bg-muted motion-reduce:animate-none', className)}
      {...props}
    />
  ),
);
Skeleton.displayName = 'Skeleton';

export { Skeleton };

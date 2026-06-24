import * as React from 'react';
import { cn } from '@/lib/utils';

// 简化版 Label：不依赖 @radix-ui/react-label，使用原生 label。
// 如需更复杂的可访问性特性，阶段后期再切换到 radix。
const Label = React.forwardRef<HTMLLabelElement, React.LabelHTMLAttributes<HTMLLabelElement>>(
  ({ className, ...props }, ref) => (
    <label
      ref={ref}
      className={cn(
        'text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70',
        className,
      )}
      {...props}
    />
  ),
);
Label.displayName = 'Label';

export { Label };

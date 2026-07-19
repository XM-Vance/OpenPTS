'use client';

// 复制按钮：点击复制指定文本到剪贴板，带「已复制」反馈。
// 项目里此前无复制组件，本文件随 MCP 接入页一同新增。
import { useState } from 'react';
import { Check, Copy } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface CopyButtonProps {
  value: string;
  label?: string;
  className?: string;
}

export function CopyButton({ value, label = '复制', className }: CopyButtonProps) {
  const [copied, setCopied] = useState(false);

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // 剪贴板 API 在非 HTTPS / 非localhost 下可能被拒，降级提示
      setCopied(false);
    }
  };

  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      className={className}
      onClick={onCopy}
    >
      {copied ? (
        <>
          <Check className="mr-1 h-3.5 w-3.5 text-green-600" />
          已复制
        </>
      ) : (
        <>
          <Copy className="mr-1 h-3.5 w-3.5" />
          {label}
        </>
      )}
    </Button>
  );
}

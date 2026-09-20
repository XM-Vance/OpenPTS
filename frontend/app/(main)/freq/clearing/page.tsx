'use client';

// 整页懒加载：页面内容（含 recharts 图表）移至 ./_view，
// 此处用 next/dynamic({ssr:false}) 瘦包装，把整页 JS 从首屏剥离、按需加载。
import dynamic from 'next/dynamic';
import { ChartLoading } from '@/components/feedback';

const View = dynamic(() => import('./_view'), {
  ssr: false,
  loading: () => (
    <ChartLoading className="min-h-[60vh]" />
  ),
});

export default function Page() {
  return <View />;
}

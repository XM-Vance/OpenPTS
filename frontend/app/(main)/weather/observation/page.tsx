'use client';

// 外部气象观测独立页：风电场逐时风速 + 水库水文逐日观测。
// 从「气象数据」拆出，归入「发电侧预测」。
// 内容含 recharts，用 next/dynamic({ssr:false}) 瘦包装，按需加载、不进首屏。
import dynamic from 'next/dynamic';
import { ChartLoading } from '@/components/feedback';

const Observation = dynamic(() => import('../_observation').then((m) => m.WeatherObservation), {
  ssr: false,
  loading: () => (
    <ChartLoading className="min-h-[60vh]" />
  ),
});

export default function Page() {
  return <Observation />;
}

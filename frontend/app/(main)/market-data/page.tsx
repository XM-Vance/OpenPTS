'use client';

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

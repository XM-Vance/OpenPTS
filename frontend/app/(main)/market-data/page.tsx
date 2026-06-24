'use client';

import dynamic from 'next/dynamic';

const View = dynamic(() => import('./_view'), {
  ssr: false,
  loading: () => (
    <div className="flex min-h-[60vh] items-center justify-center text-sm text-muted-foreground">
      加载中…
    </div>
  ),
});

export default function Page() {
  return <View />;
}

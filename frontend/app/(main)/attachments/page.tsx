'use client';

// 整页懒加载：页面内容（含 recharts 图表）移至 ./_view，
// 此处用 next/dynamic({ssr:false}) 瘦包装，把整页 JS 从首屏剥离、按需加载。
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

import { redirect } from 'next/navigation';

// 根路径直接重定向到仪表盘。未登录时 AuthContext 会接管并跳到 /login。
export default function HomePage() {
  redirect('/dashboard');
}

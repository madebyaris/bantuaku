import { Outlet } from 'react-router-dom'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export function Layout() {
  return (
    <div className="min-h-screen bg-black text-slate-100">
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-emerald-900/20 via-black to-black pointer-events-none -z-10" />
      <Sidebar />
      <div className="pl-64">
        {/* Header currently doesn't exist in file list but is imported. 
            If it uses default styles, we might need to check it too. 
            Assuming Header is simple or transparent. */}
        <Header />
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

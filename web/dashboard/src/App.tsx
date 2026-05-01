import { Button, ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'

/** 根组件：仪表盘入口占位（后续接路由与布局）。 */
function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-slate-50 p-6">
        <h1 className="text-2xl font-semibold text-slate-800">
          NexusRouter 控制台
        </h1>
        <p className="text-slate-600">Tailwind v4 + Ant Design 基线已就绪。</p>
        <Button type="primary">Ant Design 按钮</Button>
      </div>
    </ConfigProvider>
  )
}

export default App

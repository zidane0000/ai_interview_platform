import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App.tsx'
import './i18n' // 初始化 i18n
import './utils/configTest' // Environment configuration test

// 開發時關閉 StrictMode 避免重複執行，生產環境可開啟
const useStrictMode = false;

createRoot(document.getElementById('root')!).render(
  useStrictMode ? (
    <StrictMode>
      <App />
    </StrictMode>
  ) : (
    <App />
  )
)

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.less'
import MainWindow from "./pages/MainWindow.tsx";

createRoot(document.getElementById('root')!).render(
  <StrictMode>
        <MainWindow/>
  </StrictMode>,
)

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.less'
import MainWindow from "./pages/MainWindow.tsx";
import {initBase} from "./ts/api.ts";

void initBase().catch((e) => console.error(`init api base failed: ${(e as Error).message}`));

createRoot(document.getElementById('root')!).render(
  <StrictMode>
        <MainWindow/>
  </StrictMode>,
)

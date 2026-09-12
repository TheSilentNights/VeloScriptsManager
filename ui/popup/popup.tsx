import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import PopupWindow from "./PopupWindow";

createRoot(document.getElementById('popup')!).render(
  <StrictMode>
        <PopupWindow/>
  </StrictMode>,
)
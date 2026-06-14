import ReactDOM from "react-dom/client";

import { AppRoot } from "./app/AppRoot";

const rootElement = document.getElementById("root");

if (!rootElement) {
  throw new Error("missing #root element");
}

ReactDOM.createRoot(rootElement).render(<AppRoot />);

export function App() {
  return (
    <main
      style={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        margin: 0,
        background:
          "linear-gradient(135deg, rgb(242 240 233), rgb(225 234 240))",
        color: "rgb(31 41 55)",
        fontFamily: "ui-sans-serif, system-ui, sans-serif",
      }}
    >
      <section
        style={{
          width: "min(32rem, calc(100vw - 2rem))",
          padding: "2rem",
          borderRadius: "1.25rem",
          background: "rgb(255 255 255 / 0.88)",
          boxShadow: "0 24px 80px rgb(15 23 42 / 0.12)",
        }}
      >
        <p
          style={{
            margin: 0,
            fontSize: "0.8rem",
            letterSpacing: "0.08em",
            textTransform: "uppercase",
            color: "rgb(71 85 105)",
          }}
        >
          Cartulary Bootstrap
        </p>
        <h1 style={{ marginBottom: "0.75rem", fontSize: "2rem" }}>
          Web shell is running.
        </h1>
        <p style={{ margin: 0, lineHeight: 1.6 }}>
          This is the bounded Vite + React bootstrap placeholder. Workbook
          behavior, API clients, and transport contracts have not been wired
          yet.
        </p>
      </section>
    </main>
  );
}

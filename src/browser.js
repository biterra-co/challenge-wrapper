const ENDPOINT = "/__biterra__/presence";
const HEARTBEAT_MS = 15_000;

async function report(action, keepalive = false) {
  try {
    await fetch(ENDPOINT, {
      method: "POST",
      credentials: "include",
      headers: { "content-type": "application/json" },
      body: JSON.stringify({ action }),
      keepalive,
    });
  } catch {
    // Presence telemetry must never interfere with the challenge.
  }
}

export const open = () => report("open");
export const heartbeat = () => report("heartbeat");
export function close() {
  const body = JSON.stringify({ action: "close" });
  try {
    if (navigator.sendBeacon?.(ENDPOINT, new Blob([body], { type: "application/json" }))) return;
  } catch {}
  void report("close", true);
}

export function autoStart() {
  let timer;
  const stop = () => { if (timer) clearInterval(timer); timer = undefined; };
  const start = () => {
    stop();
    if (document.visibilityState !== "visible") return;
    void open();
    timer = setInterval(() => { if (document.visibilityState === "visible") void heartbeat(); }, HEARTBEAT_MS);
  };
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "visible") start();
    else { stop(); close(); }
  });
  addEventListener("pagehide", () => { stop(); close(); });
  start();
  return stop;
}

if (typeof document !== "undefined" && document.currentScript?.dataset.autostart !== "false") autoStart();

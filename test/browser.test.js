import assert from "node:assert/strict";
import test from "node:test";

const listeners = new Map();
globalThis.document = {
  currentScript: { dataset: { autostart: "false" } },
  visibilityState: "visible",
  addEventListener(name, fn) { listeners.set(name, fn); },
};
globalThis.addEventListener = (name, fn) => listeners.set(name, fn);

const calls = [];
globalThis.fetch = async (url, init) => { calls.push({ url, init }); return new Response(); };
Object.defineProperty(globalThis, "navigator", { value: { sendBeacon: () => false }, configurable: true });
let interval;
globalThis.setInterval = (fn, ms) => { interval = { fn, ms }; return 1; };
globalThis.clearInterval = () => { interval = undefined; };

const wrapper = await import("../src/browser.js");

test("sends open and 15 second visible heartbeats", async () => {
  calls.length = 0;
  wrapper.autoStart();
  await Promise.resolve();
  assert.equal(JSON.parse(calls[0].init.body).action, "open");
  assert.equal(interval.ms, 15_000);
  interval.fn();
  await Promise.resolve();
  assert.equal(JSON.parse(calls[1].init.body).action, "heartbeat");
});

test("closes when hidden without throwing on telemetry failure", async () => {
  calls.length = 0;
  wrapper.autoStart();
  document.visibilityState = "hidden";
  listeners.get("visibilitychange")();
  await Promise.resolve();
  assert.equal(JSON.parse(calls.at(-1).init.body).action, "close");
  assert.equal(calls.at(-1).init.keepalive, true);
  document.visibilityState = "visible";
});

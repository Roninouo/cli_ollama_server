// frontend/app.ts
var qs = (sel, root = document) => {
  const el = root.querySelector(sel);
  if (!el) throw new Error(`Missing element: ${sel}`);
  return el;
};
var qsa = (sel, root = document) => Array.from(root.querySelectorAll(sel));
var token = document.body.dataset.token || "";
var msgCopied = document.body.dataset.msgCopied || "Copied";
var msgLoading = document.body.dataset.msgLoading || "Loading";
var msgWorking = document.body.dataset.msgWorking || "Working";
var msgSaved = document.body.dataset.msgSaved || "Saved";
var elToast = qs("#toast");
var elOut = qs("#out");
var emptyOutputText = elOut.textContent || "";
var elChipHost = qs("#chipHost");
var elChipMode = qs("#chipMode");
var elModelsBody = qs("#modelsBody");
var modelsHintText = (qs("#modelsHint").textContent || "").trim() || "\u2014";
var elModelSearch = qs("#modelSearch");
var elModelDatalist = qs("#modelDatalist");
var elHost = qs("#host");
var elLang = qs("#lang");
var elMode = qs("#mode");
var elOllamaExe = qs("#ollamaExe");
var elUnsafe = qs("#unsafe");
var elNoProxyAuto = qs("#noProxyAuto");
var elCfgPath = qs("#cfgPath");
var elRunModel = qs("#runModel");
var elPrompt = qs("#prompt");
var elPullModel = qs("#pullModel");
var btnTheme = qs("#btnTheme");
var btnList = qs("#btnList");
var btnRun = qs("#btnRun");
var btnPull = qs("#btnPull");
var btnSave = qs("#btnSave");
var btnCopy = qs("#btnCopy");
var btnClear = qs("#btnClear");
var btnWrap = qs("#btnWrap");
var toastTimer;
var busyCount = 0;
var models = [];
function errMsg(err) {
  const anyErr = err;
  return String(anyErr?.message || anyErr || err);
}
function showError(label, err) {
  const msg = errMsg(err);
  showOutput(label, { error: msg, exitCode: 1, output: "" });
  toast(msg);
}
function toast(message) {
  if (toastTimer) window.clearTimeout(toastTimer);
  elToast.textContent = message;
  elToast.hidden = false;
  toastTimer = window.setTimeout(() => {
    elToast.hidden = true;
  }, 1600);
}
function setBusy(busy) {
  busyCount += busy ? 1 : -1;
  if (busyCount < 0) busyCount = 0;
  const on = busyCount > 0;
  for (const b of [btnList, btnRun, btnPull, btnSave]) b.disabled = on;
}
async function apiPost(path, body = {}) {
  const res = await fetch(`${path}?t=${encodeURIComponent(token)}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Token": token },
    body: JSON.stringify(body)
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}
async function apiGetConfig() {
  const res = await fetch(`/api/config?t=${encodeURIComponent(token)}`, {
    headers: { "X-Token": token }
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}
function formatOutput(label, data) {
  const lines = [];
  if (label) lines.push(`[${label}]`);
  if (data.error) lines.push(`ERROR: ${data.error}`);
  if (typeof data.exitCode === "number") lines.push(`exitCode: ${data.exitCode}`);
  if (lines.length) lines.push("");
  if (data.output) lines.push(String(data.output));
  return lines.join("\n");
}
function showOutput(label, data) {
  const next = formatOutput(label, data).trimEnd();
  if (next) elOut.textContent = next;
}
function uniq(list) {
  const out = [];
  const seen = /* @__PURE__ */ new Set();
  for (const s of list) {
    if (!s || seen.has(s)) continue;
    seen.add(s);
    out.push(s);
  }
  return out;
}
function parseModelsFromListOutput(output) {
  const lines = String(output || "").split(/\r?\n/).map((l) => l.trimEnd()).filter((l) => l.trim().length > 0);
  if (!lines.length) return [];
  const header = lines[0].trim();
  const start = /^NAME\s+/i.test(header) ? 1 : 0;
  const rows = [];
  for (let i = start; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;
    const parts = line.split(/\s{2,}/g).map((p) => p.trim()).filter(Boolean);
    if (parts.length >= 4) {
      rows.push({ name: parts[0], id: parts[1], size: parts[2], modified: parts.slice(3).join(" ") });
      continue;
    }
    const m = line.match(/^(\S+)/);
    if (m && m[1]) rows.push({ name: m[1] });
  }
  const names = uniq(rows.map((r) => r.name));
  const byName = /* @__PURE__ */ new Map();
  for (const r of rows) {
    if (!byName.has(r.name)) byName.set(r.name, r);
  }
  return names.map((n) => byName.get(n)).filter(Boolean);
}
function modelNames() {
  return models.map((m) => m.name).filter(Boolean);
}
var MODELS_COLS = 5;
function renderModelsTable(filter = "") {
  const f = filter.trim().toLowerCase();
  const rows = f ? models.filter((m) => `${m.name} ${m.id || ""}`.toLowerCase().includes(f)) : models.slice();
  elModelsBody.innerHTML = "";
  if (!rows.length) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.className = "muted";
    td.colSpan = MODELS_COLS;
    td.textContent = models.length ? "\u2014" : modelsHintText;
    tr.appendChild(td);
    elModelsBody.appendChild(tr);
    return;
  }
  for (const m of rows) {
    const tr = document.createElement("tr");
    const tdName = document.createElement("td");
    const nameBtn = document.createElement("button");
    nameBtn.type = "button";
    nameBtn.className = "cellbtn";
    nameBtn.title = m.name;
    nameBtn.setAttribute("aria-label", `Open run with model ${m.name}`);
    const title = document.createElement("div");
    title.className = "cell__title";
    title.textContent = m.name;
    nameBtn.appendChild(title);
    if (m.id) {
      const sub = document.createElement("div");
      sub.className = "cell__sub hide-sm";
      sub.textContent = m.id;
      nameBtn.appendChild(sub);
    }
    nameBtn.addEventListener("click", () => {
      elRunModel.value = m.name;
      elPullModel.value = m.name;
      location.hash = "#run";
      elPrompt.focus();
    });
    tdName.appendChild(nameBtn);
    const tdId = document.createElement("td");
    tdId.className = "hide-sm mono";
    tdId.textContent = m.id || "\u2014";
    const tdSize = document.createElement("td");
    tdSize.className = "mono";
    tdSize.textContent = m.size || "\u2014";
    const tdMod = document.createElement("td");
    tdMod.className = "hide-sm mono";
    tdMod.textContent = m.modified || "\u2014";
    const tdAct = document.createElement("td");
    tdAct.className = "actions";
    const btnRunRow = document.createElement("button");
    btnRunRow.type = "button";
    btnRunRow.className = "btn btn--ghost btn--sm";
    btnRunRow.textContent = btnRun.textContent || "Run";
    btnRunRow.title = `${btnRunRow.textContent} ${m.name}`;
    btnRunRow.setAttribute("aria-label", `Run model ${m.name}`);
    btnRunRow.addEventListener("click", (e) => {
      e.stopPropagation();
      elRunModel.value = m.name;
      location.hash = "#run";
      elPrompt.focus();
    });
    const btnPullRow = document.createElement("button");
    btnPullRow.type = "button";
    btnPullRow.className = "btn btn--ghost btn--sm";
    btnPullRow.textContent = btnPull.textContent || "Pull";
    btnPullRow.title = `${btnPullRow.textContent} ${m.name}`;
    btnPullRow.setAttribute("aria-label", `Pull model ${m.name}`);
    btnPullRow.addEventListener("click", (e) => {
      e.stopPropagation();
      elPullModel.value = m.name;
      location.hash = "#pull";
    });
    tdAct.appendChild(btnRunRow);
    tdAct.appendChild(btnPullRow);
    tr.appendChild(tdName);
    tr.appendChild(tdId);
    tr.appendChild(tdSize);
    tr.appendChild(tdMod);
    tr.appendChild(tdAct);
    elModelsBody.appendChild(tr);
  }
}
function renderModelDatalist() {
  elModelDatalist.innerHTML = "";
  for (const name of modelNames()) {
    const opt = document.createElement("option");
    opt.value = name;
    elModelDatalist.appendChild(opt);
  }
}
function setRoute(route) {
  const r = route || "models";
  for (const btn of qsa(".nav__item")) {
    const active = btn.dataset.route === r;
    btn.setAttribute("aria-current", active ? "page" : "false");
  }
  for (const view of qsa("[data-view]")) {
    view.hidden = view.id !== `view-${r}`;
  }
}
function currentRoute() {
  const h = (location.hash || "").replace(/^#/, "").trim();
  return h || "models";
}
function applyTheme(next) {
  const html = document.documentElement;
  if (next) {
    html.setAttribute("data-theme", next);
    localStorage.setItem("ollama-remote.ui.theme", next);
    return;
  }
  html.removeAttribute("data-theme");
  localStorage.removeItem("ollama-remote.ui.theme");
}
function toggleTheme() {
  const cur = document.documentElement.getAttribute("data-theme");
  if (!cur) applyTheme("dark");
  else if (cur === "dark") applyTheme("light");
  else applyTheme(null);
}
async function loadConfig() {
  const data = await apiGetConfig();
  elCfgPath.textContent = data.configPath || "";
  elHost.value = data.host || "";
  elLang.value = data.lang || "en";
  elMode.value = data.mode || "auto";
  elOllamaExe.value = data.ollamaExe || "";
  elUnsafe.checked = !!data.unsafe;
  elNoProxyAuto.checked = !!data.noProxyAuto;
  elChipHost.textContent = data.host ? `host: ${data.host}` : "host: \u2014";
  elChipMode.textContent = data.selectedMode ? `mode: ${data.selectedMode}` : `mode: ${data.mode || "auto"}`;
}
async function refreshModels() {
  setBusy(true);
  toast(msgLoading);
  try {
    const data = await apiPost("/api/list");
    showOutput("list", data);
    models = parseModelsFromListOutput(data.output || "");
    renderModelDatalist();
    renderModelsTable(elModelSearch.value);
  } finally {
    setBusy(false);
  }
}
async function runPrompt() {
  const model = elRunModel.value.trim();
  const prompt = elPrompt.value;
  setBusy(true);
  toast(msgWorking);
  try {
    localStorage.setItem("ollama-remote.ui.lastModel", model);
    localStorage.setItem("ollama-remote.ui.lastPrompt", prompt);
    const data = await apiPost("/api/run", { model, prompt });
    showOutput("run", data);
  } finally {
    setBusy(false);
  }
}
async function pullModel() {
  const model = elPullModel.value.trim();
  setBusy(true);
  toast(msgWorking);
  try {
    localStorage.setItem("ollama-remote.ui.lastModel", model);
    const data = await apiPost("/api/pull", { model });
    showOutput("pull", data);
  } finally {
    setBusy(false);
  }
}
async function saveConfig() {
  setBusy(true);
  toast(msgWorking);
  try {
    await apiPost("/api/config/set", {
      host: elHost.value,
      lang: elLang.value,
      mode: elMode.value,
      ollamaExe: elOllamaExe.value,
      unsafe: elUnsafe.checked,
      noProxyAuto: elNoProxyAuto.checked
    });
    await loadConfig();
    toast(msgSaved);
    elOut.textContent = msgSaved;
  } finally {
    setBusy(false);
  }
}
async function copyOutput() {
  const text = elOut.textContent || "";
  if (!text.trim()) return;
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    const ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    document.execCommand("copy");
    ta.remove();
  }
  toast(msgCopied);
}
function clearOutput() {
  elOut.textContent = emptyOutputText;
}
function toggleWrap() {
  const wrapLabel = btnWrap.dataset.labelWrap || btnWrap.textContent || "Wrap";
  const unwrapLabel = btnWrap.dataset.labelUnwrap || wrapLabel;
  const on = document.body.classList.toggle("wrap-output");
  btnWrap.textContent = on ? unwrapLabel : wrapLabel;
}
function restoreDrafts() {
  const lastModel = localStorage.getItem("ollama-remote.ui.lastModel") || "";
  const lastPrompt = localStorage.getItem("ollama-remote.ui.lastPrompt") || "";
  if (lastModel && !elRunModel.value) elRunModel.value = lastModel;
  if (lastModel && !elPullModel.value) elPullModel.value = lastModel;
  if (lastPrompt && !elPrompt.value) elPrompt.value = lastPrompt;
}
function wire() {
  for (const btn of qsa(".nav__item")) {
    btn.addEventListener("click", () => {
      location.hash = `#${btn.dataset.route}`;
    });
  }
  window.addEventListener("hashchange", () => setRoute(currentRoute()));
  setRoute(currentRoute());
  btnTheme.addEventListener("click", toggleTheme);
  btnList.addEventListener("click", () => refreshModels().catch((e) => showError("list", e)));
  btnRun.addEventListener("click", () => runPrompt().catch((e) => showError("run", e)));
  btnPull.addEventListener("click", () => pullModel().catch((e) => showError("pull", e)));
  btnSave.addEventListener("click", () => saveConfig().catch((e) => showError("config", e)));
  btnCopy.addEventListener("click", copyOutput);
  btnClear.addEventListener("click", clearOutput);
  btnWrap.addEventListener("click", toggleWrap);
  elModelSearch.addEventListener("input", () => renderModelsTable(elModelSearch.value));
  elPrompt.addEventListener("keydown", (e) => {
    if ((e.ctrlKey || e.metaKey) && e.key === "Enter") {
      e.preventDefault();
      runPrompt().catch((err) => showError("run", err));
    }
  });
}
function boot() {
  const savedTheme = localStorage.getItem("ollama-remote.ui.theme");
  if (savedTheme === "dark" || savedTheme === "light") applyTheme(savedTheme);
  wire();
  loadConfig().catch((err) => {
    elOut.textContent = `ERROR: ${String(err?.message || err)}`;
  });
  restoreDrafts();
}
boot();

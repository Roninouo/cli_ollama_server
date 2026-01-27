type ApiExecResponse = {
  output?: string;
  exitCode?: number;
  error?: string;
};

type ConfigResponse = {
  configPath?: string;
  host?: string;
  lang?: string;
  mode?: string;
  unsafe?: boolean;
  noProxyAuto?: boolean;
  ollamaExe?: string;
  selectedMode?: string;
};

type ModelRow = {
  name: string;
  id?: string;
  size?: string;
  modified?: string;
};

const qs = <T extends Element>(sel: string, root: Document | Element = document): T => {
  const el = root.querySelector(sel);
  if (!el) throw new Error(`Missing element: ${sel}`);
  return el as T;
};

const qsa = <T extends Element>(sel: string, root: Document | Element = document): T[] =>
  Array.from(root.querySelectorAll(sel)) as T[];

const token = document.body.dataset.token || "";
const msgCopied = document.body.dataset.msgCopied || "Copied";
const msgLoading = document.body.dataset.msgLoading || "Loading";
const msgWorking = document.body.dataset.msgWorking || "Working";
const msgSaved = document.body.dataset.msgSaved || "Saved";

const elToast = qs<HTMLDivElement>("#toast");
const elOut = qs<HTMLElement>("#out");
const emptyOutputText = elOut.textContent || "";
const elChipHost = qs<HTMLElement>("#chipHost");
const elChipMode = qs<HTMLElement>("#chipMode");

const elModelsBody = qs<HTMLTableSectionElement>("#modelsBody");
const modelsHintText = (qs<HTMLElement>("#modelsHint").textContent || "").trim() || "—";
const elModelSearch = qs<HTMLInputElement>("#modelSearch");
const elModelDatalist = qs<HTMLDataListElement>("#modelDatalist");

const elHost = qs<HTMLInputElement>("#host");
const elLang = qs<HTMLSelectElement>("#lang");
const elMode = qs<HTMLSelectElement>("#mode");
const elOllamaExe = qs<HTMLInputElement>("#ollamaExe");
const elUnsafe = qs<HTMLInputElement>("#unsafe");
const elNoProxyAuto = qs<HTMLInputElement>("#noProxyAuto");
const elCfgPath = qs<HTMLElement>("#cfgPath");

const elRunModel = qs<HTMLInputElement>("#runModel");
const elPrompt = qs<HTMLTextAreaElement>("#prompt");
const elPullModel = qs<HTMLInputElement>("#pullModel");

const btnTheme = qs<HTMLButtonElement>("#btnTheme");
const btnList = qs<HTMLButtonElement>("#btnList");
const btnRun = qs<HTMLButtonElement>("#btnRun");
const btnPull = qs<HTMLButtonElement>("#btnPull");
const btnSave = qs<HTMLButtonElement>("#btnSave");
const btnCopy = qs<HTMLButtonElement>("#btnCopy");
const btnClear = qs<HTMLButtonElement>("#btnClear");
const btnWrap = qs<HTMLButtonElement>("#btnWrap");

let toastTimer: number | undefined;
let busyCount = 0;
let models: ModelRow[] = [];

function errMsg(err: unknown): string {
  const anyErr = err as any;
  return String(anyErr?.message || anyErr || err);
}

function showError(label: string, err: unknown) {
  const msg = errMsg(err);
  showOutput(label, { error: msg, exitCode: 1, output: "" });
  toast(msg);
}

function toast(message: string) {
  if (toastTimer) window.clearTimeout(toastTimer);
  elToast.textContent = message;
  elToast.hidden = false;
  toastTimer = window.setTimeout(() => {
    elToast.hidden = true;
  }, 1600);
}

function setBusy(busy: boolean) {
  busyCount += busy ? 1 : -1;
  if (busyCount < 0) busyCount = 0;
  const on = busyCount > 0;
  for (const b of [btnList, btnRun, btnPull, btnSave]) b.disabled = on;
}

async function apiPost(path: string, body: any = {}): Promise<ApiExecResponse> {
  const res = await fetch(`${path}?t=${encodeURIComponent(token)}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-Token": token },
    body: JSON.stringify(body)
  });
  const data = (await res.json().catch(() => ({}))) as ApiExecResponse;
  if (!res.ok) throw new Error((data as any).error || `HTTP ${res.status}`);
  return data;
}

async function apiGetConfig(): Promise<ConfigResponse> {
  const res = await fetch(`/api/config?t=${encodeURIComponent(token)}`, {
    headers: { "X-Token": token }
  });
  const data = (await res.json().catch(() => ({}))) as ConfigResponse;
  if (!res.ok) throw new Error((data as any).error || `HTTP ${res.status}`);
  return data;
}

function formatOutput(label: string, data: ApiExecResponse): string {
  const lines: string[] = [];
  if (label) lines.push(`[${label}]`);
  if (data.error) lines.push(`ERROR: ${data.error}`);
  if (typeof data.exitCode === "number") lines.push(`exitCode: ${data.exitCode}`);
  if (lines.length) lines.push("");
  if (data.output) lines.push(String(data.output));
  return lines.join("\n");
}

function showOutput(label: string, data: ApiExecResponse) {
  const next = formatOutput(label, data).trimEnd();
  if (next) elOut.textContent = next;
}

function uniq(list: string[]): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const s of list) {
    if (!s || seen.has(s)) continue;
    seen.add(s);
    out.push(s);
  }
  return out;
}

function parseModelsFromListOutput(output: string): ModelRow[] {
  const lines = String(output || "")
    .split(/\r?\n/)
    .map((l) => l.trimEnd())
    .filter((l) => l.trim().length > 0);

  if (!lines.length) return [];

  const header = lines[0].trim();
  const start = /^NAME\s+/i.test(header) ? 1 : 0;

  const rows: ModelRow[] = [];
  for (let i = start; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;

    // Ollama CLI output is column-aligned; split on 2+ spaces.
    const parts = line.split(/\s{2,}/g).map((p) => p.trim()).filter(Boolean);

    // Expected: NAME, ID, SIZE, MODIFIED
    if (parts.length >= 4) {
      rows.push({ name: parts[0], id: parts[1], size: parts[2], modified: parts.slice(3).join(" ") });
      continue;
    }
    // Fallback: first token as name.
    const m = line.match(/^(\S+)/);
    if (m && m[1]) rows.push({ name: m[1] });
  }

  const names = uniq(rows.map((r) => r.name));
  const byName = new Map<string, ModelRow>();
  for (const r of rows) {
    if (!byName.has(r.name)) byName.set(r.name, r);
  }
  return names.map((n) => byName.get(n)!).filter(Boolean);
}

function modelNames(): string[] {
  return models.map((m) => m.name).filter(Boolean);
}

const MODELS_COLS = 5;

function renderModelsTable(filter = "") {
  const f = filter.trim().toLowerCase();
  const rows = f
    ? models.filter((m) => `${m.name} ${m.id || ""}`.toLowerCase().includes(f))
    : models.slice();
  elModelsBody.innerHTML = "";
  if (!rows.length) {
    const tr = document.createElement("tr");
    const td = document.createElement("td");
    td.className = "muted";
    td.colSpan = MODELS_COLS;
    td.textContent = models.length ? "—" : modelsHintText;
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
    tdId.textContent = m.id || "—";

    const tdSize = document.createElement("td");
    tdSize.className = "mono";
    tdSize.textContent = m.size || "—";

    const tdMod = document.createElement("td");
    tdMod.className = "hide-sm mono";
    tdMod.textContent = m.modified || "—";

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

function setRoute(route: string) {
  const r = route || "models";
  for (const btn of qsa<HTMLButtonElement>(".nav__item")) {
    const active = btn.dataset.route === r;
    btn.setAttribute("aria-current", active ? "page" : "false");
  }
  for (const view of qsa<HTMLElement>("[data-view]")) {
    view.hidden = view.id !== `view-${r}`;
  }
}

function currentRoute(): string {
  const h = (location.hash || "").replace(/^#/, "").trim();
  return h || "models";
}

function applyTheme(next: "dark" | "light" | null) {
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

  elChipHost.textContent = data.host ? `host: ${data.host}` : "host: —";
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
  for (const btn of qsa<HTMLButtonElement>(".nav__item")) {
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
    elOut.textContent = `ERROR: ${String((err as any)?.message || err)}`;
  });
  restoreDrafts();
}

boot();

/* IPA Downloader – frontend (vanilla JS, Tailwind) */
(() => {
  const $ = (sel, root = document) => root.querySelector(sel);
  const $$ = (sel, root = document) => Array.from(root.querySelectorAll(sel));

  const API = () => window.go && window.go.main && window.go.main.App;
  const RT = () => window.runtime;

  const esc = (s) => String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  const fmtBytes = (n) => {
    n = Number(n) || 0;
    if (n < 1024) return n + " B";
    const u = ["KB", "MB", "GB"];
    let i = -1;
    do { n /= 1024; i++; } while (n >= 1024 && i < u.length - 1);
    return n.toFixed(n < 10 ? 1 : 0) + " " + u[i];
  };
  const fmtSpeed = (kb) => (kb >= 1024 ? (kb / 1024).toFixed(1) + " MB/s" : Math.round(kb) + " KB/s");
  const fmtDate = (iso) => { try { return new Date(iso).toLocaleString("vi-VN"); } catch { return ""; } };

  let toastTimer;
  const toast = (msg, kind = "info") => {
    const el = $("#toast");
    el.textContent = msg;
    el.classList.toggle("ring-rose-500/40", kind === "error");
    el.classList.toggle("ring-emerald-500/40", kind === "ok");
    el.style.opacity = "1";
    clearTimeout(toastTimer);
    toastTimer = setTimeout(() => (el.style.opacity = "0"), 2800);
  };

  const iconHTML = (src, size = "h-12 w-12") =>
    src
      ? `<img src="${esc(src)}" class="${size} shrink-0 rounded-xl bg-ink-800 object-cover" loading="lazy" />`
      : `<div class="${size} shrink-0 rounded-xl bg-ink-800 flex items-center justify-center text-slate-600"><svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="4" y="4" width="16" height="16" rx="4"/></svg></div>`;

  // ---------------- state ----------------
  const state = {
    status: null,
    results: [],
    downloads: new Map(), // id -> progress
    library: [],
    batch: [],
  };

  // ---------------- views ----------------
  function showLogin() {
    $("#view-main").classList.add("hidden");
    $("#view-login").classList.remove("hidden");
    $("#view-login").classList.add("flex");
    setTimeout(() => $("#login-email").focus(), 50);
  }
  function showMain() {
    $("#view-login").classList.add("hidden");
    $("#view-login").classList.remove("flex");
    $("#view-main").classList.remove("hidden");
    $("#view-main").classList.add("flex");
    renderAccount();
    switchTab("search");
    loadLibrary();
  }

  function switchTab(name) {
    $$(".nav-item").forEach((b) => b.classList.toggle("active", b.dataset.tab === name));
    $$(".tab").forEach((t) => {
      const on = t.id === "tab-" + name;
      t.classList.toggle("hidden", !on);
      t.classList.toggle("flex", on);
    });
    if (name === "library") loadLibrary();
    if (name === "search") setTimeout(() => $("#search-input").focus(), 30);
  }
  $$(".nav-item").forEach((b) => b.addEventListener("click", () => switchTab(b.dataset.tab)));

  function renderAccount() {
    const s = state.status || {};
    $("#acct-name").textContent = s.name || s.email || "";
    $("#acct-email").textContent = s.email || "";
    $("#acct-country").textContent = s.country || "";
    $("#acct-avatar").textContent = ((s.name || s.email || "?")[0] || "?").toUpperCase();
    $("#app-version").textContent = "v" + (s.version || "");
    $("#library-dir").textContent = s.outputDir || "";
  }

  // ---------------- login ----------------
  $("#login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const btn = $("#login-btn");
    const err = $("#login-error");
    err.classList.add("hidden");
    btn.disabled = true;
    btn.textContent = "Đang đăng nhập…";
    try {
      const res = await API().Login($("#login-email").value, $("#login-password").value, $("#login-code").value);
      if (res.needCode) {
        $("#login-code-wrap").classList.remove("hidden");
        $("#login-code").value = "";
        $("#login-code").focus();
        btn.textContent = "Xác nhận mã";
        toast("Nhập mã 2FA Apple vừa gửi về thiết bị của bạn");
      } else {
        state.status = await API().GetStatus();
        $("#login-password").value = "";
        $("#login-code").value = "";
        $("#login-code-wrap").classList.add("hidden");
        showMain();
        toast(`Đã đăng nhập: ${res.email} (${res.country})`, "ok");
      }
    } catch (ex) {
      err.textContent = String(ex);
      err.classList.remove("hidden");
      btn.textContent = $("#login-code-wrap").classList.contains("hidden") ? "Đăng nhập" : "Xác nhận mã";
    } finally {
      btn.disabled = false;
      if (btn.textContent === "Đang đăng nhập…") btn.textContent = "Đăng nhập";
    }
  });

  $("#logout-btn").addEventListener("click", async () => {
    if (!confirm("Đăng xuất Apple ID khỏi ứng dụng?")) return;
    try { await API().Logout(); } catch {}
    state.status = null;
    showLogin();
  });

  // ---------------- search ----------------
  function appCard(app, opts = {}) {
    const dl = state.downloads.get(app.id);
    const busy = dl && !["done", "error", "cancelled"].includes(dl.status);
    const paid = app.price > 0;
    return `
      <div class="card" data-id="${app.id}">
        ${iconHTML(app.icon)}
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <div class="truncate text-sm font-medium text-white">${esc(app.name)}</div>
            ${paid ? `<span class="pill bg-amber-500/15 text-amber-300">$${app.price}</span>` : ""}
          </div>
          <div class="mt-0.5 truncate text-xs text-slate-500">${esc(app.developer || "")}${app.developer ? " · " : ""}${esc(app.bundleId)}</div>
          <div class="mt-1.5 flex flex-wrap gap-1.5">
            <span class="pill">v${esc(app.version || "?")}</span>
            ${app.sizeBytes ? `<span class="pill">${fmtBytes(app.sizeBytes)}</span>` : ""}
            ${app.minOs ? `<span class="pill">iOS ${esc(app.minOs)}+</span>` : ""}
            ${app.genre ? `<span class="pill">${esc(app.genre)}</span>` : ""}
            ${app.rating ? `<span class="pill">★ ${app.rating.toFixed(1)}</span>` : ""}
          </div>
        </div>
        <div class="flex shrink-0 flex-col items-end gap-1">
          <button class="btn-primary dl-btn px-3 py-1.5 text-xs" ${busy ? "disabled" : ""} title="${paid ? "App trả phí: tài khoản phải đã mua trên iPhone/App Store" : ""}">
            ${busy ? "Đang tải…" : "Tải IPA"}
          </button>
          <span class="text-[10px] text-slate-600">ID ${app.id}</span>
        </div>
      </div>`;
  }

  $("#search-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const term = $("#search-input").value.trim();
    if (!term) return;
    const btn = $("#search-btn");
    btn.disabled = true;
    btn.textContent = "…";
    const box = $("#search-results");
    box.innerHTML = `<div class="empty-state"><div class="spinner"></div><p>Đang tìm "${esc(term)}"…</p></div>`;
    try {
      let results;
      // Nhập đúng bundle ID hoặc app ID -> tra cứu chính xác
      if (/^\d{6,}$/.test(term) || /^[a-z0-9-]+(\.[a-z0-9-]+){2,}$/i.test(term)) {
        try {
          results = [await API().Lookup(term)];
        } catch {
          results = await API().Search(term, Number($("#search-limit").value));
        }
      } else {
        results = await API().Search(term, Number($("#search-limit").value));
      }
      state.results = results || [];
      renderResults();
    } catch (ex) {
      box.innerHTML = `<div class="empty-state"><p class="text-rose-300">${esc(ex)}</p></div>`;
    } finally {
      btn.disabled = false;
      btn.textContent = "Tìm";
    }
  });

  function renderResults() {
    const box = $("#search-results");
    if (!state.results.length) {
      box.innerHTML = `<div class="empty-state"><p>Không tìm thấy kết quả</p></div>`;
      return;
    }
    box.innerHTML = `<div class="space-y-2">${state.results.map((a) => appCard(a)).join("")}</div>`;
  }

  $("#search-results").addEventListener("click", (e) => {
    const btn = e.target.closest(".dl-btn");
    if (!btn) return;
    const id = Number(btn.closest("[data-id]").dataset.id);
    const app = state.results.find((a) => a.id === id);
    if (app) startDownload(app);
  });

  // ---------------- downloads ----------------
  async function startDownload(app) {
    try {
      await API().Download(app);
      state.downloads.set(app.id, { id: app.id, name: app.name, status: "queued", percent: 0, icon: app.icon });
      renderDownloads();
      renderResults();
      toast(`Bắt đầu tải ${app.name}`);
    } catch (ex) {
      toast(String(ex), "error");
    }
  }

  const statusLabel = {
    queued: "Trong hàng đợi",
    purchasing: "Đang lấy license…",
    downloading: "Đang tải",
    finalizing: "Đang ghi license…",
    done: "Hoàn tất",
    error: "Lỗi",
    cancelled: "Đã hủy",
  };

  function renderDownloads() {
    const list = Array.from(state.downloads.values()).reverse();
    const active = list.filter((d) => !["done", "error", "cancelled"].includes(d.status)).length;
    const badge = $("#badge-downloads");
    badge.textContent = active;
    badge.classList.toggle("hidden", active === 0);

    const box = $("#downloads-list");
    if (!list.length) {
      box.innerHTML = `<div class="empty-state"><p>Chưa có gì đang tải</p></div>`;
      return;
    }
    box.innerHTML = list
      .map((d) => {
        const isActive = !["done", "error", "cancelled"].includes(d.status);
        const color = d.status === "done" ? "bg-emerald-500" : d.status === "error" ? "bg-rose-500" : d.status === "cancelled" ? "bg-slate-500" : "bg-sky-500";
        const pct = Math.max(0, Math.min(100, d.percent || 0));
        const indeterminate = isActive && d.status !== "downloading";
        return `
        <div class="rounded-xl border border-white/5 bg-ink-900 p-3" data-id="${d.id}">
          <div class="flex items-center gap-3">
            ${iconHTML(d.icon, "h-10 w-10")}
            <div class="min-w-0 flex-1">
              <div class="flex items-center justify-between gap-2">
                <div class="truncate text-sm font-medium text-white">${esc(d.name)}</div>
                <div class="shrink-0 text-xs ${d.status === "error" ? "text-rose-300" : d.status === "done" ? "text-emerald-300" : "text-slate-400"}">${statusLabel[d.status] || d.status}</div>
              </div>
              <div class="mt-1.5 h-1.5 overflow-hidden rounded-full bg-white/5">
                <div class="h-full ${color} ${indeterminate ? "animate-pulse w-1/3" : ""} transition-all" style="${indeterminate ? "" : `width:${pct}%`}"></div>
              </div>
              <div class="mt-1 flex items-center justify-between text-[11px] text-slate-500">
                <span>${
                  d.status === "downloading"
                    ? `${fmtBytes(d.downloaded)} / ${fmtBytes(d.total)} · ${pct.toFixed(0)}% · ${fmtSpeed(d.speed || 0)}`
                    : esc(d.message || (d.status === "done" ? fmtBytes(d.total) : ""))
                }</span>
                <span class="flex gap-2">
                  ${isActive ? `<button class="cancel-btn text-rose-300 hover:underline">Hủy</button>` : ""}
                  ${d.status === "done" && d.path ? `<button class="show-btn text-sky-300 hover:underline">Mở thư mục</button>` : ""}
                  ${!isActive ? `<button class="remove-btn hover:underline">Ẩn</button>` : ""}
                </span>
              </div>
            </div>
          </div>
        </div>`;
      })
      .join("");
  }

  $("#downloads-list").addEventListener("click", async (e) => {
    const row = e.target.closest("[data-id]");
    if (!row) return;
    const id = Number(row.dataset.id);
    const d = state.downloads.get(id);
    if (e.target.closest(".cancel-btn")) API().Cancel(id);
    else if (e.target.closest(".show-btn") && d?.path) API().ShowInFolder(d.path);
    else if (e.target.closest(".remove-btn")) { state.downloads.delete(id); renderDownloads(); }
  });

  $("#clear-done-btn").addEventListener("click", () => {
    for (const [id, d] of state.downloads) if (["done", "error", "cancelled"].includes(d.status)) state.downloads.delete(id);
    renderDownloads();
  });

  function onProgress(p) {
    const prev = state.downloads.get(p.id) || {};
    state.downloads.set(p.id, { ...prev, ...p, icon: prev.icon });
    renderDownloads();
    if (p.status === "done") {
      toast(`Đã tải xong ${p.name}`, "ok");
      renderResults();
    } else if (p.status === "error") {
      toast(`${p.name}: ${p.message}`, "error");
      renderResults();
    } else if (p.status === "cancelled") {
      renderResults();
    }
  }

  // ---------------- library ----------------
  async function loadLibrary() {
    try {
      state.library = (await API().ListLibrary()) || [];
      const s = await API().GetStatus();
      state.status = s;
      $("#library-dir").textContent = s.outputDir;
      renderLibrary();
    } catch (ex) {
      toast(String(ex), "error");
    }
  }

  function renderLibrary() {
    const q = $("#library-filter").value.trim().toLowerCase();
    const items = state.library.filter((i) => !q || (i.name + " " + i.bundleId + " " + i.fileName).toLowerCase().includes(q));
    const badge = $("#badge-library");
    badge.textContent = state.library.length;
    badge.classList.toggle("hidden", state.library.length === 0);

    const box = $("#library-list");
    if (!items.length) {
      box.innerHTML = `<div class="empty-state"><p>${state.library.length ? "Không khớp bộ lọc" : "Chưa có file IPA nào"}</p></div>`;
      return;
    }
    const total = items.reduce((s, i) => s + i.sizeBytes, 0);
    box.innerHTML = `
      <div class="mb-2 text-xs text-slate-500">${items.length} file · ${fmtBytes(total)}</div>
      <div class="space-y-2">
      ${items
        .map(
          (i) => `
        <div class="card" data-file="${esc(i.fileName)}" data-path="${esc(i.path)}">
          ${iconHTML(i.icon, "h-11 w-11")}
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm font-medium text-white">${esc(i.name || i.bundleId)}</div>
            <div class="mt-0.5 truncate text-xs text-slate-500">${esc(i.fileName)}</div>
            <div class="mt-1 flex gap-1.5">
              <span class="pill">v${esc(i.version || "?")}</span>
              <span class="pill">${fmtBytes(i.sizeBytes)}</span>
              <span class="pill">${fmtDate(i.modTime)}</span>
            </div>
          </div>
          <div class="flex shrink-0 gap-1">
            <button class="btn-ghost show-btn text-xs">Mở thư mục</button>
            <button class="btn-danger del-btn">Xóa</button>
          </div>
        </div>`
        )
        .join("")}
      </div>`;
  }

  $("#library-filter").addEventListener("input", renderLibrary);
  $("#refresh-lib-btn").addEventListener("click", loadLibrary);
  $("#open-dir-btn").addEventListener("click", () => API().OpenOutputDir());
  $("#change-dir-btn").addEventListener("click", async () => {
    try {
      const dir = await API().ChooseOutputDir();
      $("#library-dir").textContent = dir;
      loadLibrary();
    } catch (ex) { toast(String(ex), "error"); }
  });
  $("#library-list").addEventListener("click", async (e) => {
    const row = e.target.closest("[data-file]");
    if (!row) return;
    if (e.target.closest(".show-btn")) API().ShowInFolder(row.dataset.path);
    else if (e.target.closest(".del-btn")) {
      if (!confirm(`Xóa file ${row.dataset.file}?`)) return;
      try { await API().DeleteFile(row.dataset.file); toast("Đã xóa", "ok"); loadLibrary(); } catch (ex) { toast(String(ex), "error"); }
    }
  });

  // ---------------- batch ----------------
  const BATCH_KEY = "ipa.batch.list";
  $("#batch-input").value = localStorage.getItem(BATCH_KEY) || "";
  $("#batch-input").addEventListener("input", (e) => localStorage.setItem(BATCH_KEY, e.target.value));

  $("#batch-check-btn").addEventListener("click", async () => {
    const lines = $("#batch-input").value.split(/\r?\n/).map((l) => l.trim()).filter((l) => l && !l.startsWith("#"));
    if (!lines.length) return toast("Danh sách trống");
    const box = $("#batch-results");
    const btn = $("#batch-check-btn");
    btn.disabled = true;
    $("#batch-download-btn").disabled = true;
    state.batch = [];
    box.innerHTML = "";
    for (const q of lines) {
      const row = document.createElement("div");
      row.className = "card py-2";
      row.innerHTML = `<div class="spinner h-4 w-4"></div><div class="truncate text-xs text-slate-400">${esc(q)}</div>`;
      box.appendChild(row);
      try {
        const app = await API().Lookup(q);
        state.batch.push(app);
        row.innerHTML = `${iconHTML(app.icon, "h-9 w-9")}<div class="min-w-0 flex-1"><div class="truncate text-sm text-white">${esc(app.name)}</div><div class="truncate text-[11px] text-slate-500">${esc(app.bundleId)} · v${esc(app.version)}${app.sizeBytes ? " · " + fmtBytes(app.sizeBytes) : ""}</div></div><span class="pill text-emerald-300">OK</span>`;
      } catch (ex) {
        row.innerHTML = `<div class="min-w-0 flex-1"><div class="truncate text-sm text-slate-300">${esc(q)}</div><div class="truncate text-[11px] text-rose-300">${esc(ex)}</div></div><span class="pill text-rose-300">Lỗi</span>`;
      }
    }
    btn.disabled = false;
    $("#batch-download-btn").disabled = state.batch.length === 0;
    $("#batch-download-btn").textContent = `Tải tất cả (${state.batch.length})`;
  });

  $("#batch-download-btn").addEventListener("click", async () => {
    for (const app of state.batch) await startDownload(app);
    switchTab("downloads");
  });

  // ---------------- init ----------------
  async function init() {
    if (!API()) {
      document.body.innerHTML = `<div class="empty-state h-screen"><p>Không kết nối được với backend Wails.</p></div>`;
      return;
    }
    RT().EventsOn("download:progress", onProgress);
    RT().EventsOn("library:changed", () => { if (!$("#tab-library").classList.contains("hidden")) loadLibrary(); else API().ListLibrary().then((l) => { state.library = l || []; renderLibrary(); }); });
    try {
      state.status = await API().GetStatus();
    } catch (ex) {
      toast(String(ex), "error");
    }
    if (state.status && state.status.loggedIn) showMain();
    else showLogin();
  }

  // Wails injects window.go / window.runtime before DOMContentLoaded in production;
  // poll briefly in case the bridge is still attaching.
  let tries = 0;
  const wait = () => (API() && RT()) || tries++ > 40 ? init() : setTimeout(wait, 50);
  document.addEventListener("DOMContentLoaded", wait);
})();

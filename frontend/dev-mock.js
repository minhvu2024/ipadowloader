/* Mock backend để xem giao diện trong trình duyệt (không có Wails).
   Dùng: mở dist/index.html kèm ?mock hoặc thêm <script src="../dev-mock.js"> */
(() => {
  const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
  const listeners = {};
  const emit = (n, d) => (listeners[n] || []).forEach((f) => f(d));
  const sample = [
    { id: 959841443, bundleId: "com.beeasy.shopee.vn", name: "Shopee: Mua Sắm Online", version: "3.52.11", price: 0, icon: "", developer: "Shopee", sizeBytes: 420e6, rating: 4.7, genre: "Shopping", minOs: "14.0" },
    { id: 579523206, bundleId: "com.zing.zalo", name: "Zalo", version: "25.03.01", price: 0, icon: "", developer: "VNG Corporation", sizeBytes: 210e6, rating: 4.3, genre: "Social Networking", minOs: "13.0" },
    { id: 1019220677, bundleId: "com.vng.zalopay", name: "ZaloPay", version: "9.4.2", price: 0, icon: "", developer: "ZION", sizeBytes: 150e6, rating: 4.5, genre: "Finance", minOs: "13.0" },
    { id: 310633997, bundleId: "net.whatsapp.WhatsApp", name: "Procreate Pocket", version: "5.0.3", price: 6.99, icon: "", developer: "Savage Interactive", sizeBytes: 88e6, rating: 4.8, genre: "Entertainment", minOs: "16.0" },
  ];
  let loggedIn = true;
  const lib = [
    { fileName: "com.zing.zalo_579523206_25.03.01.ipa", path: "C:\\Users\\Minh\\IPA\\com.zing.zalo_579523206_25.03.01.ipa", name: "Zalo", bundleId: "com.zing.zalo", appId: 579523206, version: "25.03.01", sizeBytes: 210e6, icon: "", modTime: new Date().toISOString() },
  ];
  window.go = { main: { App: {
    GetStatus: async () => ({ loggedIn, email: "minh@icloud.com", name: "Minh Vu", country: "US", outputDir: "C:\\Users\\Minh\\IPA", version: "1.0.0" }),
    Login: async (e, p, c) => { await sleep(600); if (!c) return { needCode: true, email: e }; loggedIn = true; return { email: e, name: "Minh Vu", country: "US" }; },
    Logout: async () => { loggedIn = false; },
    Search: async (t) => { await sleep(500); return sample.filter((a) => a.name.toLowerCase().includes(t.toLowerCase()) || true); },
    Lookup: async (q) => { await sleep(300); const a = sample.find((x) => x.bundleId === q || String(x.id) === q); if (!a) throw "không tìm thấy app trên store của tài khoản này"; return a; },
    Download: async (app) => {
      (async () => {
        emit("download:progress", { id: app.id, name: app.name, status: "purchasing", message: "Đang lấy license…" });
        await sleep(700);
        const total = app.sizeBytes;
        for (let i = 0; i <= 20; i++) { await sleep(150); emit("download:progress", { id: app.id, name: app.name, status: "downloading", percent: i * 5, downloaded: total * i / 20, total, speed: 8200 }); }
        emit("download:progress", { id: app.id, name: app.name, status: "finalizing", percent: 100, message: "Đang ghi license…" });
        await sleep(400);
        emit("download:progress", { id: app.id, name: app.name, status: "done", percent: 100, total, downloaded: total, path: "C:\\Users\\Minh\\IPA\\x.ipa", message: "Hoàn tất" });
        emit("library:changed");
      })();
    },
    Cancel: async () => {},
    ListLibrary: async () => lib,
    DeleteFile: async () => {},
    ChooseOutputDir: async () => "C:\\Users\\Minh\\IPA",
    OpenOutputDir: async () => {},
    ShowInFolder: async () => {},
  } } };
  window.runtime = { EventsOn: (n, f) => { (listeners[n] = listeners[n] || []).push(f); } };
})();

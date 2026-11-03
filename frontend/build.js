// Copy static files into dist (Tailwind CSS is built separately).
const fs = require("fs");
const path = require("path");
fs.mkdirSync("dist", { recursive: true });
fs.copyFileSync("index.html", path.join("dist", "index.html"));
fs.copyFileSync(path.join("src", "app.js"), path.join("dist", "app.js"));
console.log("copied index.html, app.js -> dist/");

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.js"],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        ink: {
          950: "#0f1117",
          900: "#161923",
          800: "#1e2230",
          700: "#2a2f40",
        },
      },
      fontFamily: {
        sans: ["Segoe UI Variable", "Segoe UI", "Inter", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
};

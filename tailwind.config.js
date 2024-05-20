/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templ/**/*.templ"],
  plugins: [require("daisyui")],
  theme: {
    extend: {
      spacing: {
        128: "32rem",
      },
    },
  },
  daisyui: {
    themes: ["emerald"],
  },
};

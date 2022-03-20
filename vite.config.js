import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig({
  plugins: [vue()],
  root: "./web",
  build: {
    target: "es2015",
    outDir: "../build/public"
  },
});

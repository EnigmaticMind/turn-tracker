import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import packageInfo from './package.json';

export default defineConfig({
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],
  base: process.env.NODE_ENV === 'development' ? '/' :  `/${packageInfo.name}/`,
});

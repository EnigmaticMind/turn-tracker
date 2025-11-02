import { reactRouter } from "@react-router/dev/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";
import packageInfo from './package.json';

export default defineConfig({
  plugins: [tailwindcss(), reactRouter(), tsconfigPaths()],
  base: process.env.NODE_ENV === 'development' ? '/' : `/${packageInfo.name}/`,
  
  build: {
    minify: 'esbuild', // or 'terser' for more aggressive removal
    terserOptions: {
      compress: {
        drop_console: true, // Remove all console.* calls
        drop_debugger: true,
      },
    },
  },
});

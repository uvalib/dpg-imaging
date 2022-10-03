 /*global process */

import { fileURLToPath, URL } from 'url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
   plugins: [vue()],
   resolve: {
      alias: {
         '@': fileURLToPath(new URL('./src', import.meta.url))
      }
   },
   server: { // this is used in dev mode only
      port: 8080,
      proxy: {
         '/api': {
            target: process.env.DPG_SRV,  //export DPG_SRV=http://localhost:8085
            changeOrigin: true
         },
         '/authenticate': {
            target: process.env.DPG_SRV,
            changeOrigin: true
         },
         '/config': {
            target: process.env.DPG_SRV,
            changeOrigin: true
         },
         '/healthcheck': {
            target: process.env.DPG_SRV,
            changeOrigin: true
         },
         '/version': {
            target: process.env.DPG_SRV,
            changeOrigin: true
         },
      }
   },
   css: {
     preprocessorOptions: {
       scss: {
          additionalData: `
            @import "@/assets/stylesheets/_mixins.scss";
         `
      },
     },
   },
})


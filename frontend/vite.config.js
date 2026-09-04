 /*global process */

import { fileURLToPath, URL } from 'url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import ui from '@nuxt/ui/vite'

// https://vitejs.dev/config/
export default defineConfig({
   plugins: [
      vue(),
      ui({
         ui: {
            button: {
               default: {
                  class: "cursor-pointer"
               },
               compoundVariants: [
                  {
                     color: "primary",
                     variant: "solid",
                     class: "text-white focus:outline-offset-2 focus:outline-1 focus:outline-dashed",
                  },
                  {
                     color: "secondary",
                     variant: "solid",
                     class: "border border-brand-grey-100 hover:bg-gray-200 focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100",
                  },
                  {
                     color: "error",
                     variant: "solid",
                     class: "text-white focus:outline-offset-2 focus:outline-1 focus:outline-dashed",
                  },
                  {
                     color: "neutral",
                     variant: "ghost",
                     class: "text-white hover:bg-gray-200 focus:outline-offset-2 focus:outline-1 focus:outline-dashed",
                  },
               ],
            },
            dropdownMenu: {
               slots: {
                  label: "bg-brand-grey-100 rounded-t-sm font-bold",
                  viewport: "bg-white text-black",
                  item: "hover:bg-brand-blue-alt-400 rounded-sm"
               }
            },
            header: {
               slots: {
                  root: "!bg-brand-blue h-auto", // h-auto needed to make header height include bottom slot
                  container: "!px-5 !py-5 !max-w-full",
                  right: "text-white",
                  body: "bg-brand-blue",
                  header:  "!bg-brand-blue",
                  content: "bg-brand-blue",
               }
            },
            modal: {
               rounded: 'rounded-sm',
               slots: {
                  header: "bg-brand-teal-200 flex items-center gap-0 p-2.5 sm:px-2.5 min-h-0",
                  title: 'font-semibold text-black',
                  close: 'absolute top-1.5 end-1.5 rounded-full text-black hover:bg-brand-teal-100',
                  content: "bg-white",
                  footer: "justify-end sm:px-4 p-4",
                  body: "border-0"
               }
            },
            navigationMenu: {
               defaultVariants: {
                  color: 'neutral',
               },
               slots: {
                  link: "before:rounded-none gap-1 focus-visible:before:outline-dashed focus-visible:before:outline-brand-blue-alt-200 focus-visible:before:outline-1",
                  linkLabel: "text-white hover:bg-brand-blue-alt px-2 py-1 rounded-lg",
                  linkLeadingIcon: "!text-white",
                  childLinkIcon: "!text-white",
                  childLinkLabel: "hover:bg-brand-blue-alt  px-2 py-1 rounded-lg text-white",
                  linkTrailingIcon: "!text-white",
                  viewport: "ring-black bg-brand-blue"
               },
            },
            radioGroup: {
               slots: {
                  fieldset: 'flex gap-6',
                  legend: 'mb-1 block font-medium text-default',
                  item: 'flex items-start gap-2',
                  container: 'flex items-center',
                  base: 'rounded-full ring ring-inset ring-accented overflow-hidden focus-visible:outline-none',
                  indicator:  'flex items-center justify-center size-full after:bg-default after:rounded-full ',
                  label: 'text-black',
               },
               variants: {
                  color: {
                     info: {
                        base: 'focus-visible:outline-none hover:bg-brand-blue-alt-300 focus:outline-offset-2 focus:outline-dotted focus:outline-brand-blue-alt-100',   
                     }
                  }
               }
            },
            input: {
               slots: {
                  root: '!bg-white !text-black',
                  base: '!bg-white !text-black  !border-none !ring-brand-grey-100 focus:outline-offset-2 focus:outline-dashed focus:outline-brand-blue-alt-100'
               },
            },
            select: {
               slots: {
                  base: "focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100",
                  viewport: "bg-white text-black",
                  item: [ // the items style is an ARRY and teh second elemsnt defaults to transition animattion. just override stuff at idx 0
                     'data-highlighted:not-data-disabled:before:bg-brand-blue-alt-300',
                  ],
               },
               compoundVariants: [
                  {
                     color: 'primary',
                     variant: 'outline',
                     class: 'bg-white text-black ring-brand-grey-100 hover:bg-white hover:outline-2', 
                  },
               ]
            },
            selectMenu: {
               slots: {
                  base: "!bg-white !text-black !ring-brand-grey-100  focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100",
                  content: "ring-brand-grey-100",
                  viewport: "bg-white text-black",
                  item: [
                     'data-highlighted:not-data-disabled:before:bg-brand-blue-alt-300',
                  ],
                  focusScope: 'bg-white',
                  empty: 'text-black'
               },
            },
            tabs: {
               variants: {
                  variant: {
                     link: {
                        trigger: "rounded-none hover:data-[state=inactive]:bg-brand-grey-200 hover:data-[state=inactive]:font-bold data-[state=inactive]:cursor-pointer"
                     }
                  }
               }
            },
            toast: {
               slots: {
                  progress: 'bottom-1',
                  close: 'rounded-full'
               },
               variants: {
                  color: {
                     error: {
                        root: 'bg-brand-red-100',
                        icon: 'text-black',
                        title: 'text-black font-semibold font-size-1.5',
                        description: 'text-black',   
                        close: 'text-black hover:text-black hover:bg-brand-red'
                     }
                  }
               }
            }
         }
      })
   ],
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
      preprocessorOptions : {
          scss: {
              api: "modern-compiler",
          },
      }
   },
})


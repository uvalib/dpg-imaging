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
            accordion: {
               slots: {
                  trigger: "rounded-none bg-brand-grey-200 p-2 border-1 border-brand-grey-100 flex justify-between text-lg font-semibold",
                  body: "rounded-none border-1 border-brand-grey-100 border-t-0 p-4"
               }
            },
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
                     class: "border bg-brand-grey-200 text-black border-brand-grey-100 hover:bg-gray-200 focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100",
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
                  {
                     color: "neutral",
                     variant: "outline",
                     class: "hover:bg-brand-blue-alt-300 focus:outline-offset-2 focus:outline-1 focus:outline-dashed disabled:bg-brand-grey-200 disabled:text-brand-grey-100",
                  },
               ],
            },
            dropdownMenu: {
               slots: {
                  item: "hover:bg-brand-blue-alt-300 rounded-md"
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
                  content: "outline-brand-grey outline-1",
                  close: 'absolute top-1.5 end-1.5 rounded-full text-black hover:bg-brand-teal-100',
                  footer: "justify-end sm:px-4 p-4 pt-0",
                  body: "border-0 !p-4",
                  overlay: "!bg-brand-grey/70" // the /70 sets opacity
               }
            },
            navigationMenu: {
               slots: {
                  link: "rounded-lg focus-visible:before:outline-dashed focus-visible:before:outline-brand-blue-alt-200 focus-visible:before:outline-1 hover:bg-brand-blue-alt",
                  linkLabel: "text-white",
                  linkLeadingIcon: "!text-white",
                  linkTrailingIcon: "!text-white",
                  childLink: "rounded-lg focus-visible:before:outline-dashed focus-visible:before:outline-brand-blue-alt focus-visible:before:outline-1 hover:bg-brand-blue-alt-300",
               },
               variants: {
                  active: {
                     true: {
                        childLink: 'before:bg-white hover:bg-brand-blue-alt-300 rounded-lg',
                     }
                  }
               },
               compoundVariants: [
                  {
                     disabled: false,
                     variant: 'pill',
                     highlight: true,
                     orientation: 'horizontal',
                     class: {
                        link: 'data-[state=open]:before:bg-brand-blue-alt-A'
                     }
                  },
                  {
                     highlightColor: 'primary',
                     highlight: true,
                     level: true,
                     active: true,
                     class: {
                        link: 'after:bg-brand-blue'
                     }
                  },
               ]
            },
            radioGroup: {
               slots: {
                  fieldset: 'gap-6.5',
                  item: 'gap-2',
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
                  base: '!border-none !ring-brand-grey-100 focus:outline-offset-2 focus:outline-dashed focus:outline-brand-blue-alt-100'
               },
            },
            inputNumber: {
               slots: {
                  base: '!border-none !ring-brand-grey-100 focus:outline-offset-2 focus:outline-dashed focus:outline-brand-blue-alt-100'
               }
            },
            listbox: {
               slots: {
                  item: [ // the items style is an ARRY and teh second elemsnt defaults to transition animattion. just override stuff at idx 0
                     'data-highlighted:not-data-disabled:before:bg-brand-blue-alt-300',
                  ],
               }
            },
            select: {
               slots: {
                  base: "focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100",
                  item: [ // the items style is an ARRY and teh second elemsnt defaults to transition animattion. just override stuff at idx 0
                     'data-highlighted:not-data-disabled:before:bg-brand-blue-alt-300',
                  ],
               },
               compoundVariants: [
                  {
                     color: 'primary',
                     variant: 'outline',
                     class: 'ring-brand-grey-100 hover:bg-white hover:outline-2', 
                  },
               ]
            },
            selectMenu: {
               slots: {
                  base: "!ring-brand-grey-100  focus:outline-offset-2 focus:outline-dashed focus:outline-brand-grey-100", 
                  item: [
                     'data-highlighted:not-data-disabled:before:bg-brand-blue-alt-300',
                  ],
                 
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


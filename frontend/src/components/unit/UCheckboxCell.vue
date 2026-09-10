<template>
    <!--  
      NOTES:
      Nuxt UI components in a plain Vue app are injected at build time by unplugin-vue-components, 
      which only rewrites usage in SFC templates/script-setup code. 
      Nothing registers UCheckbox in Vue's global component registry, so at the moment your cell render executes, 
      resolveComponent('UCheckbox') (or a string-based h('UCheckbox', ...)) has nothing to find.
      You can't "fix" resolution at render time — you have to hand the real component object into the render function.
    -->
    <UCheckbox />
</template>
<template>
   <div id="app">
      <div class="header" role="banner" id="uva-header">
         <div class="main-header">
            <div class="library-link">
               <a target="_blank" href="https://library.virginia.edu">
                  <UvaLibraryLogo />
               </a>
            </div>
            <div class="site-link">
               <router-link to="/">DPG Imaging</router-link>
               <p class="version">v{{ version }}</p>
            </div>
         </div>
         <div class="user-banner" v-if="jwt">
            <label>Signed in as:</label><span class="user">{{ signedInUser }}</span>
            <span class="signout" @click="signout">Sign out</span>
         </div>
      </div>
      <router-view />
      <ErrorMessage v-if="hasError" />
      <ScrollToTop />
   </div>
</template>

<script>
import UvaLibraryLogo from "@/components/UvaLibraryLogo";
import ScrollToTop from "@/components/ScrollToTop";
import { mapState, mapGetters } from "vuex";
export default {
   components: {
      UvaLibraryLogo,
      ScrollToTop,
   },
   computed: {
      ...mapState({
         hasError: (state) => state.error,
         version: (state) => state.version,
         jwt: (state) => state.user.jwt,
      }),
      ...mapGetters(["signedInUser"]),
   },
   methods: {
      signout() {
         this.$store.commit("signout")
         this.$router.push("signedout")
      }
   },
   async beforeCreate() {
      this.$store.dispatch("getVersion")
      await this.$store.dispatch("getConfig")
   },
};
</script>

<style lang="scss">
:root {
   --uvalib-brand-blue-lightest: #87b9d9;
   --uvalib-brand-blue-lighter: #3395d4;
   --uvalib-brand-blue-light: #0370b7;
   --uvalib-brand-blue: #232d4b;
   --uvalib-brand-orange-lightest: #ffead6;
   --uvalib-brand-orange: #e57200;
   --uvalib-brand-orange-dark: #b35900;
   --uvalib-blue-alt-light: #bfe7f7;
   --uvalib-blue-alt: #007bac;
   --uvalib-blue-alt-dark: #005679;
   --uvalib-blue-alt-darkest: #141e3c;
   --uvalib-teal-lightest: #c8f2f4;
   --uvalib-teal-light: #5bd7de;
   --uvalib-teal: #25cad3;
   --uvalib-teal-dark: #1da1a8;
   --uvalib-teal-darker: #16777c;
   --uvalib-green-lightest: #89cc74;
   --uvalib-green: #62bb46;
   --uvalib-green-dark: #4e9737;
   --uvalib-red-lightest: #fbcfda;
   --uvalib-red: #ef3f6b;
   --uvalib-red-emergency: #df1e43;
   --uvalib-red-darker: #b30000;
   --uvalib-red-dark: #df1e43;
   --uvalib-yellow-light: #fef6c8;
   --uvalib-yellow: #ecc602;
   --uvalib-yellow-dark: #b99c02;
   --uvalib-beige: #f7efe1;
   --uvalib-beige-dark: #c0b298;
   --uvalib-grey-lightest: #f1f1f1;
   --uvalib-grey-light: #dadada;
   --uvalib-grey: #808080;
   --uvalib-grey-dark: #565656;
   --uvalib-grey-darkest: #2b2b2b;
   --uvalib-text-light: #ffffff;
   --uvalib-text: var(--uvalib-grey-dark);
   --uvalib-text-dark: var(--uvalib-grey-darkest);

   --box-shadow: 0 3px 6px rgba(0, 0, 0, 0.16), 0 3px 6px rgba(0, 0, 0, 0.23);
}

.right-pad {
   margin-right: 10px;
}

#app {
   font-family: "franklin-gothic-urw", arial, sans-serif;
   -webkit-font-smoothing: antialiased;
   -moz-osx-font-smoothing: grayscale;
   text-align: center;
   color: var(--color-primary-text);
   margin: 0;
   padding: 0;
   background: white;

   a {
      color: var(--uvalib-blue-alt-dark);
      font-weight: 500;
      text-decoration: none;
      &:hover {
         text-decoration: underline;
      }
   }
   p.version {
      margin: 5px 0 0 0;
      font-size: 0.5em;
      text-align: right;
   }
   div.library-link {
      width: 220px;
      order: 0;
      flex: 0 1 auto;
      align-self: flex-start;
   }
   div.site-link {
      order: 0;
      font-size: 1.5em;
      a {
         color: white;
         text-decoration: none;
         &:hover {
            text-decoration: underline;
         }
      }
   }
}
input,
select {
   box-sizing: border-box;
   border-radius: 4px;
   padding: 3px 5px;
   border: 1px solid var(--uvalib-grey);
   width: 100%;
}
body {
   margin: 0;
   padding: 0;
}
div.header {
   background-color: var(--uvalib-brand-blue);
   color: white;
   padding: 1vw 20px;
   text-align: left;
   position: relative;
   box-sizing: border-box;
   .main-header {
      display: flex;
      flex-direction: row;
      flex-wrap: nowrap;
      justify-content: space-between;
      align-content: stretch;
      align-items: center;
   }
   .user-banner {
      text-align: right;
      padding: 10px 0 0 0;
      font-size: 0.8em;
      margin: 0;
      label {
         font-weight: bold;
         margin-right: 5px;
      }
      .user {
         font-weight: 100;
      }
      .signout {
         display: inline-block;
         margin-left: 10px;
         cursor: pointer;
         border: 1px solid var(--uvalib-brand-blue-light);
         padding: 2px 9px;
         border-radius: 3px;
         background: var(--uvalib-brand-blue-light);
         &:hover {
            background: var(--uvalib-brand-blue-lighter);
         }
      }
   }
}

</style>

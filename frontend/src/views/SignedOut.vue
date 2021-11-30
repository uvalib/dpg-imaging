<template>
   <div class="signedout">
      <template v-if="$route.query.expired">
         <h2>Session Expired</h2>
         <div class="message">
            <p>Your session has expired and you have been signed out.</p>
         </div>
      </template>
      <template v-else>
         <h2>Signed Out</h2>
         <div class="message">
            <p>You have successfully signed out of your account.</p>
         </div>
      </template>
      <div class="buttons"><DPGButton @click="signinClick">Sign in with NetBadge</DPGButton></div>
   </div>
</template>

<script>
export default {
   name: "signedout",
   components: {
   },
   methods: {
      signinClick() {
         this.$router.push("/")
      }
   }
};
</script>

<style scoped lang="scss">
.signedout {
   padding: 25px;
   h2 {
      color: var(--uvalib-brand-orange);
   }
   .message {
      margin-top: 40px;
      font-size: 1.15em;
   }
   .buttons {
      margin-top: 35px;
      font-size: 1.25em;
   }
}
</style>

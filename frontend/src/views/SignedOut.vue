<template>
   <div class="notfound">
      <h2>Signed Out</h2>
      <div class="message">
         <p>You have successfully signed out of your account.</p>
      </div>
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
.notfound {
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

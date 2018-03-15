<template lang="pug">
  div(class="login-view layout-padding")
    q-card.bg-white.card(inline)
      q-card-title
        span(my-slot="subtitle")
          h3.title.text-indigo.color-5 Login
      q-card-main
        form(@submit.prevent="authenticate")
          q-field.username(
            icon="fa-hashtag"
            label=""
            helper=""
            error-label="We need a valid email"
            )
            q-input(v-model="form.username" stack-label="Username")
          q-field.password(
            icon="fa-key"
            label=""
            helper=""
            error-label="Write a password"
          )
            q-input(v-model="form.password" stack-label="Password")
          .center
            q-btn(type="submit" big class="bg-primary text-white") Login
</template>
<script>
import { QInput, QField, QBtn, QCard, QCardTitle, QCardMain, Notify } from 'quasar'

export default {
  data () {
    return {
      form: {
        username: null,
        password: null
      }
    }
  },
  mounted () {
    console.log('Login view Loaded!')
  },
  methods: {
    loginError () {
      Notify.create({
        message: 'Email or password incorrect',
        icon: 'lock',
        timeout: 2500,
        color: 'negative',
        textcolor: '#fff'
      })
    },
    async authenticate () {
      let username = this.form.username
      let password = this.form.password
      try {
        let authentication = await this.$basicauth.login(username, password)
        let redirection = '/' // Default route
        if (this.$route.query.redirect && authentication) {
          // If query has a prop redirect
          redirection = this.$route.query.redirect
        }
        // Otherwise redirect to default route
        this.$router.replace(redirection)
      } catch (error) {
        // Error in Login
        console.log(error)
        this.loginError()
      }
    }

  },
  components: {
    QField,
    QInput,
    QBtn,
    QCard,
    QCardTitle,
    QCardMain
  }
}
</script>
<style lang="scss">
  .login-view {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background-color: #898989;
    .username , .password{
      margin-bottom: 2rem;
    }
    .card {
      padding: 10px;
      min-width: 400px;
      min-height: 320px;
      .title{
        margin:0;
        padding-left: 1rem;
        border-left: 3px solid rgb(37, 70, 177)
      }
    }
    form {
      max-width: 420px;
    }
  }
</style>

<!--form content-->
<!--1. Name (as in passsport)-->
<!--2. Date of birth-->
<!--3. country of residence-->
<!--4. email-->
<!--5. upload identity document (passport or driving license)-->

<template lang="pug">
  div(class="request-view layout-padding")
    q-card.bg-white.card(inline)
      q-card-title
        span(my-slot="subtitle")
          h3.title.text-indigo.color-5 Register user
      q-card-main
        form.new-user-form(@submit.prevent="createUser")
          q-field(
          icon="person"
          label=""
          helper=""
          error-label="We need a valid name")
            q-input(v-model="form.first_name" stack-label="Name (as in passsport)")

          q-field.email(
          icon="email"
          label=""
          helper=""
          error-label="We need a valid name")
            q-input(v-model="form.email" type="email" stack-label="Email")

          q-field(
            icon="public"
            label=""
            helper=""
            error-label="We need a valid name")
            q-select(
              stack-label="Country of residence"
              filter
              separator
              v-model="select"
              :options="options")

          q-field(
            icon="cake"
            label=""
            helper=""
            error-label="We need a valid name")
            q-datetime(
              :error=false
              stack-label="Birthday"
              type="date"
              v-model="form.birthday"
              color="brown"
              :min="maxAge"
              :max="minAge"
              default-view="year")

          q-field(
            icon="fa-id-card"
            label=""
            helper=""
            error-label="We need a passport image")
            q-uploader(
              color="secondary"
              auto-expand
              stack-label="Passport"
              hide-upload-button
              :url="url")

          .float-right
            q-btn(type="submit" big class="bg-primary text-white") Add
</template>

<script>
// import { mapActions, mapGetters } from 'vuex'
// import { Fire, Listen } from 'helpers'
// import { User } from 'src/app/database/UserModel'
import { QInput, QField, QBtn, QCard, QCardTitle, QCardMain, QSelect, QDatetime, QUploader, Notify, date } from 'quasar'

const today = new Date()
const { subtractFromDate } = date

export default {
  name: 'AddCredentials',
  data () {
    return {
      form: {
        first_name: null,
        last_name: null,
        email: null,
        birthday: null,
        synced: '0'
      },
      minAge: subtractFromDate(today, { year: 18 }),
      maxAge: subtractFromDate(today, { year: 80 }),
      select: '',
      options: [
        {
          label: 'Google',
          value: 'goog'
        },
        {
          label: 'Facebook',
          value: 'fb'
        },
        {
          label: 'Twitter',
          value: 'twtr'
        },
        {
          label: 'Apple Inc.',
          value: 'appl'
        },
        {
          label: 'Oracle',
          value: 'ora'
        }
      ]
    }
  },
  methods: {
    requestSuccess () {
      Notify.create({
        position: 'top',
        message: 'User added',
        icon: 'check_circle',
        timeout: 2500,
        color: 'positive',
        textcolor: '#fff'
      })
    },
    resetForm () {
      this.form = {
        first_name: null,
        last_name: null,
        email: null,
        synced: '0',
        interval: undefined
      }
    }
    // async createUser () {
    //   if (this.validForm) {
    //     // if ('serviceWorker' in navigator && 'SyncManager' in window) {
    //     //   navigator.serviceWorker.ready
    //     //     .then(function (sw) {
    //     //       return sw.sync.register('sync-new-user')
    //     //     })
    //     // }
    //     //
    //     // await User.add(this.form)
    //     this.resetForm()
    //   }
    // },
    // fire () {
    //   const CLIENT_SECRET = env('CLIENT_SECRET', null)
    //   // Fire('app.custom-event', { CLIENT_SECRET })
    // },
    // ...mapActions('users', [ 'getCurrentUser', 'getUsers' ])
  },
  computed: {
    // validForm () {
    //   return this.form.first_name && this.form.last_name && this.form.email
    // },
    // ...mapGetters('users', ['currentUser', 'users'])
  },
  mounted () {
    this.requestSuccess()
    // Listen('app.custom-event', (payload) => {
    //   console.log('a custom event was dispatched', payload)
    // })
  },
  components: { QInput, QField, QBtn, QCard, QCardTitle, QCardMain, QSelect, QDatetime, QUploader }
}
</script>

<style lang="scss">
  .request-view {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100vh;
    background-color: #898989;
    .card {
      padding: 10px;
      min-width: 400px;
      min-height: 320px;
      .title {
        margin: 0;
        padding-left: 1rem;
      }
    }
    .new-user-form {
      padding: 24px;
      max-width: 520px;
      margin: 0 auto;
      .q-field {
        margin-bottom: 1rem;
      }
    }
  }
</style>

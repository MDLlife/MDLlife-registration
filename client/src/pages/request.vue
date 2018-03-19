<template lang="pug">
  div(class="request-view layout-padding")
    q-card.bg-white.card(inline)
      q-card-title
        span(my-slot="subtitle")
          h3.title.text-indigo.color-5 Register user
      q-card-main
        form.new-user-form(@submit.prevent="submitForm")
          q-field(
          :error="!!errors.name"
          :error-label="errors.name"
          icon="fa-user")
            q-input(v-model="form.name" stack-label="Full name as in passport" @input="removeError('name')")

          q-field.email(
          :error="!!errors.email"
          :error-label="errors.email"
          icon="fa-at")
            q-input(v-model="form.email" type="email" stack-label="Email" @input="removeError('email')")

          q-field(
            :error="!!errors.country"
            :error-label="errors.country"
            icon="public")
            q-input(
              v-model="form.country"
              placeholder=""
              stack-label="Country of residence"
              @input="removeError('country')")
              q-autocomplete(
                :static-data="{field: 'value', list: countries}")

          q-field(
            :error="!!errors.birthday"
            :error-label="errors.birthday"
            icon="fa-calendar-alt")
            q-datetime(
              :error=false
              stack-label="Birthday"
              type="date"
              v-model="form.birthday"
              color="brown"
              :min="maxAge"
              :max="minAge"
              default-view="year"
              @input="removeError('birthday')")

          q-field(
            :error="!!errors.passport"
            :error-label="errors.passport"
            icon="fa-id-card")
              q-input-file(
                color="secondary"
                auto-expand
                v-model="form.passport"
                stack-label="Upload identity document (passport or driving license)"
                @input="removeError('passport')")

          q-field(
            :error="!!errors.captchaSolution"
            :error-label="errors.captchaSolution"
            inset="icon")
            .row
              .col
                img.captcha-img(
                :src="form._captchaSrc"
                alt="captcha")
                q-icon(
                  name="fa-sync-alt"
                  @click.native="reloadCaptcha()")
                audio(
                  id = "captcha-audio" style="display:none" preload=none)
              .col-6.captcha-solution
                q-input(v-model="form.captchaSolution" @input="removeError('captchaSolution')")

          .float-right
            q-btn.bg-primary.text-white(
              :loading="loading"
              type="submit"
              big
              ) Add
              q-spinner-hourglass( slot="loading" )
</template>

<script>
import { QInput, QField, QBtn, QCard, QCardTitle, QCardMain, QAutocomplete, QDatetime, QSpinnerHourglass, Notify, date } from 'quasar'
import { QInputFile } from '@components/input-file'
import countries from 'assets/countries.json'
import Config from 'src/config'

const today = new Date()
const { subtractFromDate, formatDate } = date

function parseCountries () {
  return countries.map(country => {
    return {
      label: country,
      value: country
    }
  })
}

export default {
  name: 'AddCredentials',
  data () {
    return {
      form: {
        name: null,
        email: null,
        country: null,
        birthday: null,
        passport: [],
        captchaSolution: null,
        captchaId: null,
        _captchaSrc: null
      },
      minAge: subtractFromDate(today, { year: 18 }),
      maxAge: subtractFromDate(today, { year: 80 }),
      errors: {},
      countries: parseCountries(),
      loading: false
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
      this.errors = {}

      this.form.name = null
      this.form.email = null
      this.form.country = null
      this.form.birthday = null
      this.form.passport = []
    },
    removeError (name) {
      this.errors[name] = null
    },
    setCaptcha () {
      const form = this.form

      this.$axios.get(Config('api.captcha_id'))
        .then(function (response) {
          if (response.status === 200) {
            form.captchaSolution = null
            form.captchaId = response.data
            form._captchaSrc = Config('api.captcha') + form.captchaId + '.png'
          }
        })
        .catch(function (error) {
          console.log(error.response)
        })
    },
    reloadCaptcha () {
      let src = this.form._captchaSrc
      if (!src) return
      let p = src.indexOf('?')
      if (p >= 0) {
        src = src.substr(0, p)
      }

      this.form._captchaSrc = src + '?reload=' + (new Date()).getTime()
    },
    submitForm () {
      const self = this
      const formData = new FormData()

      this.loading = true

      formData.append('name', this.form.name || '')
      formData.append('email', this.form.email || '')
      formData.append('birthday', formatDate(this.form.birthday, 'YYYY-MM-DD'))
      formData.append('country', this.form.country || '')
      formData.append('captchaId', this.form.captchaId || '')
      formData.append('captchaSolution', this.form.captchaSolution || '')
      formData.append('passport', this.form.passport[0])

      this.$axios.post(Config('api.add_whitelist'), formData)
        .then(function (response) {
          console.log(response)
          debugger
          if (response.data && response.data.success) {
            self.resetForm()
            self.requestSuccess()
          }
        })
        .catch(function (error) {
          if (error && error.response) {
            error = error.response
          }

          if (error) {
            self.errors = {}

            switch (error.status) {
              case 422:
                self.errors = error.data.errors || {}
                break
              default:
                Notify.create({
                  position: 'top',
                  message: 'Server error code: ' + error.status,
                  icon: 'warning',
                  timeout: 5000,
                  color: 'negative',
                  textcolor: '#fff'
                })
                break
            }
          }
        })
        .finally(function () {
          self.setCaptcha()
          self.loading = false
        })
    }
  },
  computed: {},
  mounted () {
    this.setCaptcha()
  },
  components: { QInput, QField, QBtn, QCard, QCardTitle, QCardMain, QAutocomplete, QDatetime, QInputFile, QSpinnerHourglass }
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

      .captcha-img {
        border: 1px solid #eee;
        width: 100%;
        padding: 0.5rem 1rem;
      }

      .captcha-solution {
        padding: 2rem 0 0 1.5rem;
      }
    }
  }
</style>

<template lang="pug">
  q-layout(view="lHh Lpr lFf")
    q-layout-header
      q-toolbar
        q-btn(
          flat
          dense
          round
          @click="leftDrawerOpen = !leftDrawerOpen")
          q-icon(name="menu")

        q-toolbar-title MDLlife App
          div(slot="subtitle") Running on Quasar v{{ $q.version }}

        q-btn(
          flat
          dense
          round
          @click.native="route('/')")
          q-icon(name="person_add")
        q-btn(
          flat
          dense
          round
          @click.native="route('/check')")
          q-icon(name="spellcheck")

    q-layout-drawer(
      v-model="leftDrawerOpen"
      content-class="bg-grey-2")
      q-list(
        no-border
        link
        inset-delimiter)
        q-list-header Menu
        q-item(@click.native="route('/')")
          q-item-side(icon="person_add")
          q-item-main(label="Add whitelist" sublabel="Add personal information")
        q-item(@click.native="route('/check')")
          q-item-side(icon="spellcheck")
          q-item-main(label="Check" sublabel="Verify whitelists")

    q-page-container
      router-view
    q-page-container
</template>

<script>
import { openURL } from 'quasar'

export default {
  name: 'LayoutDefault',
  data () {
    return {
      leftDrawerOpen: this.$q.platform.is.desktop
    }
  },
  methods: {
    openURL,
    logout () {
      // this.$oauth.logout()
      this.$router.push('/login')
    },
    route (path) {
      this.$router.push(path)
    }
  }
}
</script>

<style>
</style>

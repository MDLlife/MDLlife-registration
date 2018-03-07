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
          @click.native="route('/request')")
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
        q-list-header Links
        q-item(@click.native="openURL('http://quasar-framework.org')")
          q-item-side(icon="school")
          q-item-main(label="Docs" sublabel="quasar-framework.org")

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

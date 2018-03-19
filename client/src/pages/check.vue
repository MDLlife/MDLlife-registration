<template lang="pug">
  q-page.docs-table( padding )
    q-table.responsive(
      ref="table"
      color="primary"
      :data="serverData"
      :columns="columns"
      :visible-columns="visibleColumns"
      :filter="filter"
      row-key="name"
      :pagination.sync="serverPagination"
      @request="request"
      :loading="loading")

      q-tr( slot="header" slot-scope="props")
        q-th( v-for="col in props.cols" :key="col.name" :props="props") {{ col.label }}

      template( slot="body" slot-scope="props" )
        q-tr.image-row( :props="props" )
          q-td( colspan="100%" )
            .row
              .col-12.text-left
                img(
                  v-if="imageRegex.test(props.row.Passport.Extension)"
                  :src="props.row.Passport.Src"
                  alt="")
                br
                q-btn(
                  outline
                  color="primary"
                  icon="fa-download"
                  @click.native="downloadBase64File(props.row.Passport.Src, 'passport-' + props.row.Passport.Id + '.' + props.row.Passport.Extension)")
              .col-12.col-xl-3.td-column(
                v-if="props.colsMap['name']") Full name: {{ props.row.Name }}
              .col-12.col-xl-3.td-column(
                v-if="props.colsMap['birthday']") Birthday: {{ props.row.Birthday }}
              .col-12.col-xl-3.td-column(
                v-if="props.colsMap['country']") Country: {{ props.row.Country }}
              .col-12.col-lx-3.text-right( v-if="props.colsMap['action']" )
                q-btn(
                  v-if="!isDisableAction(props.row.VerificationStage, 'declined')"
                  :loading="loadingActions"
                  round
                  color="negative"
                  icon="fa-times"
                  @click.native="actionRequest(props.row.Id, 'decline')")
                q-btn(
                  v-if="!isDisableAction(props.row.VerificationStage, 'question')"
                  :loading="loadingActions"
                  round
                  color="warning"
                  icon="fa-question"
                  @click.native="actionRequest(props.row.Id, 'question')")
                q-btn(
                  v-if="!isDisableAction(props.row.VerificationStage, 'accepted')"
                  :loading="loadingActions"
                  round
                  color="positive"
                  icon="fa-check"
                  @click.native="actionRequest(props.row.Id, 'accept')")

      template( slot="top-left" slot-scope="props" )
        .column
          .q-table-title Inquires
          br
          q-search.col-6(
            hide-underline
            color="secondary"
            v-model="filter")
      template( slot="top-right" slot-scope="props" )
        q-select(
          hide-underline
          :value="selectStage"
          @change="val => { selectStage = val; request() }"
          :options="stageOptions")
        q-table-columns(
          color="secondary"
          class="q-mr-sm"
          v-model="visibleColumns"
          :columns="columns")
</template>

<style>
</style>

<script>
import { QBtn, QCheckbox, QSearch, QSelect, QTable, QTableColumns, QTd, QTh, QTr, debounce } from 'quasar'

const imageRegex = /(jpe?g|png|gif|bmp)/

const debounceRequest = debounce(function (self, props) {
  self.loading = true
  if (!props) {
    props = {
      pagination: self.serverPagination,
      filter: self.filter
    }
  }

  self.$axios.get('/admin/whitelist/list', {params: Object.assign(props.pagination, {search: props.filter, stage: self.selectStage})})
    .then(function (response) {
      console.log(response)
      if (response.data) {
        self.serverData = response.data.data || []
        self.serverPagination = response.data.pagination
      }
    })
    .catch(function (error) {
      console.log(error.response)
    })
    .finally(function () {
      self.loading = false
    })
}, 500)

export default {
  data () {
    return {
      serverData: [],
      serverPagination: {
        page: 1,
        rowsNumber: 10 // specifying this determines pagination is server-side
      },
      columns: [
        {
          name: 'name',
          required: true,
          label: 'Full name',
          align: 'left',
          field: 'Name',
          sortable: true
        },
        { name: 'birthday', label: 'Birthday', field: 'Birthday', sortable: true },
        { name: 'country', label: 'Country', field: 'Country', sortable: true },
        { name: 'action', label: 'Action' }
      ],

      filter: '',
      visibleColumns: ['name', 'birthday', 'country', 'action'],
      selected: [
        // initial selection
        { name: 'Ice cream sandwich' }
      ],
      loading: false,
      loadingActions: false,
      imageRegex,
      selectStage: 'all',
      stageOptions: [
        {
          label: 'All confirmed',
          icon: 'fa-list-ul',
          value: 'all'
        },
        {
          label: 'Declined',
          icon: 'fa-times',
          color: 'negative',
          value: 'declined'
        },
        {
          label: 'Question',
          icon: 'fa-question',
          value: 'question'
        },
        {
          label: 'Accepted',
          icon: 'fa-check',
          value: 'accepted'
        },
        {
          label: 'Confirmed',
          icon: 'fa-user',
          value: 'confirmed'
        },
        {
          label: 'Unconfirmed',
          icon: 'fa-user-secret',
          value: 'unconfirmed'
        }
      ]
    }
  },
  methods: {
    downloadBase64File (str, fileName) {
      str = str.replace(/^data:[^;]+/, 'data:application/octet-stream')
      var link = document.createElement('a')
      link.download = fileName
      link.href = str
      link.click()
    },
    request (props) {
      debounceRequest(this, props)
    },
    isDisableAction (stageId, current) {
      switch (stageId) {
        case 2:
          return current === 'declined'
        case 3:
          return current === 'question'
        case 4:
          return true
      }

      return false
    },
    actionRequest (id, action) {
      if (this.loadingActions) return
      let self = this
      this.loadingActions = true
      this.$axios.post('/admin/whitelist/' + action + '/' + id)
        .then(function (response) {
          console.log(response)
          self.request()
        })
        .catch(function (error) {
          console.log(error.response)
        })
        .finally(function () {
          self.loadingActions = false
        })
    }
  },
  mounted () {
    this.request()
  },
  components: { QBtn, QCheckbox, QSearch, QSelect, QTable, QTableColumns, QTd, QTh, QTr }
}
</script>

<style lang="stylus">
  .docs-table
    .q-btn
      margin 5px
    tr.image-row
      img
        max-width 100%
      .td-column
        padding .5rem 1rem
</style>

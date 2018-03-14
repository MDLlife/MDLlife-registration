<template>
  <div
      class="q-input-file relative-position"
      :class="classes"
      @dragover.prevent.stop="__onDragOver"
  >
    <q-input-frame
        ref="input"

        :prefix="prefix"
        :suffix="suffix"
        :stack-label="stackLabel"
        :float-label="floatLabel"
        :error="error"
        :warning="warning"
        :disable="disable"
        :inverted="inverted"
        :invertedLight="invertedLight"
        :dark="dark"
        :hide-underline="hideUnderline"
        :before="before"
        :after="after"
        :color="color"
        :align="align"
        :no-parent-field="noParentField"

        :length="queueLength"
        additional-length
    >
      <div class="col q-input-target"
           @click.native="__pick"
           :disabled="addDisabled"
      >
        <div
            class="col ellipsis"
            :class="alignClass"
        >
          {{ label }}
        </div>

        <q-icon
            :name="$q.icon.uploader.add"
            class="q-input-file-pick-button q-if-control relative-position overflow-hidden"
        >
        </q-icon>
        <input
            type="file"
            ref="file"
            class="q-input-file-input absolute-full cursor-pointer"
            :accept="extensions"
            v-bind.prop="{multiple: multiple}"
            @change="__add"
        >
      </div>

      <q-icon
          v-if="hasExpandedContent"
          slot="after"
          :name="$q.icon.uploader.expand"
          class="q-if-control generic_transition"
          :class="{'rotate-180': expanded}"
          @click.native="expanded = !expanded"
      />
    </q-input-frame>

    <q-slide-transition>
      <div v-show="expanded">
        <q-list :dark="dark" class="q-input-file-files q-py-none scroll" :style="filesStyle">
          <q-item
              v-for="file in files"
              :key="file.name + file.__timestamp"
              class="q-input-file-file q-pa-xs"
          >

            <q-item-side v-if="file.__img" :image="file.__img.src"/>
            <q-item-side v-else :icon="$q.icon.uploader.file" :color="color"/>

            <q-item-main :label="file.name" :sublabel="file.__size"/>

            <q-item-side right>
              <q-item-tile
                  :icon="$q.icon.uploader.clear"
                  :color="color"
                  class="cursor-pointer"
                  @click.native="__remove(file)"
              />
            </q-item-side>
          </q-item>
        </q-list>
      </div>
    </q-slide-transition>

    <div
        v-if="dnd"
        class="q-input-file-dnd flex row items-center justify-center absolute-full"
        :class="dndClass"
        @dragenter.prevent.stop
        @dragover.prevent.stop
        @dragleave.prevent.stop="__onDragLeave"
        @drop.prevent.stop="__onDrop"
    ></div>
  </div>
</template>

<script>
import {
  QInputFrame,
  QSpinner,
  QIcon,
  QProgress,
  QItem,
  QItemSide,
  QItemMain,
  QItemTile,
  QList,
  QSlideTransition,
  format
} from 'quasar'
import FrameMixin from 'quasar-framework/src/mixins/input-frame'
const { humanStorageSize } = format

export default {
  name: 'QInputFile',
  mixins: [FrameMixin],
  components: {
    QInputFrame,
    QSpinner,
    QIcon,
    QProgress,
    QList,
    QItem,
    QItemSide,
    QItemMain,
    QItemTile,
    QSlideTransition
  },
  props: {
    value: {
      type: Array,
      default: () => []
    },
    extensions: String,
    multiple: Boolean,
    noThumbnails: Boolean,
    autoExpand: Boolean,
    expandStyle: [Array, String, Object],
    expandClass: [Array, String, Object]
  },
  data () {
    return {
      queue: [],
      totalSize: 0,
      focused: false,
      dnd: false,
      expanded: false,
      files: this.value
    }
  },
  computed: {
    queueLength () {
      return this.queue.length
    },
    hasExpandedContent () {
      return this.files.length > 0
    },
    label () {
      const total = humanStorageSize(this.totalSize)
      return `${this.queueLength} (${total})`
    },
    addDisabled () {
      return !this.multiple && this.queueLength >= 1
    },
    filesStyle () {
      if (this.maxHeight) {
        return { maxHeight: this.maxHeight }
      }
    },
    dndClass () {
      const cls = [`text-${this.color}`]
      if (this.isInverted) {
        cls.push('inverted')
      }
      return cls
    },
    classes () {
      return {
        'q-input-file-expanded': this.expanded,
        'q-input-file-dark': this.dark,
        'q-input-file-files-no-border': this.isInverted || !this.hideUnderline
      }
    },
    computedExtensions () {
      if (this.extensions) {
        return this.extensions.split(',').map(ext => {
          ext = ext.trim()
          // support "image/*"
          if (ext.endsWith('/*')) {
            ext = ext.slice(0, ext.length - 1)
          }
          return ext
        })
      }
    }
  },
  watch: {
    hasExpandedContent (v) {
      if (v === false) {
        this.expanded = false
      } else if (this.autoExpand) {
        this.expanded = true
      }
    },
    value (v) {
      this.files = this.value
    }
  },
  methods: {
    __onDragOver () {
      this.dnd = true
    },
    __onDragLeave () {
      this.dnd = false
    },
    __onDrop (e) {
      this.dnd = false
      let files = e.dataTransfer.files
      if (files.length === 0) {
        return
      }
      files = this.multiple ? files : [ files[0] ]
      if (this.extensions) {
        files = this.__filter(files)
        if (files.length === 0) {
          return
        }
      }
      this.__add(null, files)
    },
    __filter (files) {
      return Array.prototype.filter.call(files, file => {
        return this.computedExtensions.some(ext => {
          return file.type.toUpperCase().startsWith(ext.toUpperCase()) ||
              file.name.toUpperCase().endsWith(ext.toUpperCase())
        })
      })
    },
    __add (e, files) {
      if (this.addDisabled) {
        return
      }
      files = Array.prototype.slice.call(files || e.target.files)
      this.$refs.file.value = ''
      let filesReady = [] // List of image load promises
      files = files.filter(file => !this.queue.some(f => file.name === f.name))
        .map(file => {
          file.__size = humanStorageSize(file.size)
          file.__timestamp = new Date().getTime()
          if (this.noThumbnails || !file.type.toUpperCase().startsWith('IMAGE')) {
            this.queue.push(file)
          } else {
            const reader = new FileReader()
            let p = new Promise((resolve, reject) => {
              reader.onload = (e) => {
                let img = new Image()
                img.src = e.target.result
                file.__img = img
                this.queue.push(file)
                this.__computeTotalSize()
                resolve(true)
              }
              reader.onerror = (e) => {
                reject(e)
              }
            })
            reader.readAsDataURL(file)
            filesReady.push(p)
          }
          return file
        })
      if (files.length > 0) {
        this.files = this.files.concat(files)
        Promise.all(filesReady).then(() => {
          this.$emit('add', files)
          this.$emit('input', this.files)
        })
        this.__computeTotalSize()
      }
    },
    __computeTotalSize () {
      this.totalSize = this.queueLength
        ? this.queue.map(f => f.size).reduce((total, size) => total + size)
        : 0
    },
    __remove (file) {
      const
        name = file.name
      this.queue = this.queue.filter(obj => obj.name !== name)
      file.__removed = true
      this.files = this.files.filter(obj => obj.name !== name)
      this.__computeTotalSize()
    },
    __pick () {
      if (!this.addDisabled && this.$q.platform.is.mozilla) {
        this.$refs.file.click()
      }
    }
  }
}
</script>

<style lang="stylus">
  .q-input-file-expanded .q-if
    border-bottom-left-radius 0
    border-bottom-right-radius 0

  .q-input-file-input
    opacity 0
    max-width 100%
    min-height 100%
    width 100%
    bottom -8px
    font-size 0
  .q-input-file-pick-button[disabled] .q-input-file-input
    display none
  .q-input-target[disabled] .q-input-file-input
    display none

  .q-input-target:hover .q-if-control
    opacity 0.7

  .q-input-file-files
    border 1px solid $grey-4
    font-size 14px
    max-height 500px
  .q-input-file-files-no-border .q-input-file-files
    border-top 0 !important
  .q-input-file-file:not(:last-child)
    border-bottom 1px solid $grey-4
  .q-input-file-progress-bg, .q-input-file-progress-text
    pointer-events none
  .q-input-file-progress-bg
    height 100%
    opacity .2
  .q-input-file-progress-text
    font-size 40px
    opacity .1
    right 44px
    bottom 0

  .q-input-file-dnd
    outline 2px dashed currentColor
    outline-offset -6px
    background rgba(255, 255, 255, .6)
    &.inverted
      background rgba(0, 0, 0, .3)

  .q-input-file-dark
    .q-input-file-files
      color white
      border 1px solid $field-dark-label-color
    .q-input-file-bg
      color white
    .q-input-file-progress-text
      opacity .2
    .q-input-file-file:not(:last-child)
      border-bottom 1px solid $dark
</style>

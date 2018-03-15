import axios from 'axios'
import Config from 'src/config'

export default ({ Vue }) => {
  axios.defaults.baseURL = Config('api.api_url') // global config

  // config for vue files, global not works
  Vue.prototype.$axios = axios.create({
    baseURL: Config('api.api_url')
  })
}

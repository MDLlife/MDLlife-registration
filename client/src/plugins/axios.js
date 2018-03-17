import axios from 'axios'
import Config from 'src/config'

export default ({ Vue }) => {
  axios.defaults.baseURL = Config('api.api_url') // global config

  // apply global instance
  Vue.prototype.$axios = axios
}

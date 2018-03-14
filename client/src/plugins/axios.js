import axios from 'axios'
import Config from 'src/config'

export default ({ Vue }) => {
  Vue.prototype.$axios = axios.create({
    baseURL: Config('api.api_url')
  })
}

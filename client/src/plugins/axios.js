import axios from 'axios'
import Config from 'src/config'
import Router from 'src/router'

export default ({ Vue }) => {
  axios.defaults.baseURL = Config('api.api_url') // global config

  // apply global instance
  Vue.prototype.$axios = axios

  Vue.prototype.$axios.interceptors.response.use((response) => {
    return response
  }, err => {
    const error = err.response || err
    if (error.status === 401 && error.config && !error.config.__isRetryRequest && Router.currentRoute.path !== '/login') {
      Router.replace('/logout')
    }

    return Promise.reject(error)
  })
}

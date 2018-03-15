import Http from 'axios'
import Config from 'src/config'

export default {
  async attemptLogin (credentials) {
    try {
      let response = await Http.get(Config('api.basic_auth'), {
        auth: credentials,
        withCredentials: true
      })
      return new Promise(resolve => resolve(response))
    } catch (error) {
      return new Promise((resolve, reject) => reject(error))
    }
  },
  addAuthorizationHeader (header) {
    Http.defaults.headers.common['Authorization'] = header
  },
  registerInterceptor (cb) {
    Http.interceptors.response.use((response) => {
      return response
    }, err => {
      const error = err.response

      if (error.status === 401 && error.config && !error.config.__isRetryRequest) {
        if (cb) cb()
      }

      return Promise.reject(error)
    })
  }
}

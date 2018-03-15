/***************************************************
 * [Quasar Cookies] http://quasar-framework.org/components/cookies.html
 * [Quasar LocalStorage] http://quasar-framework.org/components/web-storage.html
 **************************************************/

import { Cookies, LocalStorage } from 'quasar'
import AuthService from './auth.service'

class BasicAuth {
  constructor () {
    this.storages = {
      Cookies,
      LocalStorage
    }
    this.session = this.storages['LocalStorage']
  }

  logout () {
    this.session.remove('basic_auth')
    AuthService.addAuthorizationHeader('')
  }

  guest () {
    return !this.session.has('basic_auth')
  }

  isAuthenticated () {
    return this.session.has('basic_auth')
  }

  login (username, password) {
    let self = this
    let data = {
      username: username,
      password: password
    }

    // We merge grant type and client secret stored in configuration
    return new Promise((resolve, reject) => {
      AuthService.attemptLogin(data)
        .then(response => {
          self.storeSession(data)
          self.addAuthHeaders()
          AuthService.registerInterceptor(function () {
            self.$router.replace('/logout')
          })
          resolve(response)
        })
        .catch(error => {
          console.log('Authentication error: ', error)
          reject(error)
        })
    })
  }

  getAuthHeader () {
    if (this.session.has('basic_auth')) {
      return this.getItem('basic_auth')
    }
    return null
  }

  getItem (key) {
    return this.session.get.item(key)
  }

  addAuthHeaders () {
    let header = this.getAuthHeader()
    AuthService.addAuthorizationHeader(header)
  }

  storeSession (data) {
    this.session.set('basic_auth', 'Basic ' + btoa(data.username + ':' + data.password))
  }
}

export default BasicAuth

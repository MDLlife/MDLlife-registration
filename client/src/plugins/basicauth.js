import BasicAuth from 'src/app/auth'

export default ({ app, router, Vue }) => {
  Vue.prototype.$basicauth = new BasicAuth()
}

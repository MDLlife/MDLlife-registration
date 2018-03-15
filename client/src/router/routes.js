import BasicAuth from 'src/app/auth'
const basicAuth = new BasicAuth()

export default [
  {
    path: '/',
    component: () => import('layouts/default'),
    children: [
      { path: '', component: () => import('pages/request') },
      { path: 'login', name: 'app.login', component: () => import('pages/login') }
    ]
  },

  {
    path: '/',
    component: () => import('layouts/default'),
    beforeEnter: requireAuth,
    children: [
      { path: 'check', component: () => import('pages/check') }
    ]
  },

  { path: '/logout',
    name: 'app.logout',
    beforeEnter (to, from, next) {
      basicAuth.logout()
      next('/')
    }
  },

  { // Always leave this as last one
    path: '*',
    component: () => import('pages/404')
  }
]

function requireAuth (to, from, next) {
  if (!basicAuth.isAuthenticated()) {
    next({
      path: '/login',
      query: { redirect: to.fullPath }
    })
  } else {
    next()
  }
}

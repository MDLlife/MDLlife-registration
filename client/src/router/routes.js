
export default [
  {
    path: '/',
    component: () => import('layouts/default'),
    beforeEnter: requireAuth,
    children: [
      { path: 'check', component: () => import('pages/check') }
    ]
  },

  {
    path: '/',
    component: () => import('layouts/default'),
    children: [
      { path: 'request', component: () => import('pages/request') }
    ]
  },

  { path: '/login', name: 'app.login', component: () => import('pages/login') },
  { path: '/logout',
    name: 'app.logout',
    beforeEnter (to, from, next) {
      // auth.logout()
      next('/')
    }
  },

  { // Always leave this as last one
    path: '*',
    component: () => import('pages/404')
  }
]

function requireAuth (to, from, next) {
  // if (!auth.loggedIn()) {
  // next({
  //   path: '/login',
  //   query: { redirect: to.fullPath }
  // })
  // } else {
  next()
  // }
}

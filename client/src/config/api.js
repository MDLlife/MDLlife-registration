const api_url = env('BASE_URL', 'http://localhost:8000/')

export default {
  api_url: api_url,
  add_whitelist: 'whitelist/request',
  captcha_id: 'captcha/id',
  captcha: api_url + 'captcha/',

  basic_auth: 'admin/basic-auth'
}

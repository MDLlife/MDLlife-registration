const api_url = env('BASE_URL', 'http://localhost:8000/')

export default {
  api_url: api_url,
  add_whitelist: api_url + 'whitelist/request',
  captcha_id: api_url + 'captcha/id',
  captcha: api_url + 'captcha/'
}

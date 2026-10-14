import type { paths } from '@/schemas/types'
import createClient from 'openapi-fetch'

// Use relative path in dev (Vite proxy handles /api -> localhost:3000)
// Use full URL if running standalone
const baseUrl = import.meta.env.PROD
  ? `http://${window.location.hostname}:3000`
  : ''

export const api = createClient<paths>({ baseUrl })

api.use({
  onRequest({ request }) {
    const token = localStorage.getItem('access_token')
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`)
    }
    // If baseUrl is empty (dev mode), ensure URL is relative
    if (!baseUrl && !request.url.startsWith('http')) {
      // openapi-fetch handles relative URLs automatically with empty baseUrl
    }
    return request
  },
})

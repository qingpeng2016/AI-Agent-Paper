import type { ApiEnvelope } from '../types/user'
import type { RegisterCouponPromo, UserCouponList } from '../types/coupon'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createCouponApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async registerPromo(): Promise<RegisterCouponPromo> {
      const res = await http.get<ApiEnvelope<RegisterCouponPromo>>(
        '/api/v1/coupon-campaigns/register-promo',
      )
      return unwrap(res)
    },
    async listMine(): Promise<UserCouponList> {
      const res = await http.get<ApiEnvelope<UserCouponList>>('/api/v1/users/coupons')
      return unwrap(res)
    },
  }
}

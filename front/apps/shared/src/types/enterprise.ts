export type EnterpriseProduct = {
  code: string
  name: string
  badge?: string
  price_hint: string
  seats: string
  features: string[]
  tagline: string
  button_label: string
  is_featured: boolean
  sort: number
}

export type EnterpriseProductListResponse = {
  products: EnterpriseProduct[]
}

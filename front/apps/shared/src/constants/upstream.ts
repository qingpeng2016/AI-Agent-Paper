import type { ProductsCategoryName } from '../types/product'

export const UPSTREAM_LABEL: Record<ProductsCategoryName, string> = {
  openai: 'ChatGPT',
  anthropic: 'Claude',
  xai: 'Grok',
  gemini: 'Gemini',
  perplexity: 'Perplexity',
  cursor: 'Cursor',
}

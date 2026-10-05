export type NavPlanItem = {
  label: string
  price: string
  href: string
  featured?: boolean
}

export type NavMegaColumn = {
  title: string
  items: NavPlanItem[]
}

const p = {
  gptGo: '/p/gpt-go',
  chatgptPlus: '/p/chatgpt-plus',
  pro5x: '/p/chatgpt-pro-5x',
  pro20x: '/p/chatgpt-pro-20x',
  claudePro: '/p/claude-pro',
  claudeMax5x: '/p/claude-max-5x',
  grokSuper: '/p/grok-super',
  geminiPro: '/p/gemini-pro',
  cursorPro: '/p/cursor-pro',
  cursorProPlus: '/p/cursor-pro-plus',
  perplexityPro: '/p/perplexity-pro',
  catalog: '/#catalog',
} as const

/** 套餐购买 — 大面板 */
export const navMegaMenu: NavMegaColumn[] = [
  {
    title: 'ChatGPT',
    items: [
      { label: 'ChatGPT Plus', price: '¥178/月', href: p.chatgptPlus },
      { label: 'GPT Go', price: '¥89/月', href: p.gptGo },
      { label: 'Pro 5X', price: '¥899/月', href: p.pro5x },
      { label: 'Pro 20X', price: '¥1799/月', href: p.pro20x },
      { label: 'Codex', price: '¥599/月', href: p.pro5x },
      { label: 'Images 2.5', price: '¥998/月', href: p.chatgptPlus },
    ],
  },
  {
    title: 'Claude',
    items: [
      { label: 'Claude Pro', price: '¥219/月', href: p.claudePro },
      { label: 'Claude Max 5X', price: '¥999/月', href: p.claudeMax5x },
      { label: 'Claude Max 20X', price: '¥1899/月', href: p.claudeMax5x },
      { label: 'Claude Code', price: '¥599/月', href: p.claudePro },
    ],
  },
  {
    title: 'Grok',
    items: [
      { label: 'SuperGrok', price: '¥790/季', href: p.grokSuper },
      { label: 'SuperGrok Heavy', price: '¥790/季', href: p.grokSuper },
      { label: 'Grok Imagine', price: '¥590/季', href: p.grokSuper },
    ],
  },
  {
    title: 'Gemini',
    items: [
      { label: 'Google AI Pro', price: '¥189/月', href: p.geminiPro },
      { label: 'Google AI Ultra', price: '¥1800/月', href: p.geminiPro },
      { label: 'Nano Banana / Veo', price: '¥199/月', href: p.geminiPro },
    ],
  },
  {
    title: 'Cursor',
    items: [
      { label: 'Cursor Pro', price: '¥198/月', href: p.cursorPro },
      { label: 'Cursor Pro+', price: '¥598/月', href: p.cursorProPlus },
      { label: 'Cursor Ultra', price: '¥1280/月', href: p.cursorProPlus },
      { label: 'Cloud Agents', price: '¥998/月', href: p.cursorProPlus },
    ],
  },
]

export type NavDropItem = {
  label: string
  href: string
  price?: string
  featured?: boolean
}

export type NavMenuDropdown = {
  id: string
  label: string
  hotTagName?: string
  items: NavDropItem[]
}

export type NavMenuPlainLink = {
  id: string
  label: string
  plainLink: true
  to: string
}

export type NavMenuEntry = NavMenuDropdown | NavMenuPlainLink

export function isNavDropdown(entry: NavMenuEntry): entry is NavMenuDropdown {
  return !('plainLink' in entry)
}

export const navBrandDropdowns: NavMenuEntry[] = [
  {
    id: 'chatgpt',
    label: 'ChatGPT',
    items: navMegaMenu[0].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  },
  {
    id: 'claude',
    label: 'Claude',
    items: navMegaMenu[1].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  },
  {
    id: 'cursor',
    label: 'Cursor',
    items: navMegaMenu[4].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  },
  {
    id: 'enterprise',
    label: '企业采购',
    plainLink: true,
    to: '/enterprise',
  },
]

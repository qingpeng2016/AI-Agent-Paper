export interface TutorialCategory {
  id: number
  code: string
  name: string
  sort: number
}

export interface TutorialArticle {
  slug: string
  category_code: string
  category_name: string
  date: string
  title: string
  excerpt: string
  body?: string[]
}

export interface TutorialListResponse {
  title: string
  categories: TutorialCategory[]
  articles: TutorialArticle[]
}

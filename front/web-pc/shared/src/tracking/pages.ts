/** 路由 → 埋点 page_id / page_title（轨迹文案固定简体中文） */
export function trackMetaForRoute(
  name: string | symbol | null | undefined,
  fullPath: string,
): { page_id: string; page_path: string; page_title: string } {
  const n = name == null ? '' : String(name)
  switch (n) {
    case 'home':
      return { page_id: 'home', page_path: fullPath, page_title: '首页' }
    case 'product':
      return { page_id: 'product_detail', page_path: fullPath, page_title: '商品详情' }
    case 'member':
      return { page_id: 'member', page_path: fullPath, page_title: '会员中心' }
    case 'enterprise':
      return { page_id: 'enterprise', page_path: fullPath, page_title: '企业采购' }
    case 'blog':
      return { page_id: 'blog', page_path: fullPath, page_title: '教程列表' }
    case 'blog-article':
      return { page_id: 'blog_article', page_path: fullPath, page_title: '教程文章' }
    case 'login':
      return { page_id: 'login', page_path: fullPath, page_title: '登录' }
    case 'register':
      return { page_id: 'register', page_path: fullPath, page_title: '注册' }
    default:
      return { page_id: n || 'unknown', page_path: fullPath, page_title: n || '页面' }
  }
}

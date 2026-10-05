import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath, pathToFileURL } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const blogModule = await import(
  pathToFileURL(path.join(__dirname, '../src/mocks/blog.ts')).href,
)
const { blogPosts } = blogModule

function esc(s) {
  return String(s).replace(/\\/g, '\\\\').replace(/'/g, "''")
}

function jsonBody(paragraphs) {
  const parts = paragraphs.map((p) => `'${esc(p)}'`)
  return `JSON_ARRAY(${parts.join(', ')})`
}

const lines = [
  '-- 从前端 mocks/blog.ts 导入的全部教程文章（可重复执行）',
  '',
]

blogPosts.forEach((post, idx) => {
  lines.push(
    `INSERT INTO \`tutorial_article\` (\`category_id\`, \`slug\`, \`title\`, \`excerpt\`, \`body\`, \`published_at\`, \`sort\`, \`status\`)`,
    `SELECT c.id, '${esc(post.slug)}', '${esc(post.title)}', '${esc(post.excerpt)}', ${jsonBody(post.body)}, '${post.date}', ${1000 - idx}, 'published'`,
    `FROM \`tutorial_category\` c WHERE c.code = '${esc(post.category)}' LIMIT 1`,
    `ON DUPLICATE KEY UPDATE`,
    `  \`category_id\` = VALUES(\`category_id\`),`,
    `  \`title\` = VALUES(\`title\`),`,
    `  \`excerpt\` = VALUES(\`excerpt\`),`,
    `  \`body\` = VALUES(\`body\`),`,
    `  \`published_at\` = VALUES(\`published_at\`),`,
    `  \`sort\` = VALUES(\`sort\`),`,
    `  \`status\` = VALUES(\`status\`);`,
    '',
  )
})

const out = path.join(
  __dirname,
  '../../../../docs/migrations/20260928_tutorial_articles_seed.sql',
)
fs.writeFileSync(out, lines.join('\n'), 'utf8')
console.log(`Wrote ${blogPosts.length} articles to ${out}`)

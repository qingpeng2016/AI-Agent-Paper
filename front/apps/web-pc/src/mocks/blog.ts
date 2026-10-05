export type BlogCategoryId =
  | 'all'
  | 'plus'
  | 'pro'
  | 'perplexity'
  | 'payment'
  | 'tutorial'
  | 'quality'
  | 'account'
  | 'pricing'

export interface BlogCategory {
  id: BlogCategoryId
  label: string
}

export interface BlogPost {
  slug: string
  category: Exclude<BlogCategoryId, 'all'>
  categoryLabel: string
  date: string
  title: string
  excerpt: string
  body: string[]
}

export const blogPageMeta = {
  title: 'AI Plan 教程与资讯',
}

export const blogCategories: BlogCategory[] = [
  { id: 'all', label: '全部' },
  { id: 'plus', label: 'Plus 代充' },
  { id: 'pro', label: 'Pro 升级' },
  { id: 'perplexity', label: 'Perplexity' },
  { id: 'payment', label: '充值 / 支付' },
  { id: 'tutorial', label: '功能教程' },
  { id: 'quality', label: '降智 / 认知' },
  { id: 'account', label: '账号 / 注册' },
  { id: 'pricing', label: '价格 / 对比' },
]

export const blogPosts: BlogPost[] = [
  {
    slug: 'payment-card-shop-security',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-09-23',
    title: '假如签名能被伪造，你的发卡站还会发货吗？',
    excerpt:
      '以「攻击者能构造任意签名合法的请求」为判定标准，按下单、支付、回调、发货七站走查支付链路，附自查表与应急流程。',
    body: [
      '自营或对接发卡时，最容易出问题的是回调验签与发货幂等。任何「只验参数、不验签名」或「签名算法可被重放」的环节，都可能被 0 元购。',
      '建议按链路逐站自查：下单金额是否与回调一致、回调 IP 是否可伪造、发货是否依赖可预测订单号。二开版本尤其要对照官方补丁。',
      '应急上先暂停自动发货、冻结可疑订单，再轮换密钥并补审计日志。AI Plan 代充走平台统一支付，用户侧无需自建发卡链路。',
    ],
  },
  {
    slug: 'cursor-payment-china',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-08-20',
    title: 'Cursor 国内怎么付款？虚拟卡 / 双币卡被拒的解法',
    excerpt:
      'Cursor 订阅只收海外信用卡，国内双币卡和多数虚拟卡会在 3DS 或风控被拒。讲清原因与国内支付宝 / 微信代充到自有账号的做法。',
    body: [
      'Cursor 账单走 Stripe 类通道，国内银行常因 3DS、账单地址或风控直接失败，并非「卡有问题」这么简单。',
      '虚拟卡成功率因 issuer 而异，反复试卡可能触发账号风控。更省心的路径是用平台代充，充到你已注册的 Cursor 邮箱账号。',
      '开通后在会员中心查看订单与教程，按文档配置 BYOK 或 Cloud Agents 即可。',
    ],
  },
  {
    slug: 'cursor-pro-plus-vs-ultra',
    category: 'pro',
    categoryLabel: 'Pro 升级',
    date: '2026-08-19',
    title: 'Cursor Pro+ vs Ultra：Agent 额度怎么算才划算',
    excerpt:
      'Pro+ 与 Ultra 主要差在 Agent 额度倍数，可用模型与功能一致。讲清额度消耗、Auto 模式与按用量选档。',
    body: [
      'Pro 是基础档；Pro+ 约 3× Agent 额度，Ultra 约 20×。日常补全与小任务 Pro 够用，长跑 Agent 流水线才需要考虑升档。',
      'Auto 模式在官方策略下通常不计入 Agent 额度，但具体以 Cursor 当期说明为准。',
      '国内可在 AI Plan 选购对应 SKU，失败按平台规则退款或重试。',
    ],
  },
  {
    slug: 'cursor-account-banned',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-08-19',
    title: 'Cursor 账号被封怎么办？常见原因与申诉思路',
    excerpt:
      '梳理反代、异常登录、拼车连坐等常见触发点，以及申诉与降低风险的实用建议。',
    body: [
      '先区分「订阅掉了」和「账号封禁」：前者多为支付失败，后者常伴随邮件或无法登录。',
      '共用账号、多地同时登录、使用未授权中转，都容易触发风控。独享 + 固定干净环境更稳。',
      '代充到自有账号并保留订单凭证，便于平台协助核对与质保。',
    ],
  },
  {
    slug: 'cursor-carpool-vs-solo',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-08-18',
    title: 'Cursor 拼车 vs 独享：掉车、连坐与团队设置',
    excerpt:
      '拼车便宜但掉车与连坐风险高。对比独享代充到本人账号的真实成本与风险。',
    body: [
      '拼车本质是多人共用权限，一人违规可能导致整车不可用，且往往无法自行改密或绑定。',
      '独享是你自己的 Cursor 账号，代充只负责订阅周期，权限边界清晰。',
      '选购时优先看清是否「充到本人账号」以及失败 / 封号质保口径。',
    ],
  },
  {
    slug: 'cursor-agent-quota-upgrade',
    category: 'pro',
    categoryLabel: 'Pro 升级',
    date: '2026-08-18',
    title: 'Cursor 撞额度了？该升 Pro+ 还是 Ultra',
    excerpt:
      '先分清用法问题还是真该升档，再按撞限频率在 Pro / Pro+ / Ultra 之间选择。',
    body: [
      '频繁撞限前先检查是否重复跑大上下文、是否误用高消耗 Agent 模式。',
      '若每天多次触顶，Pro+ 往往是甜点；全天满负荷流水线再考虑 Ultra。',
      '升档可在平台补购对应套餐，不必换账号。',
    ],
  },
  {
    slug: 'chatgpt-carpool-vs-solo',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-07-06',
    title: 'ChatGPT 拼车 vs 独享：掉车、封号与连坐风险（2026）',
    excerpt:
      'Team 车与合租便宜，但一号多人用易连累整车。独享可改密、风险边界清晰。',
    body: [
      '拼车常见问题是多地登录、成员拒付或违规导致整车失效，且隐私不可控。',
      '独享代充到你自己注册的 OpenAI 账号，订阅状态可在官网与会员中心对照。',
      '配合失败退款政策，掉订阅不等于白花钱。',
    ],
  },
  {
    slug: 'gemini-student-offer-expire',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-07-06',
    title: 'Gemini 学生优惠到期 / 被取消怎么续？',
    excerpt:
      'Google AI Pro 学生权益失效后的续费路径：重新验证或正规代充，勿伪造学生资格。',
    body: [
      '学生免费档到期后会回到免费版，Pro 功能与云盘额度一并降级。',
      '仍符合学生条件可重新走 Google 验证；否则用 AI Plan 代充 Google AI Pro / Ultra。',
      '伪造学生信息可能导致 Google 账号风险，不建议走灰色路径。',
    ],
  },
  {
    slug: 'chatgpt-subscription-lost',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-07-06',
    title: 'ChatGPT 账号被封 / 掉订阅了怎么办？',
    excerpt:
      '分清掉订阅与封禁，走官方申诉，并选择带质保的代充渠道止损。',
    body: [
      '掉订阅：Plus 变 Free，多为支付或地区问题，有时可恢复订阅。',
      '封禁：常收到邮件或无法登录，需按官方流程申诉，成功率因因而异。',
      '下单前确认平台失败退款与封号协助口径，降低不可逆损失。',
    ],
  },
  {
    slug: 'claude-max-vs-api',
    category: 'pricing',
    categoryLabel: '价格 / 对比',
    date: '2026-07-06',
    title: 'Claude Max 20X vs API：重度用户成本账',
    excerpt:
      '包月 Max 与按 token 计费的 API，在 Claude Code 重度场景下谁更省？',
    body: [
      'Max 是固定月费，适合持续、高强度对话与 Code；API 随用量线性增长，峰值账单可能远超 Max。',
      '需要自动化集成、自有后端调度时 API 更合适；个人重度编码 Max 往往更省心。',
      '国内 Max 可通过平台代充，无需海外卡。',
    ],
  },
  {
    slug: 'chatgpt-pro-5x-20x',
    category: 'pro',
    categoryLabel: 'Pro 升级',
    date: '2026-07-06',
    title: 'ChatGPT 撞额度 / 限速了？该升 Pro 吗',
    excerpt:
      'Plus 触顶时如何选 Pro 5X 与 20X，以及国内代充开通方式。',
    body: [
      'Plus 适合日常；Codex、Deep Research 等高频触顶说明需要更高配额档。',
      'Pro 5X 与 20X 主要差在用量上限，按你每周撞限次数选档即可。',
      '型号与额度以 OpenAI 当期政策为准，教程会随官方更新。',
    ],
  },
  {
    slug: 'chatgpt-pro-vs-claude-max',
    category: 'pricing',
    categoryLabel: '价格 / 对比',
    date: '2026-07-06',
    title: '重度用户怎么选：ChatGPT Pro vs Claude Max',
    excerpt:
      '按编程 / Agent 场景定生态，再按撞限频率定 5X 或 20X。',
    body: [
      'ChatGPT 侧强项在 Codex、Images 与生态整合；Claude 侧强项在长上下文与 Claude Code。',
      '不要两线同时买满配，先定主力再升档，避免重复订阅。',
      '平台提供多品牌 SKU，可按需单开或组合。',
    ],
  },
  {
    slug: 'claude-ban-email',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-07-06',
    title: 'Claude 封号邮件长什么样？如何确认真被封',
    excerpt:
      '区分网络问题、掉订阅与封禁，识别钓鱼邮件特征。',
    body: [
      '打不开 Claude 不一定是封号，先换节点与检查订阅状态。',
      '官方邮件通常来自 Anthropic 域名，勿点可疑链接输入密码。',
      '确认封禁后走申诉，并核对代充订单是否含协助与质保。',
    ],
  },
  {
    slug: 'claude-code-rate-limit',
    category: 'pro',
    categoryLabel: 'Pro 升级',
    date: '2026-07-06',
    title: 'Claude Code 老是限流 / 撞额度怎么办？',
    excerpt:
      '限流来自订阅档位与时间窗口，掉线多为环境或拼车问题。',
    body: [
      'Pro / Max 5X / 20X 对应不同用量上限，撞墙说明该升档或拆任务。',
      '精简上下文、分步提交、避免无意义重试，比多开小号更有效。',
      '独享账号 + 稳定网络，可显著减少「假掉线」。',
    ],
  },
  {
    slug: 'claude-carpool-risk',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-07-06',
    title: '拼车 vs 独享 Claude：为什么便宜拼车易连坐',
    excerpt:
      '一号多人用的风控与连坐风险，独享代充到本人账号的差异。',
    body: [
      '拼车无法保证其他成员的使用习惯，违规登录会波及全车。',
      '独享可自管密码与 2FA，代充仅覆盖订阅周期。',
      '选购时看清「充到谁账号」与质保条款。',
    ],
  },
  {
    slug: 'grok-subscribe-guide',
    category: 'pricing',
    categoryLabel: '价格 / 对比',
    date: '2026-07-02',
    title: 'Grok 值得订阅吗？SuperGrok 与国内开通',
    excerpt:
      'SuperGrok 与 Heavy 怎么选，以及国内支付路径简述。',
    body: [
      'Grok 强调实时信息与 X 生态整合，适合需要联网检索的场景。',
      'SuperGrok Heavy 面向更高配额需求，一般用户可从标准档试起。',
      '国内可通过平台选购对应 Grok SKU，充到自有 xAI 账号。',
    ],
  },
  {
    slug: 'gemini-china-guide',
    category: 'pricing',
    categoryLabel: '价格 / 对比',
    date: '2026-07-02',
    title: 'Gemini 国内怎么用？Google AI Pro / Ultra 选购',
    excerpt:
      '无海外卡如何开通，Nano Banana、Veo 等功能与档位区别。',
    body: [
      'Google AI Pro 覆盖 Gemini 高级能力与云存储；Ultra 面向更高需求。',
      '国内直连需稳定网络，支付可走 AI Plan 代充。',
      '功能名称与配额以 Google 官方为准，教程会定期更新。',
    ],
  },
  {
    slug: 'claude-vs-chatgpt-coding',
    category: 'quality',
    categoryLabel: '降智 / 认知',
    date: '2026-07-02',
    title: 'Claude vs ChatGPT 写代码怎么选（2026）',
    excerpt:
      '从工具链、上下文、Codex 与 Claude Code 体验对比，非跑分复测。',
    body: [
      'ChatGPT 侧 Codex 与 IDE 整合深；Claude Code 在长文件与指令遵循上口碑好。',
      '团队可双开但建议定主力，避免重复付费。',
      '两者国内均可代充，按已有账号生态选即可。',
    ],
  },
  {
    slug: 'claude-pro-max-choose',
    category: 'pricing',
    categoryLabel: '价格 / 对比',
    date: '2026-07-02',
    title: 'Claude Pro / Max 怎么选？5X、20X 建议',
    excerpt:
      '按用量与 Claude Code 频率在 Pro、Max 5X、20X 之间决策。',
    body: [
      '日常对话 Pro 足够；Code 高频触顶再升 Max 5X；全天满负荷考虑 20X。',
      'Max 与 Pro 模型访问策略以 Anthropic 官方为准。',
      '平台 SKU 与官方周期对齐，到期前可在会员中心续费。',
    ],
  },
  {
    slug: 'gpt-recharge-faq',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-06-16',
    title: 'GPT 充值常见问题：没到账、排队、能退吗',
    excerpt:
      '国内充值 GPTPlus 的高频问题一次答清。',
    body: [
      '正常代充约 1–5 分钟到账，高峰可能排队，以会员中心订单状态为准。',
      '长时间未到账先查邮箱与 OpenAI 订阅页，再联系客服带订单号。',
      '失败或未履约按平台规则退款，勿重复下单造成重复扣款。',
    ],
  },
  {
    slug: 'gpt-plus-buy-flow',
    category: 'plus',
    categoryLabel: 'Plus 代充',
    date: '2026-06-16',
    title: 'GPT 会员怎么买？2026 购买全流程',
    excerpt:
      '选套餐、支付宝 / 微信付款、激活到自有账号的三步说明。',
    body: [
      '在首页或商品页选择 Go / Plus / Pro 档，登录后下单。',
      '支付完成后系统自动履约，无需把密码交给客服。',
      '开通后在 OpenAI 官网与会员中心核对订阅周期。',
    ],
  },
  {
    slug: 'gpt-register-china',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-06-16',
    title: 'GPT 怎么注册？国内邮箱 / 谷歌账号教程',
    excerpt:
      '注册步骤、节点要求与常见失败原因。',
    body: [
      '可用邮箱或 Google 账号注册 OpenAI，部分时段需短信验证。',
      '注册阶段与使用阶段都建议稳定、干净的网络环境。',
      '注册完成后即可在平台代充 Plus，无需海外信用卡。',
    ],
  },
  {
    slug: 'gpt-payment-methods',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-06-16',
    title: 'GPT 怎么付款？银行卡被拒与支付宝代充',
    excerpt:
      'OpenAI 支付方式、国内卡被拒原因与代充方案。',
    body: [
      '官网直充主要面向海外卡，国内多数双币卡会被拒或 3DS 失败。',
      '虚拟卡并非万能，反复失败可能触发账号审查。',
      '支付宝 / 微信代充到自有账号是目前国内最省心路径。',
    ],
  },
  {
    slug: 'gptplus-getting-started',
    category: 'tutorial',
    categoryLabel: '功能教程',
    date: '2026-06-16',
    title: 'GPTPlus 新手怎么用？开通后第一件事',
    excerpt:
      '确认订阅、认识旗舰模型与常用功能入口。',
    body: [
      '先在 OpenAI 设置中确认 Plus / Pro 已生效。',
      '根据当期官方说明选择对话模型与 Codex、出图等功能入口。',
      '遇到限速先查用量，再考虑升 Pro 档。',
    ],
  },
  {
    slug: 'gpt-deep-research-limits',
    category: 'tutorial',
    categoryLabel: '功能教程',
    date: '2026-06-16',
    title: 'GPT Deep Research 次数限制与用法',
    excerpt:
      '各档次数差异与如何把每次研究用在刀刃上。',
    body: [
      'Deep Research 消耗专属配额，Plus 与 Pro 上限不同，数值随官方调整。',
      '用清晰的研究问题与范围，避免重复跑同一课题浪费次数。',
      '触顶后可等待窗口刷新或升档。',
    ],
  },
  {
    slug: 'gpt-dumb-detection',
    category: 'quality',
    categoryLabel: '降智 / 认知',
    date: '2026-06-16',
    title: 'GPT「降智」是什么？怎么检测和解决',
    excerpt:
      '限额降级、节点、上下文与模型选错的可操作排查。',
    body: [
      '触顶后模型行为变化常被误认为降智，先查用量与当前模型。',
      '过长上下文会拖累质量，必要时开新对话。',
      '网络不稳定会导致超时与半截回答，换节点再试。',
    ],
  },
  {
    slug: 'gpt-avoid-ban',
    category: 'account',
    categoryLabel: '账号 / 注册',
    date: '2026-06-16',
    title: '如何防止 GPT 账号被封？防风控习惯',
    excerpt:
      '固定干净登录环境、少共享、少违规自动化。',
    body: [
      '避免一号多人、多地乱跳 IP、使用来路不明中转。',
      '不要购买来路不明的成品号，代充到自注册账号更安全。',
      '敏感自动化需遵守 OpenAI 使用政策。',
    ],
  },
  {
    slug: 'choose-recharge-platform',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-06-16',
    title: '如何挑选靠谱的 GPT 代充平台？',
    excerpt:
      '五条标准：不交密码、失败退款、真人客服、价格异常、充到本人号。',
    body: [
      '任何索要账号密码的「代充」都该警惕。',
      '价格远低于市场需怀疑来源，失败应可退款或重试。',
      'AI Plan 充到用户自有账号，订单与客服可在站内追溯。',
    ],
  },
  {
    slug: 'recharge-vs-official',
    category: 'payment',
    categoryLabel: '充值 / 支付',
    date: '2026-06-16',
    title: '代充 vs 官网直充 vs 买账号：哪种靠谱',
    excerpt:
      '成功率、安全性与成本对比，国内用户怎么选。',
    body: [
      '官网直充最直但国内支付门槛高；买账号隐私与找回风险大。',
      '正规代充平衡了支付便利与账号归属。',
      '拼车最低成本但风险最高，不适合长期生产使用。',
    ],
  },
  {
    slug: 'gptplus-vs-pro',
    category: 'pro',
    categoryLabel: 'Pro 升级',
    date: '2026-06-16',
    title: 'GPTPlus vs Pro 怎么选？别一上来买最贵',
    excerpt:
      '按触顶频率在 Plus、Pro 5X、20X 之间分档。',
    body: [
      '大多数用户 Plus 即可；Codex / Agent 高频再升 Pro。',
      '20X 面向极重度，普通办公无需一步到位。',
      '可先 Plus 试跑一个月再决定是否升档。',
    ],
  },
  {
    slug: 'gpt-model-choose',
    category: 'quality',
    categoryLabel: '降智 / 认知',
    date: '2026-06-16',
    title: 'GPT 模型怎么选？档位与能力差异',
    excerpt:
      'Go / Plus / Pro 主要是额度与工具差异，选型以官方当期为准。',
    body: [
      '对话模型名称与能力随官方更新，勿用过期营销话术对照。',
      '编程优先看 Codex 是否在订内；出图看 Images 权益。',
      '不确定时从 Plus 起步，按用量升 Pro。',
    ],
  },
  {
    slug: 'perplexity-pro-guide',
    category: 'perplexity',
    categoryLabel: 'Perplexity',
    date: '2026-06-10',
    title: 'Perplexity Pro 值得开吗？与国内开通方式',
    excerpt:
      '检索增强对话适合什么人，以及平台代充路径。',
    body: [
      'Pro 提供更强模型与更多 Pro 搜索次数，适合调研与写作。',
      '与 ChatGPT 检索互补，可按需单开。',
      '国内可在 catalog 选购 Perplexity SKU。',
    ],
  },
  {
    slug: 'alipay-wechat-gptplus',
    category: 'plus',
    categoryLabel: 'Plus 代充',
    date: '2026-06-15',
    title: '国内怎么充 GPTPlus？支付宝 / 微信代充攻略',
    excerpt:
      '完整流程、避坑点与到账时间说明。',
    body: [
      '登录 → 选 Plus 月卡 → 扫码支付 → 等待自动履约。',
      '勿向私人转账「人工代充」，务必走站内订单。',
      '到账后 OpenAI 订阅页应显示 Plus 有效期。',
    ],
  },
]

export function getBlogPost(slug: string): BlogPost | undefined {
  return blogPosts.find((p) => p.slug === slug)
}

export function categoryLabel(id: BlogCategoryId): string {
  return blogCategories.find((c) => c.id === id)?.label ?? id
}

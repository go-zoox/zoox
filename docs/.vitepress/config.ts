import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: 'Zoox',
  description: 'A Lightweight, High Performance Go Web Framework',
  base: '/zoox/', // GitHub Pages 路径
  ignoreDeadLinks: true, // 忽略死链接（某些文档可能尚未创建）
  
  head: [
    ['link', { rel: 'icon', href: '/zoox/favicon.ico' }],
    ['meta', { name: 'keywords', content: 'go, golang, web framework, zoox, http, router, middleware' }],
  ],

  themeConfig: {
    // https://vitepress.dev/reference/default-theme-config
    logo: '/logo.svg',

    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/getting-started/installation' },
      { text: '指南', link: '/guides/routing' },
      { text: 'API 参考', link: '/api-reference/application' },
      { text: '示例', link: '/examples/rest-api' },
      { text: 'GitHub', link: 'https://github.com/go-zoox/zoox' },
    ],

    sidebar: {
      '/getting-started/': [
        {
          text: '快速开始',
          items: [
            { text: '安装指南', link: '/getting-started/installation' },
            { text: '5分钟快速开始', link: '/getting-started/quick-start' },
            { text: '第一个应用', link: '/getting-started/first-app' },
            { text: '常见示例', link: '/getting-started/examples' },
          ],
        },
      ],
      '/guides/': [
        {
          text: '核心指南',
          items: [
            { text: '路由系统', link: '/guides/routing' },
            { text: '中间件', link: '/guides/middleware' },
            { text: 'Context API', link: '/guides/context' },
            { text: '模板引擎', link: '/guides/templates' },
            { text: '配置管理', link: '/guides/configuration' },
          ],
        },
      ],
      '/components/': [
        {
          text: '内置组件',
          items: [
            { text: '缓存系统', link: '/components/cache' },
            { text: '会话管理', link: '/components/session' },
            { text: 'JWT认证', link: '/components/jwt' },
          ],
        },
      ],
      '/middleware/': [
        {
          text: '中间件',
          items: [
            { text: '中间件概览', link: '/middleware/overview' },
            { text: '认证中间件', link: '/middleware/authentication' },
            { text: '安全中间件', link: '/middleware/security' },
          ],
        },
      ],
      '/advanced/': [
        {
          text: '高级功能',
          items: [
            { text: 'WebSocket', link: '/advanced/websocket' },
            { text: 'JSON-RPC', link: '/advanced/jsonrpc' },
            { text: '代理功能', link: '/advanced/proxy' },
            { text: '定时任务', link: '/advanced/cron-jobs' },
          ],
        },
      ],
      '/api-reference/': [
        {
          text: 'API 参考',
          items: [
            { text: 'Application', link: '/api-reference/application' },
            { text: 'Context', link: '/api-reference/context' },
            { text: 'Router', link: '/api-reference/router' },
            { text: '中间件列表', link: '/api-reference/middleware-list' },
          ],
        },
      ],
      '/examples/': [
        {
          text: '示例项目',
          items: [
            { text: 'RESTful API', link: '/examples/rest-api' },
            { text: 'WebSocket 应用', link: '/examples/real-time-app' },
            { text: '静态文件服务', link: '/examples/static-files' },
            { text: 'JSON-RPC 服务器', link: '/examples/jsonrpc-server' },
            { text: 'API Gateway', link: '/examples/api-gateway' },
            { text: '微服务架构', link: '/examples/microservice' },
          ],
        },
      ],
      '/': [
        {
          text: '文档',
          items: [
            { text: '首页', link: '/' },
            { text: '最佳实践', link: '/best-practices' },
          ],
        },
        {
          text: '快速开始',
          items: [
            { text: '安装指南', link: '/getting-started/installation' },
            { text: '5分钟快速开始', link: '/getting-started/quick-start' },
          ],
        },
        {
          text: '核心指南',
          items: [
            { text: '路由系统', link: '/guides/routing' },
            { text: '中间件', link: '/guides/middleware' },
            { text: 'Context API', link: '/guides/context' },
          ],
        },
        {
          text: 'API 参考',
          items: [
            { text: 'Application', link: '/api-reference/application' },
            { text: 'Context', link: '/api-reference/context' },
            { text: 'Router', link: '/api-reference/router' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/go-zoox/zoox' },
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2024 Zoox Team',
    },

    search: {
      provider: 'local',
    },

    editLink: {
      pattern: 'https://github.com/go-zoox/zoox/edit/master/docs/:path',
      text: '在 GitHub 上编辑此页',
    },

    lastUpdated: {
      text: '最后更新于',
      formatOptions: {
        dateStyle: 'short',
        timeStyle: 'medium',
      },
    },

    docFooter: {
      prev: '上一页',
      next: '下一页',
    },

    outline: {
      label: '页面导航',
    },

    returnToTopLabel: '返回顶部',
    sidebarMenuLabel: '菜单',
    darkModeSwitchLabel: '主题',
    lightModeSwitchTitle: '切换到浅色模式',
    darkModeSwitchTitle: '切换到深色模式',
  },

  // 多语言支持
  locales: {
    root: {
      label: '中文',
      lang: 'zh-CN',
    },
  },

  // Markdown 配置
  markdown: {
    lineNumbers: true,
    theme: {
      light: 'github-light',
      dark: 'github-dark',
    },
  },
})

import { defineConfig } from 'rspress/config';

export default defineConfig({
  root: 'docs',
  title: 'gh-repo-settings',
  description: 'Manage GitHub repository settings via YAML configuration',
  base: '/gh-repo-settings/',
  lang: 'en',
  locales: [
    {
      lang: 'en',
      label: 'English',
      title: 'gh-repo-settings',
      description: 'Manage GitHub repository settings via YAML configuration',
    },
    {
      lang: 'ja',
      label: '日本語',
      title: 'gh-repo-settings',
      description: 'YAML 設定で GitHub リポジトリ設定を管理',
    },
  ],
  themeConfig: {
    socialLinks: [
      {
        icon: 'github',
        mode: 'link',
        content: 'https://github.com/myzkey/gh-repo-settings',
      },
    ],
    nav: [
      {
        text: 'Guide',
        link: '/guide/getting-started',
      },
    ],
    sidebar: {
      '/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
            { text: 'Commands', link: '/guide/commands' },
            { text: 'Configuration', link: '/guide/configuration' },
          ],
        },
      ],
      '/ja/': [
        {
          text: 'ガイド',
          items: [
            { text: 'はじめに', link: '/ja/guide/getting-started' },
            { text: 'コマンド', link: '/ja/guide/commands' },
            { text: '設定リファレンス', link: '/ja/guide/configuration' },
          ],
        },
      ],
    },
    locales: [
      {
        lang: 'en',
        label: 'English',
        outlineTitle: 'On this page',
        prevPageText: 'Previous',
        nextPageText: 'Next',
      },
      {
        lang: 'ja',
        label: '日本語',
        outlineTitle: 'このページの内容',
        prevPageText: '前へ',
        nextPageText: '次へ',
        nav: [
          {
            text: 'ガイド',
            link: '/ja/guide/getting-started',
          },
        ],
      },
    ],
  },
});

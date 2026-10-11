import { defineConfig } from "@rspress/core";
import { pluginSitemap } from "@rspress/plugin-sitemap";
import path from "node:path";
import pluginKatex from "rspress-plugin-katex";

export default defineConfig({
  root: "content",
  title: "regi",
  description: "A package manager for declarative entities.",
  lang: "en",
  logoText: "regi",
  base: "/regi/",
  siteOrigin: "https://ju4n97.github.io",
  globalStyles: path.join(__dirname, "theme/index.css"),
  head: [["meta", { name: "theme-color", content: "#d97706" }]],
  route: {
    cleanUrls: true,
  },
  llms: true,
  markdown: {
    link: { checkDeadLinks: true, checkAnchors: true },
    image: { checkDeadImages: true },
  },
  themeConfig: {
    darkMode: "auto",
    enableContentAnimation: false,
    enableAppearanceAnimation: false,
    lastUpdated: true,
    enableScrollToTop: true,
    llmsUI: true,
    editLink: {
      docRepoBaseUrl: "https://github.com/ju4n97/regi/tree/main/website/docs",
    },
    socialLinks: [
      {
        icon: "github",
        mode: "link",
        content: "https://github.com/ju4n97/regi",
      },
    ],
  },
  plugins: [
    pluginSitemap(),
    pluginKatex(),
  ],
});

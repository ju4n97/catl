import { defineConfig } from "@rspress/core";
import { pluginSitemap } from "@rspress/plugin-sitemap";
import {
  transformerNotationDiff,
  transformerNotationErrorLevel,
  transformerNotationHighlight,
} from "@shikijs/transformers";
import path from "node:path";
import pluginKatex from "rspress-plugin-katex";
import pluginMermaid from "rspress-plugin-mermaid";

export default defineConfig({
  root: "content",
  title: "catl",
  description: "A package manager for declarative entities.",
  lang: "en",
  logoText: "catl",
  base: "/catl/",
  outDir: "dist",
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
    shiki: {
      transformers: [
        transformerNotationHighlight(),
        transformerNotationDiff(),
        transformerNotationErrorLevel(),
      ],
    },
  },
  themeConfig: {
    darkMode: "auto",
    enableContentAnimation: false,
    enableAppearanceAnimation: false,
    lastUpdated: true,
    enableScrollToTop: true,
    llmsUI: true,
    editLink: {
      docRepoBaseUrl: "https://github.com/ju4n97/catl/tree/main/website/docs",
    },
    socialLinks: [
      {
        icon: "github",
        mode: "link",
        content: "https://github.com/ju4n97/catl",
      },
    ],
  },
  plugins: [
    pluginSitemap(),
    pluginKatex(),
    pluginMermaid(),
  ],
});

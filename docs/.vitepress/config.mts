import { type DefaultTheme, defineConfig, type UserConfig } from "vitepress";
import { withSidebar } from "vitepress-sidebar";
import type { VitePressSidebarOptions } from "vitepress-sidebar/types";

const config: UserConfig<DefaultTheme.Config> = {
  title: "低レベルコンテナランタイム自作講座",
  description: "コンテナ技術の地盤を理解する",
  head: [
    [
      "link",
      {
        rel: "icon",
        type: "image/png",
        href: "/favicon-96x96.png",
        sizes: "96x96",
      },
    ],
    ["link", { rel: "icon", type: "image/svg+xml", href: "/favicon.svg" }],
    [
      "link",
      {
        rel: "apple-touch-icon",
        sizes: "180x180",
        href: "/apple-touch-icon.png",
      },
    ],
    ["link", { rel: "shortcut icon", href: "/favicon.ico" }],
    ["link", { rel: "manifest", href: "/site.webmanifest" }],
    [
      "meta",
      {
        property: "og:title",
        content: "低レベルコンテナランタイム自作講座",
      },
    ],
    [
      "meta",
      {
        property: "og:description",
        content: "コンテナ技術の地盤を理解する",
      },
    ],
    [
      "meta",
      {
        property: "og:url",
        content: "https://coding-container-runtime.logica0419.dev",
      },
    ],
    [
      "meta",
      {
        property: "og:image",
        content: "https://coding-container-runtime.logica0419.dev/image.png",
      },
    ],
    [
      "script",
      { type: "application/ld+json" },
      `{
        "@context" : "https://schema.org",
        "@type" : "WebSite",
        "name" : "低レベルコンテナランタイム自作講座",
        "url" : "https://coding-container-runtime.logica0419.dev/"
      }`,
    ],
  ],
  srcDir: ".",
  lastUpdated: true,
  sitemap: {
    hostname: "https://coding-container-runtime.logica0419.dev",
    lastmodDateOnly: false,
  },
  themeConfig: {
    nav: [{ text: "Home", link: "/" }],
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/logica0419/coding-container-runtime",
      },
    ],
  },
};

const sidebarConfigs: VitePressSidebarOptions = {
  documentRootPath: "/",
  collapsed: false,
  useTitleFromFileHeading: true,
  useFolderTitleFromIndexFile: true,
  useFolderLinkFromIndexFile: true,
};

export default defineConfig(withSidebar(config, sidebarConfigs));

---
version: alpha
name: "Telegram"
website: "https://telegram.org"
description: >-
  A messaging app whose marketing site looks like it was last redesigned in 2014 and never touched again — pure white canvas, a single saturated cyan (#0088cc) carrying every link, heading, and the paper-airplane wordmark, and the entire system rendered in the platform Lucida Grande system stack with zero web fonts. Below the fold a 3-by-3 grid of cartoon mascot ducks labels nine product principles (Simple, Private, Synced, Fast, Powerful, Open, Secure, Social, Expressive) — each label sits in the same single cyan at 26px / weight 400 / -1px tracking, with one short sentence underneath in graphite text. The chrome is so sparse it reads almost as a wiki: no rounded buttons, no shadows, no hover states declared anywhere on the page.

seo:
  title: "Telegram Design System for React — cyan #0088cc, Lucida Grande, 14 components"
  metaDescription: "Telegram's marketing system as a DESIGN.md file. One cyan voltage, the Lucida Grande system stack, no rounded chrome, no shadow tier. Tokens for React, Next.js, and AI coding tools."
  highlights:
    - "Single cyan voltage — #0088cc covers every link, heading, brand mark, and product-principle label, never reserved for one CTA the way most SaaS systems hold accents"
    - "Lucida Grande system stack — no web font, no Inter, no Helvetica Neue Display; the page renders in whatever the OS ships, which keeps page-weight near-zero"
    - "Anti-design as positioning — no rounded chrome, no shadow tier, no hover states declared; the visual restraint is a privacy posture against the Meta family"
    - "26px / weight 400 / -1px tracking display — every principle label uses the same tight setting, making the 3-by-3 mascot grid scan as one editorial unit"
    - "Cartoon-duck mascots as iconography — Telegram's signature illustration style replaces the gradient-icon convention every messaging peer leans on"
  tags:
    - "Communication & Messaging"
  lastUpdated: "2026-05-18"
  author:
    name: "Dov Azencot"
    url: "https://x.com/dovazencot"
  opening: |
    Telegram's marketing site is the loudest argument against the past decade of brand-system inflation. The page is pure white. The paper-airplane mark sits at the top in cyan. The wordmark "Telegram" sits in cyan at 26px, the dek "a new era of messaging" in graphite below at 20px Helvetica Neue Light, and that is the entire above-fold composition — no hero illustration, no gradient, no shadow, no rounded CTA pill. Below the fold the page lists nine product principles in a 3-by-3 grid of cartoon ducks, each label rendered in the same single cyan, each sentence in the same single graphite. Where WhatsApp paints its hero in jade green and pushes a download-app CTA, where Signal leans on an indigo gradient and animated illustrations, Telegram refuses every convention of the category — there is no CTA above the fold, no app-store badge in the hero, no scroll-driven motion, no captured email field. The restraint is the position: a Dubai-based messenger that has shed Western design tropes alongside Western advertising tracking.

    The DESIGN.md file packages the system into a machine-readable spec for React tooling. Inside: 5 color tokens — a single cyan voltage, four neutral inks — drawn from the seven colors the page renders at all; 7 typography tokens running the platform Lucida Grande stack (with one Helvetica Neue Light moment for the dek); zero declared border-radius tokens because the captured surface has no rounded chrome (the input field on the search overlay rounds to 0px, the news cards have no radius, the icons are circular only because they are SVGs); 8 spacing values on a 4px base; and 14 component definitions covering the cyan link, the cyan heading, the graphite body, the cartoon-duck principle tile, the news card, and the footer columns.

    Feed this file to an AI coding tool and it reproduces Telegram's specific moves: a single saturated cyan instead of a hero-friendly indigo, the system Lucida Grande stack instead of Inter or Geist, zero corner-radius tokens, zero shadow tier, and a 3-by-3 product-principle grid where the labels share the same color and weight rather than tiering visually. Borrow the move only if your positioning is willing to read as anti-design — Telegram gets away with it because the brand's central promise is the absence of the chrome (no algorithmic feed, no ad inventory, no engagement-metric pressure). Most teams need at least one rounded surface to look "real" in 2026.
  related:
    - href: "/design"
      title: "Browse all design systems"
      description: "The full directory of DESIGN.md files on shadcn.io, with live mockups for each."
    - href: "https://telegram.org"
      title: "Telegram — official site"
      description: "Telegram's public marketing site — the source of truth for the live tokens captured in this file."
    - href: "https://github.com/google-labs-code/design.md"
      title: "The DESIGN.md specification"
      description: "Google Labs' open spec for machine-readable design system files — the format this page is built on."
  questions:
    - id: "primary-color"
      title: "What is Telegram's primary brand color?"
      answer: "Telegram's brand voltage is a single saturated cyan-teal — #0088cc — and it does the work of every chromatic element on the marketing surface. It carries the paper-airplane logomark, the Telegram wordmark, every heading from h3 down to small caps, every link in the body, the download-platform labels under the device mockups, and the nine product-principle labels under the cartoon ducks. The page captures 128 total uses of this single hex — 64 as text, 64 as border — versus zero declared brand secondary. Where most messaging apps reserve their brand color for a CTA pill, Telegram uses cyan as the universal voice of every interactive surface."
    - id: "typography"
      title: "What typeface does Telegram use, and what should I use as a substitute?"
      answer: "Telegram runs the platform Lucida Grande system stack — fallback order Lucida Grande, Lucida Sans Unicode, Arial, Helvetica, Verdana — for every body, label, heading, and nav element. There is one second-family moment on the page: the hero dek 'a new era of messaging' renders in Helvetica Neue Light at 20px / weight 300. Display headings sit at 26px / weight 400 with -1px tracking; body at 15px / weight 400; small labels at 12px / 400; bold labels at 15px / 700. The closest open-source substitute is the system font stack itself; if you need a single web font, Verdana ships everywhere and matches the Lucida proportions at body sizes. The system avoids any web-font request entirely, which is part of the privacy posture — no fonts.googleapis.com call."
    - id: "no-chrome"
      title: "Why does Telegram's site look so deliberately undesigned?"
      answer: "Telegram's marketing surface ships zero declared border-radius tokens, zero shadow tokens, and zero hover-state declarations on the captured page. The visual restraint is positioning: Telegram is a Dubai-based messenger that markets itself in opposition to the Meta family (WhatsApp, Messenger, Instagram DMs) on grounds of privacy, group-size limits, and refusal of the advertising business model. The plain-text wiki aesthetic is the brand promise made visible — no engagement-optimization design language, no algorithmic-feed icon vocabulary, no growth-team CTA stack. The choice survives despite every UX-modernization argument because changing it would undercut the position."
    - id: "principle-grid"
      title: "What is the 3-by-3 product-principle grid below the fold?"
      answer: "Telegram's below-fold composition is a nine-cell grid labeled 'Why Telegram?' with one cartoon-duck mascot per cell and one principle word underneath: Simple, Private, Synced, Fast, Powerful, Open, Secure, Social, Expressive. Every label renders in the same cyan #0088cc at 26px / weight 400 / -1px tracking — there is no visual hierarchy between principles. The mascots are part of the Telegram illustration library (the same duck appears in the in-app sticker pack) and replace the gradient icons every messaging peer leans on. The grid is the signature marketing composition; on smaller breakpoints it collapses to 2-up but the label treatment never changes."
    - id: "use-in-project"
      title: "Can I use this DESIGN.md to build my own messaging-app marketing site?"
      answer: "Yes — the file is designed to be fed into Claude, Cursor, or any AI tool that reads structured design tokens. The agent will reproduce Telegram's specific moves: single cyan voltage carrying every chromatic surface, system Lucida Grande stack instead of a web font, zero corner-radius tokens, zero shadow tier, and a flat 3-by-3 product-principle grid where the labels share the same color and weight. The tokens reference resolve cleanly — `{colors.primary}` for the cyan, `{typography.display-md}` for the cyan headings, `{typography.body-md}` for the graphite body. One caveat: the anti-design move only works if your brand promise is the absence of chrome. Most messaging-app marketing pages will feel hollow without at least one rounded CTA pill."

mockups:
  - "marketing-hero"
  - "link-in-bio-profile"

colors:
  primary: "#0088cc"
  ink: "#333333"
  ink-soft: "#222222"
  ink-muted: "#888888"
  ink-faint: "#a2a2a2"
  canvas: "#ffffff"
  shadow: "#000000"

typography:
  display-md:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 26px
    fontWeight: 400
    lineHeight: 28.6px
    letterSpacing: "-1px"
  display-sm:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 23px
    fontWeight: 500
    lineHeight: 25.3px
    letterSpacing: 0
  hero-dek:
    fontFamily: "HelveticaNeue-Light, \"Helvetica Neue Light\", \"Helvetica Light\", Helvetica, Arial, Verdana, sans-serif"
    fontSize: 20px
    fontWeight: 300
    lineHeight: 29.6px
    letterSpacing: 0
  heading-md:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 16px
    fontWeight: 700
    lineHeight: 25.6px
    letterSpacing: 0
  body-md:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 15px
    fontWeight: 400
    lineHeight: 23.7px
    letterSpacing: 0
  body-md-bold:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 15px
    fontWeight: 700
    lineHeight: 23.7px
    letterSpacing: 0
  body-sm:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 23px
    letterSpacing: 0
  label-sm:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 14px
    fontWeight: 700
    lineHeight: 15.4px
    letterSpacing: 0
  caption:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 12px
    fontWeight: 400
    lineHeight: 18px
    letterSpacing: 0
  caption-bold:
    fontFamily: "\"Lucida Grande\", \"Lucida Sans Unicode\", Arial, Helvetica, Verdana, sans-serif"
    fontSize: 12px
    fontWeight: 700
    lineHeight: 18px
    letterSpacing: 0

rounded:
  none: "0px"
  sm: "2px"
  md: "6px"
  lg: "9px"

spacing:
  xs: "2px"
  sm: "6px"
  md: "9px"
  base: "10px"
  lg: "15px"
  xl: "17px"
  2xl: "20px"
  3xl: "40px"

components:
  button-primary:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.primary}"
    typography: "{typography.body-md}"
    rounded: "{rounded.none}"
    padding: "0px"
    borderColor: "{colors.primary}"
  button-secondary:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.body-md}"
    rounded: "{rounded.none}"
    padding: "0px"
  top-nav:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.primary}"
    typography: "{typography.body-sm}"
    rounded: "{rounded.none}"
    padding: "15.5px 15px"
  nav-link:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.body-sm}"
    padding: "0px 15px"
  hero-heading:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.display-md}"
    padding: "20px 0px 9px"
  hero-dek:
    backgroundColor: "transparent"
    textColor: "{colors.ink-muted}"
    typography: "{typography.hero-dek}"
    padding: "1px 0px 10px"
  section-heading:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.display-sm}"
    padding: "20px 0px 9px"
  principle-label:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.display-md}"
    padding: "9px 0px"
  body-paragraph:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.body-md}"
  body-paragraph-muted:
    backgroundColor: "transparent"
    textColor: "{colors.ink-muted}"
    typography: "{typography.body-md}"
  link-cyan:
    backgroundColor: "transparent"
    textColor: "{colors.primary}"
    typography: "{typography.body-md}"
  news-card:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.body-md}"
    rounded: "{rounded.none}"
    padding: "0px"
  text-input:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.body-md}"
    rounded: "{rounded.none}"
    padding: "6px"
    borderColor: "{colors.ink-faint}"
  footer:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.ink}"
    typography: "{typography.caption}"
    padding: "40px 15px"
---

> **本项目不再使用这份系统。** 这里描述的是 telegram.org **营销官网**（直角、Lucida Grande、零圆角、零阴影），
> 与 Telegram 客户端的观感无关——见文末 Known Gaps 的最后一条。`web/` 的控制台改为对齐 **Telegram for macOS 客户端**：
> 圆角 10-14px、SF/系统字体栈、蓝色选中行、渐变头像、涂鸦聊天壁纸、带尾巴的气泡。令牌在 `web/src/index.css`。

## Overview

Telegram's marketing site is the loudest argument against the past decade of brand-system inflation. **Anti-design as posture.** The page is pure white. The paper-airplane mark sits at the top in cyan. The wordmark "Telegram" sits in cyan at 26px, the dek "a new era of messaging" in graphite below at 20px Helvetica Neue Light, and that is the entire above-fold composition — no hero illustration, no gradient, no shadow, no rounded CTA pill. Where WhatsApp paints its hero in jade green and pushes a download-app CTA above the fold, and where Signal leans on an indigo gradient with animated illustrations, Telegram refuses every convention of the messaging-app category. There is no captured email field, no scroll-driven motion, no app-store badge in the hero band.

The system's chromatic restraint is the central move. There is exactly one brand color — cyan (`{colors.primary}` — #0088cc) — and it does the work of every chromatic element on the page. It carries the paper-airplane logomark, the Telegram wordmark, every heading from h3 down to small caps, every link in the body, the download-platform labels under the device mockups, and the nine product-principle labels under the cartoon ducks. The page captures 128 total uses of this single hex versus zero declared brand secondary. This is the inverse of the held-in-reserve discipline at Airbnb (Rausch red used only on CTAs) or Cloudflare (Kumo orange used only as the hero canvas) — Telegram uses cyan as the universal voice of every interactive surface, including text links inside body copy. The neutral tier is a four-step graphite ladder (`{colors.ink}` graphite, `{colors.ink-soft}` near-black, `{colors.ink-muted}` mid-gray, `{colors.ink-faint}` light-gray) on a pure-white `{colors.canvas}` floor.

Typography is the platform Lucida Grande system stack for everything — no web font, no fonts.googleapis.com call, no fallback gymnastics. The one second-family moment on the page is the hero dek "a new era of messaging" in Helvetica Neue Light at 20px / weight 300; every other text element on the page resolves to Lucida Grande on macOS, Lucida Sans Unicode on Windows, or Arial as the universal fallback. Display headings sit at 26px / weight 400 with -1px tracking; body at 15px / weight 400; small captions at 12px / 400. There is no display tier above 26px on the captured page — even the "Why Telegram?" section heading sits at 23px.

**Key Characteristics:**
- Pure white canvas (`{colors.canvas}` — #ffffff), no off-white, no cream variant.
- Single cyan voltage (`{colors.primary}` — #0088cc) carrying every chromatic surface: logo, headings, links, principle labels, platform tags.
- Lucida Grande system stack for every text element — zero web font requests.
- Display headings at 26px / weight 400 with -1px tracking — the system's loudest typographic moment.
- Zero declared corner-radius tokens — buttons, inputs, news cards, and platform tags all sit at 0px rounding.
- Zero shadow tier — no drop shadow, no inner shadow, no soft-blur elevation anywhere on the page.
- 3-by-3 "Why Telegram?" principle grid (`Simple` / `Private` / `Synced` / `Fast` / `Powerful` / `Open` / `Secure` / `Social` / `Expressive`) with cartoon-duck mascots replacing gradient icons.
- News-card pair below the hero — two horizontal cards with full-bleed photos, cyan headlines, and graphite dek text; no border, no shadow, no hover state.

## Colors

### Brand

- **Telegram Cyan** (`{colors.primary}` — #0088cc): frequency 128. Used as text (64) and border (64), never as background fill. The single chromatic voice of the system — logo, wordmark, every heading, every link, every principle label, every platform tag. The frequency split tells the story: Telegram does not use the cyan to fill buttons, it uses it to letterform-tint every interactive surface.

### Text

- **Ink** (`{colors.ink}` — #333333): frequency 238. Used as text (120) and border (118). The default body color and the dominant border tone for hairline rules. Pure black would feel harsh against the white canvas; this mid-graphite softens without sliding toward the warmer charcoal grays peers use.
- **Ink Soft** (`{colors.ink-soft}` — #222222): frequency 4. Used as text (2) and border (2). Reserved for slightly emphasized labels.
- **Ink Muted** (`{colors.ink-muted}` — #888888): frequency 6 — used on the hero dek "a new era of messaging" and on secondary labels under news cards. The extraction merges two near-identical grays into this single token.
- **Ink Faint** (`{colors.ink-faint}` — #a2a2a2): frequency 18. Used on the timestamp captions in the Recent News rail and on metadata labels.

### Surface

- **Canvas** (`{colors.canvas}` — #ffffff): frequency 2 as background. The entire page floor. No cream variant, no off-white surface tier, no charcoal-near-white dark mode declared.

### Hairline / Shadow

- **Shadow** (`{colors.shadow}` — #000000): frequency 8 — used as text (4) and border (4). Reserved for the absolute-black footer labels and a small set of hairline borders where the graphite would be too soft.

## Typography

### Font Family

The system runs the platform **Lucida Grande system stack** for every text element — fallback order `"Lucida Grande", "Lucida Sans Unicode", Arial, Helvetica, Verdana, sans-serif`. On macOS the resolved face is Lucida Grande; on Windows it is Lucida Sans Unicode; on Linux it falls through to Arial. There is no web-font request — no fonts.googleapis.com, no self-hosted .woff2 — which keeps the page-weight near-zero and avoids the third-party-tracking surface that Google Fonts introduces.

One second-family moment exists on the captured page: the hero dek "a new era of messaging" renders in **Helvetica Neue Light** at 20px / weight 300. It is the only HelveticaNeue-Light declaration in the entire system; every other surface resolves to Lucida Grande.

### Hierarchy

| Token | Size | Weight | Line Height | Use |
|---|---|---|---|---|
| `{typography.display-md}` | 26px | 400 | 28.6px | "Telegram" wordmark, every section heading, all nine principle labels |
| `{typography.display-sm}` | 23px | 500 | 25.3px | "Why Telegram?" section h3 |
| `{typography.hero-dek}` | 20px | 300 | 29.6px | "a new era of messaging" (Helvetica Neue Light) |
| `{typography.heading-md}` | 16px | 700 | 25.6px | "Recent News" sidebar h4 |
| `{typography.body-md}` | 15px | 400 | 23.7px | Default running text, principle descriptions |
| `{typography.body-md-bold}` | 15px | 700 | 23.7px | Inline emphasis, news-card headlines |
| `{typography.body-sm}` | 14px | 400 | 23px | Nav-link labels |
| `{typography.label-sm}` | 14px | 700 | 15.4px | "Telegram for Android", "Telegram for iPhone / iPad" download labels |
| `{typography.caption}` | 12px | 400 | 18px | Date stamps in the Recent News rail |
| `{typography.caption-bold}` | 12px | 700 | 18px | Recent News date headers ("May 7", "May 4", "May 1") |

### Principles

Display weight tops out at 500, used only on the "Why Telegram?" section heading. The dominant display token sits at weight 400 — the same weight as body text — and gets its prominence from size (26px) and tracking (-1px), not from boldness. There is no 700-tier moment for a hero h1 because there is no hero h1; the wordmark "Telegram" is the loudest typographic moment in the system at 26px / 400. Bold weight 700 is reserved for download platform labels, date stamps, and news-card headlines — small-format emphasis only.

### Note on Font Substitutes

Lucida Grande is a system font on macOS; Lucida Sans Unicode is its Windows equivalent. There is no open-source substitute that matches exactly, but the system stack already handles the fallback gracefully — Arial at the same weights reads close at body sizes. If you want a single web font for cross-platform parity, **Verdana** ships everywhere and matches the Lucida proportions at the 12-15px tier; **Tahoma** is the closest Windows-native equivalent. Avoid Inter or Helvetica Neue as substitutes — both feel too clearly "designed" against the deliberately-undesigned Lucida vibe.

## Layout

### Spacing System

- **Base unit:** 1px (with 3px as a common module).
- **Tokens:** `{spacing.xs}` 2px · `{spacing.sm}` 6px · `{spacing.md}` 9px · `{spacing.base}` 10px · `{spacing.lg}` 15px · `{spacing.xl}` 17px · `{spacing.2xl}` 20px · `{spacing.3xl}` 40px.
- **Section padding (vertical):** 20px above and 9px below most section headings — a tight rhythm that keeps the page scannable as a single editorial column.
- **Hero padding:** 160px above the wordmark on desktop — the only generous vertical breathing room in the system, used to push the centered Telegram mark into a visual safe zone.
- **Footer padding:** 40px outer, 15px gutter between the four column groups.

### Grid & Container

- **Max content width:** ~620px on the main editorial column — narrower than nearly every peer marketing site, which keeps the single-column reading rhythm intact.
- **Hero:** centered single-column with the paper-airplane mark, wordmark, dek, and download links stacked vertically.
- **Device-mockup band:** two-up grid showing iOS and Android device screenshots, with the "Telegram for Android" and "Telegram for iPhone / iPad" labels rendered in the platform-tag bold treatment.
- **Recent News rail:** two-card horizontal grid with full-bleed photos above cyan headlines and graphite dek.
- **"Why Telegram?" grid:** 3-by-3 desktop, collapsing to 2-up on tablet, with cartoon-duck mascots above each cyan principle label.
- **Footer:** four-column grid (About / Mobile Apps / Desktop Apps / Platform) of plain-text link lists, no visual separation between columns.

### Rhythm

The page reads as a vertical scroll of editorial bands separated by the same `{spacing.3xl}` (40px) gap. There is no atmospheric color shift between bands, no background variation, no parallax — every section terminates on the same pure-white canvas. Section transitions are marked only by a heading change from cyan body-tier to cyan display-tier.

## Elevation

The system has **no shadow tier**. Zero. The captured page declares no drop-shadow, no inner-shadow, no soft-blur elevation on any surface. News cards, principle tiles, and device mockups all sit flush on the white canvas; the only thing distinguishing one band from another is the heading color shift.

- **Flat (no shadow):** every surface on the page — 100% of the captured chrome.
- **Hairline rules:** thin 1px `{colors.ink}` rules separate the footer columns and the Recent News date headers from their content. No drop-shadow ever stands in for the rule.
- **Device mockups:** the iOS and Android phone screenshots include their own rendered drop-shadow as part of the image asset, not as a CSS shadow declaration. The shadow is photography, not chrome.

## Shapes

The system has **essentially no corner-radius tier**. The captured page renders almost every surface at 0px rounding:

- `{rounded.none}` 0px — buttons (which are inline text links, not pill chrome), news cards, principle tiles, footer columns, the hero device mockup container.
- `{rounded.sm}` 2px — declared as a fallback for any input that needs the slightest softening; rarely rendered.
- `{rounded.md}` 6px — declared as a fallback for download-link buttons on certain breakpoints; not present in the captured hero.
- `{rounded.lg}` 9px — declared as a fallback for the platform-tag pill on the device-mockup band; not present in the captured hero.

There is no pill tier, no full-circle radius, no large-card 16-24px rounding. The icons in the principle grid are circular because they are SVG illustrations of round-bodied cartoon ducks, not because any corner-radius is applied. The absence of a rounding scale is the position — the system reads as plain document chrome by design.

## Components

**`button-primary`** — There is no signature CTA pill. The closest analog is the cyan-on-white inline link, rendered in `{typography.body-md}` at `{colors.primary}` with a 1px `{colors.primary}` bottom border on hover (the border is not captured at rest). No padding, no rounded radius, no fill.

**`button-secondary`** — Identical to the primary in the captured surface: cyan text link with no fill and no rounded chrome. The system does not distinguish primary and secondary action by visual weight; the distinction is positional (download links sit in the hero, navigation links sit in the top bar).

**`top-nav`** — Plain white surface with cyan inline-text links (Home, FAQ, Apps, API, Wallpapers) and a right-aligned cyan EN-language toggle. No background fill, no border, no shadow. 15.5px vertical padding, 15px horizontal padding.

**`nav-link`** — Cyan text on transparent, `{typography.body-sm}` (14px / 400), 0px 15px padding. No hover background visible in the captured surface; the link likely picks up an underline on hover but it is not declared in the resting state.

**`hero-heading`** — The "Telegram" wordmark, cyan `{colors.primary}` text at `{typography.display-md}` (26px / 400 / -1px tracking). 20px top / 9px bottom padding. The 26px size is identical to every other display heading on the page — there is no separate hero-h1 tier.

**`hero-dek`** — "a new era of messaging" in `{colors.ink-muted}` graphite at `{typography.hero-dek}` (20px / 300 Helvetica Neue Light). The only Helvetica Neue Light moment on the page; every other text element resolves to Lucida Grande.

**`section-heading`** — Cyan `{colors.primary}` text at `{typography.display-sm}` (23px / 500) for "Why Telegram?", "Recent News". 20px top / 9px bottom padding.

**`principle-label`** — The nine cyan labels under the cartoon ducks: `Simple` / `Private` / `Synced` / `Fast` / `Powerful` / `Open` / `Secure` / `Social` / `Expressive`. Each renders in `{typography.display-md}` (26px / 400 / -1px tracking) — identical to the wordmark treatment. The repetition is the editorial move: every principle gets the same typographic weight, so the reader is not pre-sorted toward any single virtue.

**`body-paragraph`** — Default ink-graphite text at `{typography.body-md}` (15px / 400). The principle descriptions ("Telegram lets you access your chats from multiple devices.") and the news-card dek.

**`body-paragraph-muted`** — `{colors.ink-muted}` text at `{typography.body-md}` for the news-card date stamp and the secondary labels under download links.

**`link-cyan`** — Inline cyan link inside body copy, `{typography.body-md}` color `{colors.primary}`. The dominant interactive element on the page — every clickable text element inherits this treatment.

**`news-card`** — White canvas, no border, no shadow, no rounded radius. Full-bleed photo at the top, cyan headline at `{typography.body-md-bold}` (15px / 700) below, graphite dek at `{typography.body-md}`, gray date stamp at the bottom in `{typography.caption}`. The lack of visual containment is the point — news cards sit in the editorial flow like wiki entries.

**`text-input`** — White fill, ink-graphite text, `{rounded.none}` 0px radius, 6px padding, 1px `{colors.ink-faint}` border. Appears on the FAQ search overlay; not visible in the captured marketing surface.

**`footer`** — Pure white canvas, ink-graphite labels at `{typography.caption}` (12px / 400), with 12px / 700 column headers. Four columns (About / Mobile Apps / Desktop Apps / Platform). 40px outer padding, 15px gutter. No surface contrast against the page floor.

## Do's and Don'ts

**Do** treat the brand voltage as the universal voice of every interactive surface. The system uses one cyan for the logo, every heading, every link, and every principle label — there is no "primary" / "secondary" / "tertiary" hierarchy split by color. Multiplying the palette into a second brand accent would destroy the single-voice discipline.

**Do** ship the system Lucida Grande stack and skip the web font entirely. The page-weight benefit is real, but the bigger move is what it signals — no Google Fonts request means no third-party tracking surface, which aligns with the privacy posture.

**Do** keep display weight at 500 max, and use the 26px / weight 400 / -1px tracking treatment as the dominant display moment. Bumping the wordmark to 700 would turn it into a generic SaaS shout and undercut the deliberately-undesigned vibe.

**Do** render the nine "Why Telegram?" principle labels at identical typographic weight. The visual flatness across `Simple` / `Private` / `Synced` / `Fast` / `Powerful` / `Open` / `Secure` / `Social` / `Expressive` is the editorial move — no principle is pre-ranked above the others.

**Don't** introduce a rounded-pill CTA in the hero. The hero has no CTA at all in the captured surface — adding a fully-rounded download button (the WhatsApp move) would import the engagement-design language Telegram defines itself against.

**Don't** apply a drop-shadow to the news cards or principle tiles. The system has zero shadow tier — softening the cards with even a 2-4px shadow would push the page toward Material-style elevation and away from the wiki-document aesthetic.

**Don't** swap Lucida Grande for Inter or Helvetica Neue as a "type-update" pass. Both fonts read as clearly-designed type choices; Lucida reads as a system default. The "didn't choose a font" affect is part of the position.

**Don't** rework the principle grid into a card stack with backgrounds and rounded corners (the post-2018 SaaS feature-grid convention). The flat 3-by-3 mascot grid with no card containment is Telegram's signature marketing composition — adding containment would make it look like every other startup feature section.

## Known Gaps

- **Hover and focus states:** the captured page declares no hover background, no focus ring, no underline-on-hover state for the cyan links. The product app uses an underline; the marketing site does not.
- **Mobile breakpoint behavior:** captured at desktop only. The 3-by-3 principle grid collapses to 2-up on tablet and 1-up on phone, but the breakpoint values are not exposed in the marketing CSS.
- **Dark mode:** no dark variant on the captured marketing surface. The product app ships dark, but the website is light-only.
- **Motion:** the page has zero captured animation — no scroll-driven motion, no entrance fade, no hover transition. The product app uses Lottie animations for the duck mascots, but the marketing site renders them as static SVGs.
- **Search overlay:** the FAQ search lives behind a click and is not exposed in the captured first-paint surface.
- **Product surfaces:** this DESIGN.md captures the telegram.org marketing site only. The Telegram Web app (web.telegram.org) and the desktop / mobile clients carry a much richer token system — chat bubbles, sticker-pack chrome, in-app sidebar, voice-call surface — that is not represented here.
- **Cyan brand variants:** the system declares no hover-cyan or pressed-cyan token. The single #0088cc is used at all states; any state variation lives in opacity or underline rather than a separate color value.

/**
 * Centralized theme configuration for RSLBot
 * Based on original rslbot-infra landing page styling
 */

export const theme = {
  colors: {
    // Primary brand colors
    primary: {
      dark: "#322B6D", // Main brand color (headings, emphasis)
      purple: "#352C71", // Feature card gradient start
      violet: "#743AA0", // Feature card gradient end
      magenta: "#71389E", // Pricing card gradient start
      pink: "#B96FC8", // Pricing card gradient end
    },

    // Background colors
    background: {
      white: "#FFFFFF",
      light: "#F9FAFB",
    },

    // Text colors
    text: {
      white: "#FFFFFF",
      dark: "#1F2937",
      gray: "#6B7280",
    },

    // UI states
    ui: {
      hover: "#DCDCDC",
      border: "#E5E7EB",
    },
  },

  gradients: {
    // Feature cards gradient
    feature: "linear-gradient(180deg, #352C71 0%, #743AA0 100%)",

    // Pricing cards gradient
    pricing: "linear-gradient(180deg, #71389E 0%, #9553B3 49.65%, #B96FC8 100%)",
  },

  shadows: {
    card: "2px 4px 15px rgba(0, 0, 0, 0.25)",
    sm: "0 1px 2px 0 rgb(0 0 0 / 0.05)",
    md: "0 4px 6px -1px rgb(0 0 0 / 0.1)",
    lg: "0 10px 15px -3px rgb(0 0 0 / 0.1)",
  },

  borderRadius: {
    sm: "0.375rem", // 6px
    md: "0.5rem", // 8px
    lg: "0.75rem", // 12px
    xl: "0.9375rem", // 15px
    "2xl": "1rem", // 16px
    full: "9999px",
  },

  spacing: {
    cardMinHeight: "450px",
    cardMaxWidth: "315px",
    pricingMinHeight: "400px",
  },
} as const

export type Theme = typeof theme

// Discord invite link
export const DISCORD_INVITE = "https://discord.gg/jy8eDQjCt6"

// External links
export const links = {
  discord: DISCORD_INVITE,
  forum: "https://community.rslbot.com",
  releases: "https://community.rslbot.com/c/eb2-releases/8",
} as const

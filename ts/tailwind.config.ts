import type { Config } from "tailwindcss"
import daisyui from "daisyui"
import { fontVariationSettings } from "./plugins/fontVariationPlugin"

export default {
  content: ["./app/**/{**,.client,.server}/**/*.{js,jsx,ts,tsx}"],
  plugins: [daisyui, fontVariationSettings],
  theme: {
    extend: {
      backgroundImage: {
        "divider-vertical":
          "linear-gradient(180deg, rgba(234, 168, 121, 0) 0%, rgba(255,255,255,0.3) 100%)",
        "divider-vertical-reverse":
          "linear-gradient(0deg, rgba(234, 168, 121, 0) 0%, rgba(255,255,255,0.3) 100%)",
        "divider-horizontal":
          "linear-gradient(90deg, #eaa879 0%, rgba(201, 132, 87, 0) 0%, rgba(255,255,255,0.3) 53.60%, rgba(161, 82, 25, 0) 100%)",
      },
    },
  },
  safelist: [
    "font-normal",
    "font-semibold",
    "font-bold",

    "text-3xl",
    "md:text-4xl",
    "text-xl",
    "md:text-2xl",
    "text-lg",
    "text-base",

    "text-opacity-87",
    "text-opacity-60",
    "text-opacity-100",
    "text-primary",
    "text-secondary",
    "text-accent",

    "bg-neutral",
    "border-primary",
    "border-primary/20",
    "hover:border-primary/40",

    "w-7",
    "w-20",
  ],
  daisyui: {
    themes: [
      {
        raidbot2: {
          // Based on original rslbot-infra colors
          primary: "#743AA0", // Purple from feature cards
          "primary-content": "#FFFFFF",
          secondary: "#9553B3", // Mid-tone from pricing gradient
          "secondary-content": "#FFFFFF",
          accent: "#B96FC8", // Pink from pricing cards
          "accent-content": "#FFFFFF",
          neutral: "#322B6D", // Dark brand color
          "neutral-content": "#FFFFFF",
          "base-100": "#FFFFFF", // White background
          "base-200": "#F9FAFB", // Light gray
          "base-300": "#E5E7EB", // Border color
          "base-content": "#1F2937", // Dark text
          info: "#352C71", // Feature card purple
          "info-content": "#FFFFFF",
          success: "#10B981",
          "success-content": "#FFFFFF",
          warning: "#F59E0B",
          "warning-content": "#FFFFFF",
          error: "#EF4444",
          "error-content": "#FFFFFF",
        },
      },
    ],
    darkTheme: "raidbot2",
    base: true,
    styled: true,
    utils: true,
    prefix: "",
    logs: true,
    themeRoot: ":root",
  },
} satisfies Config

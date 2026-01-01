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
        light: {
          // Light theme based on original rslbot-infra colors
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
      {
        dark: {
          // Dark theme variant
          primary: "#9553B3", // Lighter purple for dark mode
          "primary-content": "#FFFFFF",
          secondary: "#B96FC8", // Pink accent
          "secondary-content": "#FFFFFF",
          accent: "#C084FC", // Brighter purple accent
          "accent-content": "#FFFFFF",
          neutral: "#1E1B4B", // Darker navy
          "neutral-content": "#E5E7EB",
          "base-100": "#0F0E1A", // Very dark background
          "base-200": "#1A1825", // Slightly lighter
          "base-300": "#2D2A3E", // Border color for dark
          "base-content": "#E5E7EB", // Light text
          info: "#8B5CF6", // Brighter purple for info
          "info-content": "#FFFFFF",
          success: "#34D399",
          "success-content": "#FFFFFF",
          warning: "#FBBF24",
          "warning-content": "#1F2937",
          error: "#F87171",
          "error-content": "#FFFFFF",
        },
      },
    ],
    darkTheme: "dark",
    base: true,
    styled: true,
    utils: true,
    prefix: "",
    logs: true,
    themeRoot: ":root",
  },
} satisfies Config

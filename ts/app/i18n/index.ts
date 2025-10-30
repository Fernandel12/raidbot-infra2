import { createCookie } from "@remix-run/cloudflare"

export const supportedLanguages = ["en", "pt-BR", "tw", "ru", "ko"]
export const defaultLanguage = "en"

// Create a cookie to store the user's language preference
export const i18nCookie = createCookie("i18n", {
  path: "/",
  httpOnly: true,
  secure: process.env.NODE_ENV === "production",
  sameSite: "lax",
  maxAge: 31536000, // 1 year
})

// Re-export the improved loader from root-loader
export { i18nLoader } from "./root-loader"

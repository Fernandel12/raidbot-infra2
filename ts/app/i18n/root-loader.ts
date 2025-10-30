import type { LoaderFunctionArgs } from "@remix-run/cloudflare"
import { defaultLanguage, supportedLanguages } from "."
import { i18nCookie } from "./cookie"

export async function getLocale(request: Request) {
  // First check URL query parameter
  const url = new URL(request.url)
  const urlLang = url.searchParams.get("lang")
  if (urlLang && supportedLanguages.includes(urlLang)) {
    return {
      locale: urlLang,
      setCookie: true, // Flag to indicate cookie should be set
    }
  }

  // Get locale from cookie
  const cookieHeader = request.headers.get("Cookie")
  const cookieValue: unknown = await i18nCookie.parse(cookieHeader)
  const locale = typeof cookieValue === "string" ? cookieValue : null

  // Check if cookie locale is supported
  if (locale && supportedLanguages.includes(locale)) {
    return { locale, setCookie: false }
  }

  // Try to get locale from Accept-Language header
  const acceptLanguage = request.headers.get("Accept-Language")
  if (acceptLanguage) {
    // Parse Accept-Language header
    const langCodes = acceptLanguage.split(",").map((lang) => lang.split(";")[0].trim())

    // Direct match for our codes
    for (const code of langCodes) {
      // First try exact match
      if (supportedLanguages.includes(code)) {
        return { locale: code, setCookie: false }
      }

      // Handle Chinese variants (zh-TW would match to our 'tw' code)
      if (code === "zh-TW" && supportedLanguages.includes("tw")) {
        return { locale: "tw", setCookie: false }
      }

      // Handle Brazilian Portuguese
      if ((code === "pt-BR" || code === "pt") && supportedLanguages.includes("pt-BR")) {
        return { locale: "pt-BR", setCookie: false }
      }

      // Handle general language code (first 2 chars)
      const shortCode = code.substring(0, 2)
      if (supportedLanguages.includes(shortCode)) {
        return { locale: shortCode, setCookie: false }
      }
    }
  }

  // Default to English
  return { locale: defaultLanguage, setCookie: false }
}

export async function i18nLoader({ request }: LoaderFunctionArgs) {
  const result = await getLocale(request)

  // If we need to set a cookie, create a proper response with headers
  if (result.setCookie) {
    const cookieValue = await i18nCookie.serialize(result.locale)
    return {
      locale: result.locale,
      headers: {
        "Set-Cookie": cookieValue,
      },
    }
  }

  return { locale: result.locale }
}

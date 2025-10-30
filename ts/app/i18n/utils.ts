import i18n from "~/components/I18nProvider"
import { defaultLanguage } from "~/i18n"

import translationsEN from "./locales/en/translations.json"
import translationsRU from "./locales/ru/translations.json"
import translationsTW from "./locales/tw/translations.json"
import translationsKO from "./locales/ko/translations.json"
import translationsPTBR from "./locales/pt-BR/translations.json"

// Static resources map for direct access
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const staticResources: Record<string, any> = {
  en: translationsEN,
  ru: translationsRU,
  tw: translationsTW,
  ko: translationsKO,
  "pt-BR": translationsPTBR,
}

/**
 * Get a translation value without using React hooks
 * This can be used in non-React contexts like meta functions
 */
export function getTranslation(key: string, locale?: string) {
  // Try to get current locale from i18n, fallback to default
  const currentLocale = locale || (i18n.isInitialized ? i18n.language : defaultLanguage)

  try {
    // First try to get it from the initialized i18n instance
    if (i18n.isInitialized) {
      const value = i18n.t(key, { lng: currentLocale })
      // If we got something that's not the key itself, return it
      if (value !== key) {
        return value
      }
    }

    // Fallback to direct access from static resources
    const parts = key.split(".")
    // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment
    let result = staticResources?.[currentLocale] || staticResources?.[defaultLanguage]

    // Navigate through the nested structure
    for (const part of parts) {
      if (result && typeof result === "object" && part in result) {
        // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access
        result = result[part]
      } else {
        // If key not found, try with default language
        if (currentLocale !== defaultLanguage) {
          return getTranslation(key, defaultLanguage)
        }
        return key // Last fallback - return the key itself
      }
    }

    return typeof result === "string" ? result : key
  } catch (error) {
    console.error(`[I18N] Error getting translation for ${currentLocale}.${key}:`, error)
    return key
  }
}

/**
 * Helper function to get a translation function with a specific locale
 */
export function getTranslator(locale?: string) {
  const currentLocale = locale || (i18n.isInitialized ? i18n.language : defaultLanguage)

  return (key: string) => getTranslation(key, currentLocale)
}

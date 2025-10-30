import { defaultLanguage, supportedLanguages } from "./index"

type TranslationValue = string | string[] | TranslationNamespace | TranslationNamespace[]

type TranslationNamespace = {
  [key: string]: TranslationValue
}

import translationsEN from "./locales/en/translations.json"
import translationsRU from "./locales/ru/translations.json"
import translationsTW from "./locales/tw/translations.json"
import translationsKO from "./locales/ko/translations.json"
import translationsPTBR from "./locales/pt-BR/translations.json"

// Define the translations structure
const translations: Record<string, TranslationNamespace> = {
  en: translationsEN,
  "pt-BR": translationsPTBR,
  ru: translationsRU,
  tw: translationsTW,
  ko: translationsKO,
}

// Function to get translations for a specific locale
export function getTranslation(locale: string): TranslationNamespace {
  // Validate locale
  const validLocale = supportedLanguages.includes(locale) ? locale : defaultLanguage

  // Return translations for this locale
  if (translations[validLocale]) {
    return translations[validLocale]
  }

  // Fallback to English
  return translations[defaultLanguage] || {}
}

// Function to translate a specific key
export function translate(locale: string, key: string, params?: Record<string, string>): string {
  try {
    const translation = getTranslation(locale)

    // Split the key by dots to handle nested objects
    const parts = key.split(".")
    let result: TranslationValue | undefined = translation

    // Navigate through the nested structure
    for (const part of parts) {
      if (result && typeof result === "object" && part in result) {
        result = (result as Record<string, TranslationValue>)[part]
      } else {
        // Key not found, fallback to English
        if (locale !== defaultLanguage) {
          return translate(defaultLanguage, key, params)
        }
        // If already in English and key not found, return the key itself
        return key
      }
    }

    // If the result is not a string, return the key
    if (typeof result !== "string") {
      return key
    }

    // Replace parameters if any
    if (params) {
      return Object.entries(params).reduce((str, [paramKey, paramValue]) => {
        return str.replace(new RegExp(`{{${paramKey}}}`, "g"), paramValue)
      }, result)
    }

    return result
  } catch (error) {
    console.error("Translation error:", error)
    return key
  }
}

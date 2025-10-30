import { useLocation, useRevalidator } from "@remix-run/react"
import i18next from "i18next"
import LanguageDetector from "i18next-browser-languagedetector"
import { createContext, useEffect, useRef, useState } from "react"
import { I18nextProvider, initReactI18next } from "react-i18next"
import { defaultLanguage, supportedLanguages } from "~/i18n"

import translationsEN from "~/i18n/locales/en/translations.json"
import translationsRU from "~/i18n/locales/ru/translations.json"
import translationsTW from "~/i18n/locales/tw/translations.json"
import translationsKO from "~/i18n/locales/ko/translations.json"
import translationsPTBR from "~/i18n/locales/pt-BR/translations.json"

// Initialize i18next instance for client
const i18n = i18next.createInstance()

// Initialize with translations for all supported languages
const resources = {
  en: {
    translation: translationsEN,
  },
  "pt-BR": {
    translation: translationsPTBR,
  },
  ru: {
    translation: translationsRU,
  },
  tw: {
    translation: translationsTW,
  },
  ko: {
    translation: translationsKO,
  },
}

// Track initialization state
let isInitialized = false

// Initialize i18next with safe configuration
function initializeI18n() {
  if (isInitialized) return Promise.resolve()

  return i18n
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
      fallbackLng: defaultLanguage,
      supportedLngs: supportedLanguages,
      interpolation: {
        escapeValue: false,
      },
      detection: {
        order: ["htmlTag", "cookie", "querystring"],
        lookupCookie: "i18n",
        lookupQuerystring: "lang", // Add this to detect from query params
        caches: ["cookie"],
      },
      resources,
    })
    .then(() => {
      isInitialized = true
    })
    .catch((error) => {
      console.error("[I18N] Error initializing i18next:", error)
      // Continue anyway to avoid blocking the UI
      isInitialized = true
    })
}

// Create a context to expose the changeLanguage function
export const I18nContext = createContext<{
  i18n: typeof i18next
  changeLanguage: (locale: string) => Promise<boolean>
  locale: string
  isReady: boolean
}>({
  i18n,
  changeLanguage: () => Promise.resolve(false),
  locale: defaultLanguage,
  isReady: false,
})

interface I18nProviderProps {
  children: React.ReactNode
  locale: string
}

export function I18nProvider({ children, locale }: I18nProviderProps) {
  const [isReady, setIsReady] = useState(false)
  const location = useLocation()
  const [initializationError, setInitializationError] = useState<string | null>(null)
  const revalidator = useRevalidator()
  const readyRef = useRef(false)

  // Check URL parameters directly to ensure consistency with server
  useEffect(() => {
    if (typeof window !== "undefined") {
      const url = new URL(window.location.href)
      const urlLang = url.searchParams.get("lang")

      if (
        urlLang &&
        supportedLanguages.includes(urlLang) &&
        isInitialized &&
        i18n.language !== urlLang
      ) {
        // If URL has language param, manually change it to ensure consistency
        i18n
          .changeLanguage(urlLang)
          .catch((e) => console.error("[I18N] URL param language change error:", e))
      }
    }
  }, [location.search, isReady])

  // Initialize i18next on first render
  useEffect(() => {
    const setup = async () => {
      try {
        await initializeI18n()

        // Set the language - prioritize URL param if it exists
        let langToUse = locale
        if (typeof window !== "undefined") {
          const url = new URL(window.location.href)
          const urlLang = url.searchParams.get("lang")
          if (urlLang && supportedLanguages.includes(urlLang)) {
            langToUse = urlLang
          }
        }

        if (i18n.language !== langToUse) {
          await i18n.changeLanguage(langToUse)
        }

        // Update both state and ref
        readyRef.current = true
        setIsReady(true)
      } catch (error) {
        console.error("[I18N] Setup error:", error)
        setInitializationError(String(error))

        // Fallback to English
        try {
          await i18n.changeLanguage("en")
          readyRef.current = true
          setIsReady(true)
        } catch (fallbackError) {
          console.error("[I18N] Even fallback failed:", fallbackError)
          // Continue anyway to show something
          readyRef.current = true
          setIsReady(true)
        }
      }
    }

    const timeoutId = setTimeout(() => {
      if (!readyRef.current) {
        console.warn("[I18N] Setup is taking too long, forcing ready state")
        readyRef.current = true
        setIsReady(true)
      }
    }, 2000)

    void setup()

    return () => clearTimeout(timeoutId)
  }, [locale])

  // No need for route-based namespace loading since we use a single translation file

  // This function will be exposed to components that need to change language
  const changeLanguage = async (newLocale: string) => {
    try {
      // Change language in i18next
      await i18n.changeLanguage(newLocale)

      // Update document lang attribute for SEO
      const htmlLang = newLocale === "tw" ? "zh-TW" : newLocale
      document.documentElement.setAttribute("lang", htmlLang)

      // Trigger revalidation of the current route
      // This will re-run loaders and meta functions
      revalidator.revalidate()

      return true
    } catch (error) {
      console.error("[I18N] Language change error:", error)
      return false
    }
  }

  // Create the context value
  const contextValue = {
    i18n,
    changeLanguage,
    locale,
    isReady,
  }

  // If not ready yet, show a minimal loading indicator instead of null
  if (!isReady) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-pulse p-4 text-neutral-400">Loading...</div>
      </div>
    )
  }

  // If there was an error, we'll still render but show an error in dev mode
  if (initializationError && process.env.NODE_ENV === "development") {
    console.error("[I18N] Rendering with initialization error:", initializationError)
  }

  return (
    <I18nextProvider i18n={i18n}>
      <I18nContext.Provider value={contextValue}>{children}</I18nContext.Provider>
    </I18nextProvider>
  )
}

// Export the instance for use in components
export default i18n

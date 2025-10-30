import { Form, useSubmit } from "@remix-run/react"
import { useContext, useEffect, useRef, useState } from "react"
import { I18nContext } from "~/components/I18nProvider"
import { supportedLanguages } from "~/i18n"

// TODO: Fix accessibility - add keyboard support for language selector
/* eslint-disable jsx-a11y/click-events-have-key-events, jsx-a11y/no-static-element-interactions */

// Language names in their native language
const languageNames: Record<string, string> = {
  en: "English",
  "pt-BR": "Português",
  ru: "Русский",
  tw: "繁體中文",
  ko: "한국어",
}

// Language flags - SVG flags from CDN
// Using United States flag for English as requested
const languageFlags: Record<string, { src: string; alt: string }> = {
  en: {
    src: "https://cdnjs.cloudflare.com/ajax/libs/flag-icon-css/6.6.6/flags/4x3/gb.svg",
    alt: "British Flag",
  },
  "pt-BR": {
    src: "https://cdnjs.cloudflare.com/ajax/libs/flag-icon-css/6.6.6/flags/4x3/br.svg",
    alt: "Brazil Flag",
  },
  ru: {
    src: "https://cdnjs.cloudflare.com/ajax/libs/flag-icon-css/6.6.6/flags/4x3/ru.svg",
    alt: "Russia Flag",
  },
  // No flag for Chinese - using "中" character to avoid geopolitical issues
  tw: {
    src: "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 48 36'%3E%3Crect width='48' height='36' fill='%234a5568'/%3E%3Ctext x='24' y='26' font-family='Arial' font-size='20' font-weight='bold' fill='white' text-anchor='middle'%3E中%3C/text%3E%3C/svg%3E",
    alt: "Chinese",
  },
  ko: {
    src: "https://cdnjs.cloudflare.com/ajax/libs/flag-icon-css/6.6.6/flags/4x3/kr.svg",
    alt: "South Korea Flag",
  },
}

export default function LanguageSelector() {
  const { i18n, changeLanguage } = useContext(I18nContext)
  const [isOpen, setIsOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const submit = useSubmit()
  const formRef = useRef<HTMLFormElement>(null)

  // Handle clicking outside to close dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (ref.current && !ref.current.contains(event.target as Node)) {
        setIsOpen(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => {
      document.removeEventListener("mousedown", handleClickOutside)
    }
  }, [])

  // Handle language change
  const handleLanguageChange = async (lang: string) => {
    try {
      // First update the cookie via form submission
      if (formRef.current) {
        const input = formRef.current.elements.namedItem("lang") as HTMLInputElement
        if (input) {
          input.value = lang
          submit(formRef.current, {
            method: "post",
            action: "/set-language",
          })
        }
      }

      // Then immediately change the language client-side and revalidate
      await changeLanguage(lang)

      setIsOpen(false)
    } catch (error) {
      console.error("Error changing language:", error)
    }
  }

  // Get current language display name with fallback
  const getCurrentLanguageName = () => {
    try {
      return languageNames[i18n.language] || languageNames.en
    } catch (error) {
      console.error("Error getting language name:", error)
      return "English" // Fallback
    }
  }

  const getCurrentLanguageFlag = () => {
    try {
      return languageFlags[i18n.language] || languageFlags.en
    } catch (error) {
      console.error("Error getting language flag:", error)
      return languageFlags.en // Fallback
    }
  }

  const currentLanguage = getCurrentLanguageName()
  const currentFlag = getCurrentLanguageFlag()

  return (
    <div className="relative" ref={ref}>
      {/* Hidden form for language submission */}
      <Form method="post" action="/set-language" ref={formRef} className="hidden">
        <input type="hidden" name="lang" />
      </Form>

      {/* Language selector - Navbar integrated version */}
      <div
        onClick={() => setIsOpen(!isOpen)}
        className="flex items-center gap-1 text-neutral-content hover:text-primary cursor-pointer transition-colors p-1"
      >
        <img
          src={currentFlag.src}
          alt={currentFlag.alt}
          className="w-5 h-4 object-cover rounded-sm"
        />
        <span className="text-sm font-medium hidden sm:inline">{currentLanguage}</span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className={`transition-transform ${isOpen ? "rotate-180" : ""}`}
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </div>

      {/* Dropdown menu - Only showing active languages */}
      {isOpen && (
        <div className="absolute right-0 mt-2 w-40 bg-base-300 rounded-lg shadow-lg border border-primary/20 z-50">
          <div className="p-1 space-y-1">
            {/* Dropdown menu */}
            {supportedLanguages.map((lang) => (
              <button
                key={lang}
                onClick={() => void handleLanguageChange(lang)}
                className={`w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center ${
                  i18n.language === lang
                    ? "bg-primary/20 text-primary font-medium"
                    : "hover:bg-base-200 text-neutral-content/80 hover:text-neutral-content"
                }`}
              >
                <img
                  src={languageFlags[lang].src}
                  alt={languageFlags[lang].alt}
                  className="w-5 h-4 mr-2 object-cover rounded-sm"
                />
                {languageNames[lang]}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

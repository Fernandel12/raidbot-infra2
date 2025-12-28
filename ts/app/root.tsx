import type { LinksFunction, LoaderFunctionArgs, LinkDescriptor } from "@remix-run/cloudflare"
import { json, redirect } from "@remix-run/cloudflare"
import {
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  useLoaderData,
  useLocation,
} from "@remix-run/react"
import { StrictMode, useEffect, useState } from "react"
import { I18nProvider } from "~/components/I18nProvider"
import Footer from "~/components/layout/Footer"
import Navbar from "~/components/layout/Navbar"
import FloatingDiscord from "~/components/ui/FloatingDiscord"
import { supportedLanguages } from "~/i18n"
import { i18nLoader } from "~/i18n/root-loader"
import { checkExistingSession } from "~/lib/services/auth"

import "./tailwind.css"

// Function to get canonical URL
function getCanonicalUrl(path: string): string {
  return `https://rslbot.com${path}`
}

export const links: LinksFunction = (): LinkDescriptor[] => {
  const linkDescriptors: LinkDescriptor[] = [
    { rel: "preconnect", href: "https://fonts.googleapis.com" },
    {
      rel: "preconnect",
      href: "https://fonts.gstatic.com",
      crossOrigin: "anonymous",
    },
    {
      rel: "stylesheet",
      href: "https://fonts.googleapis.com/css2?family=Manrope:wght@700&family=Noto+Sans:wght@400;500;700&family=Noto+Sans+TC:wght@400;500;700&display=swap",
    },
  ]
  return linkDescriptors
}

// Add the loader function to handle i18n
export async function loader({ request, params, context }: LoaderFunctionArgs) {
  try {
    const url = new URL(request.url)
    const hasLangParam = url.searchParams.has("lang")
    const i18nData = await i18nLoader({ request, params, context })

    // If a cookie needs to be set from URL param and we have a lang parameter
    if (i18nData.headers && hasLangParam) {
      // Redirect to the same URL without the lang parameter,
      // but keep the cookie header
      url.searchParams.delete("lang")
      return redirect(url.pathname + url.search, {
        headers: i18nData.headers,
      })
    }

    // If headers exist but no need to redirect
    if (i18nData.headers) {
      return json(
        { locale: i18nData.locale },
        {
          headers: i18nData.headers,
        }
      )
    }

    return { locale: i18nData.locale }
  } catch (error) {
    console.error("[ROOT] Error in root loader:", error)
    return { locale: "en" } // Fallback to English
  }
}

function ClientOnly({ children }: { children: React.ReactNode }) {
  const [isClient, setIsClient] = useState(false)

  useEffect(() => {
    setIsClient(true)
  }, [])

  return isClient ? <>{children}</> : null
}

function GoogleAnalytics() {
  useEffect(() => {
    // Check if scripts are already present to prevent duplicates
    if (document.querySelector('script[src*="googletagmanager.com/gtag/js"]')) {
      return // Scripts already exist, don't add more
    }

    // Add the gtag.js script
    const scriptTag = document.createElement("script")
    scriptTag.async = true
    scriptTag.src = "https://www.googletagmanager.com/gtag/js?id=G-MFWLL5VQ62" // Replace with your GA4 measurement ID
    document.head.appendChild(scriptTag)

    // Create and add the initialization script with first-party cookie settings
    const initScript = document.createElement("script")
    initScript.type = "text/javascript"
    initScript.textContent = `
      window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('js', new Date());
      gtag('config', 'G-MFWLL5VQ62', {
        'cookie_domain': 'rslbot.com',
        'cookie_flags': 'SameSite=None;Secure',
        'transport_url': 'https://rslbot.com'
      });
    `
    document.head.appendChild(initScript)
  }, [])

  return null
}

// Component to check session validity on app initialization
function SessionCheck() {
  const [initialized, setInitialized] = useState(false)

  useEffect(() => {
    // Only run once on client
    if (typeof window !== "undefined" && !initialized) {
      const validateSession = async () => {
        try {
          await checkExistingSession()
        } catch (error) {
          console.error("Session validation error:", error)
          // Session was cleared in checkExistingSession if invalid
        } finally {
          setInitialized(true)
        }
      }

      // Assign setTimeout to a variable
      const timeoutId = setTimeout(() => void validateSession(), 100)

      // Return cleanup function to clear the timeout
      return () => clearTimeout(timeoutId)
    }
  }, [initialized])

  return null
}

export default function App() {
  const location = useLocation()
  const loaderData = useLoaderData<typeof loader>()
  const locale = loaderData?.locale || "en"

  return (
    <html lang={locale === "tw" ? "zh-TW" : locale} data-theme="raidbot2">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
        <link rel="canonical" href={getCanonicalUrl(location.pathname)} />

        {/* Add alternate language links for SEO */}
        {supportedLanguages.map((lang) => (
          <link
            key={lang}
            rel="alternate"
            hrefLang={lang === "tw" ? "zh-TW" : lang}
            href={`https://rslbot.com${location.pathname}?lang=${lang}`}
          />
        ))}
      </head>
      <body>
        <I18nProvider locale={locale}>
          <div
            id="main"
            className="relative flex flex-col items-center justify-start overflow-visible p-0 gap-0 h-min min-h-lvh w-full"
          >
            <StrictMode>
              <Navbar />
              <ClientOnly>
                <SessionCheck />
                <GoogleAnalytics />
                <FloatingDiscord />
              </ClientOnly>
              <Outlet />
              <Footer />
            </StrictMode>
          </div>
        </I18nProvider>
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  )
}

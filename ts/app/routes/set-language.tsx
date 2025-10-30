import type { ActionFunction } from "@remix-run/cloudflare"
import { json, redirect } from "@remix-run/cloudflare"
import { defaultLanguage, i18nCookie, supportedLanguages } from "~/i18n"

export const action: ActionFunction = async ({ request }) => {
  const formData = await request.formData()
  const langValue = formData.get("lang")
  const lang = typeof langValue === "string" ? langValue : defaultLanguage

  // Validate language
  const validLang = supportedLanguages.includes(lang) ? lang : defaultLanguage

  // Get the referrer URL to redirect back to the same page
  const redirectTo = request.headers.get("Referer") || "/"

  // For XHR/fetch requests (client-side submissions), return JSON
  const isXHR =
    request.headers.get("X-Requested-With") === "XMLHttpRequest" ||
    request.headers.get("Accept")?.includes("application/json")

  if (isXHR) {
    return json(
      { success: true, language: validLang },
      {
        headers: {
          "Set-Cookie": await i18nCookie.serialize(validLang),
        },
      }
    )
  }

  // For regular form submissions (like from search engines), do a full redirect
  return redirect(redirectTo, {
    headers: {
      "Set-Cookie": await i18nCookie.serialize(validLang),
    },
  })
}

// If someone visits this route directly, redirect to home
export const loader = () => redirect("/")

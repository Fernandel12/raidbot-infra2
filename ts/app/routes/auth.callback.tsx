import { useNavigate } from "@remix-run/react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { handleDiscourseLogin } from "~/lib/services/auth"

export default function Callback() {
  const navigate = useNavigate()
  const [error, setError] = useState<string | null>(null)
  const { t } = useTranslation()

  useEffect(() => {
    const processCallback = async () => {
      try {
        // Get the SSO parameters from the URL
        const url = new URL(window.location.href)
        const sso = url.searchParams.get("sso")
        const sig = url.searchParams.get("sig")

        if (!sso || !sig) {
          throw new Error(t("errors.authRequired"))
        }

        // Handle the login process - this will:
        // 1. Save the Discourse user data (with username and avatar)
        // 2. Validate the session with the backend
        await handleDiscourseLogin(sso, sig)

        // Redirect to home on success
        navigate("/")
      } catch (error) {
        console.error("Auth callback error:", error)
        setError(error instanceof Error ? error.message : t("auth.loginError"))

        // Redirect after a short delay
        setTimeout(() => navigate("/"), 2000)
      }
    }

    void processCallback()
  }, [navigate, t])

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h2 className="text-xl font-semibold mb-4">{t("auth.loginError")}</h2>
          <p className="text-error mb-4">{error}</p>
          <p>{t("auth.redirecting")}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex items-center justify-center min-h-screen">
      <div className="text-center">
        <h2 className="text-xl font-semibold mb-4">{t("auth.processingLogin")}</h2>
        <div className="bg-primary/30 rounded-full h-2 w-32 mx-auto animate-pulse"></div>
      </div>
    </div>
  )
}

import type { LoaderFunctionArgs, MetaFunction } from "@remix-run/cloudflare"
import { json } from "@remix-run/cloudflare"
import { useSearchParams } from "@remix-run/react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import { getTranslator } from "~/i18n/utils"

export const loader = ({ request }: LoaderFunctionArgs) => {
  const url = new URL(request.url)
  const sessionId = url.searchParams.get("session_id")
  const token = url.searchParams.get("token")
  const payerId = url.searchParams.get("PayerID")

  // Determine if payment was successful based on parameters
  const isSuccess = !!(sessionId || (token && payerId))

  return json({ isSuccess })
}

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  // Get translator without hooks
  const t = getTranslator()

  // Use dynamic title based on payment status
  const title = data?.isSuccess
    ? t("paymentSuccess.successTitle")
    : t("paymentSuccess.failedTitle")

  return [
    { title },
    { name: "description", content: t("paymentSuccess.paymentStatus") },
    { property: "og:title", content: title },
    { property: "og:description", content: t("paymentSuccess.paymentStatus") },
    { property: "og:type", content: "website" },
  ]
}

export default function PaymentSuccessPage() {
  const [status, setStatus] = useState({ loading: true, success: false, message: "" })
  const [searchParams] = useSearchParams()
  const { t } = useTranslation()

  useEffect(() => {
    // Stripe returns session_id for successful payments
    const sessionId = searchParams.get("session_id")

    // PayPal returns token and PayerID for successful payments
    const token = searchParams.get("token")
    const payerId = searchParams.get("PayerID")

    // Check for Stripe success parameters
    if (sessionId) {
      setStatus({
        loading: false,
        success: true,
        message: t("paymentSuccess.successMessage"),
      })
    }
    // Check for PayPal success parameters
    else if (token && payerId) {
      setStatus({
        loading: false,
        success: true,
        message: t("paymentSuccess.successMessage"),
      })
    } else {
      setStatus({
        loading: false,
        success: false,
        message: t("paymentSuccess.errorMessage"),
      })
    }
  }, [searchParams, t])

  return (
    <section className="py-16 px-4 max-w-7xl mx-auto">
      <div className="flex flex-col gap-6 items-center justify-center">
        <h1 className="text-4xl md:text-5xl font-semibold text-white">
          {status.success ? t("paymentSuccess.title") : t("paymentSuccess.paymentStatus")}
        </h1>

        <div className="w-full max-w-lg text-center">
          {status.loading ? (
            <div className="flex flex-col items-center justify-center p-8">
              <div className="w-12 h-12 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
              <p className="mt-4 text-white/70">
                {t("paymentSuccess.verifyingPayment")}
              </p>
            </div>
          ) : (
            <div
              className={`p-6 rounded-lg ${status.success ? "bg-base-100 border border-green-500" : "bg-base-100 border border-yellow-500"}`}
            >
              <div className="text-center mb-4">
                {status.success ? (
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-16 w-16 mx-auto text-green-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                ) : (
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-16 w-16 mx-auto text-yellow-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
                    />
                  </svg>
                )}
              </div>

              <p className="text-lg mb-2 text-white/80">
                {status.message}
              </p>

              {status.success && (
                <>
                  <p className="text-sm mb-4 text-white/70">
                    {t("paymentSuccess.processingNote")}
                  </p>

                  <div className="flex flex-col gap-2 mt-6">
                    <a
                      href="/licenses"
                      className="btn border-primary border rounded-full py-3 px-6
                        bg-gradient-to-b from-transparent to-primary/30
                        hover:to-primary/40 transition-all duration-200"
                    >
                      {t("paymentSuccess.viewLicenses")}
                    </a>

                    <p className="text-xs italic text-white/70">
                      {t("paymentSuccess.refreshNote")}
                    </p>
                  </div>
                </>
              )}
            </div>
          )}
        </div>
      </div>
    </section>
  )
}

import type { MetaFunction } from "@remix-run/cloudflare"
import { useState } from "react"
import { useTranslation } from "react-i18next"
import { createPayPalCheckout } from "~/lib/api/payment"
import { useSessionStore } from "~/store/sessionStore"

export const meta: MetaFunction = () => {
  return [
    { title: "RaidBot - Purchase License" },
    { name: "description", content: "Purchase a license for RaidBot and unlock all features." },
    { property: "og:title", content: "RaidBot - Purchase License" },
    { property: "og:description", content: "Purchase a license for RaidBot and unlock all features." },
    { property: "og:type", content: "website" },
  ]
}

export default function PurchasePage() {
  const { t } = useTranslation()

  const [selectedDuration, setSelectedDuration] = useState("ONE_MONTH")
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const { isAuthenticated, ssoToken } = useSessionStore()

  const handleCheckout = async (e: React.FormEvent) => {
    e.preventDefault()

    if (!isAuthenticated || !ssoToken) {
      setError("You must be logged in to make a purchase.")
      return
    }

    try {
      setIsSubmitting(true)
      setError(null)

      // Use PayPal checkout
      const checkoutData = await createPayPalCheckout(ssoToken, {
        licenseDuration: selectedDuration,
      })

      if (!checkoutData.checkoutUrl) {
        throw new Error("No checkout URL received")
      }

      // Redirect to PayPal checkout
      window.location.href = checkoutData.checkoutUrl
    } catch (error) {
      console.error("Error creating checkout:", error)
      setError(error instanceof Error ? error.message : "An error occurred")
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <section className="py-10 md:py-16 px-4 max-w-7xl mx-auto">
      <div className="flex flex-col gap-6 items-center justify-center mb-10">
        <h1 className="text-4xl md:text-5xl font-semibold text-white">
          RaidBot License
        </h1>
        <p className="text-center max-w-2xl text-white/80">
          Purchase a license to unlock all bot features.
        </p>
      </div>

      {/* Main content with stacked layout */}
      <div className="flex flex-col gap-8 max-w-4xl mx-auto">
        {/* Features comparison */}
        <div className="w-full space-y-6">
          {/* Feature List */}
          <div className="bg-base-200 rounded-xl p-6 shadow-md">
            <h3 className="text-xl font-bold mb-4 text-primary">
              Premium Features
            </h3>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 1 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 2 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 3 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 4 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 5 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 6 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 7 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 8 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 9 - Placeholder</span>
              </div>
              <div className="flex items-start gap-2">
                <span className="text-primary text-xl">✓</span>
                <span>Feature 10 - Placeholder</span>
              </div>
            </div>
          </div>

          {/* License Information - Prominent Notice */}
          <div className="bg-primary/10 border-2 border-primary/30 rounded-xl p-4 shadow-md">
            <div className="flex items-center gap-2">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-6 w-6 text-primary"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              <h3 className="text-xl font-bold text-primary">
                {t("license.licenseType")}
              </h3>
            </div>
            <p className="mt-2 font-medium text-white">
              Each license is valid for one machine only. Multiple devices require separate licenses.
            </p>
          </div>
        </div>

        {/* Purchase form */}
        <div className="w-full">
          <div className="bg-base-100 rounded-xl p-6 lg:p-8 shadow-lg border border-primary/20">
            {error && (
              <div className="bg-red-500 border border-red-400 text-red-700 px-4 py-3 rounded mb-6">
                <p className="font-bold">{t("global.error")}</p>
                <p className="text-red-900">{error}</p>
              </div>
            )}

            <form onSubmit={(e) => void handleCheckout(e)} className="space-y-6">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <h3 className="text-xl font-bold text-white">
                    Select License Duration
                  </h3>
                </div>

                {/* Pricing cards */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                  <label
                    htmlFor="duration-week"
                    aria-label="Select 1 week license"
                    className={`flex flex-col p-5 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedDuration === "ONE_WEEK"
                        ? "border-primary bg-primary/5 shadow-md"
                        : "border-base-300 hover:border-primary/50"
                    }`}
                  >
                    <div className="flex items-center gap-3 mb-2">
                      <input
                        id="duration-week"
                        type="radio"
                        name="duration"
                        value="ONE_WEEK"
                        checked={selectedDuration === "ONE_WEEK"}
                        onChange={() => setSelectedDuration("ONE_WEEK")}
                        className="radio radio-primary"
                      />
                      <span className="font-semibold text-lg">
                        {t("license.duration.oneWeek")}
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-baseline">
                        <span className="text-primary text-4xl font-bold">
                          $5
                        </span>
                        <span className="text-base-content/70 ml-1">
                          /week
                        </span>
                      </div>
                    </div>
                  </label>

                  <label
                    htmlFor="duration-month"
                    aria-label="Select 1 month license"
                    className={`flex flex-col p-5 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedDuration === "ONE_MONTH"
                        ? "border-primary bg-primary/5 shadow-md"
                        : "border-base-300 hover:border-primary/50"
                    }`}
                  >
                    <div className="flex items-center gap-3 mb-2">
                      <input
                        id="duration-month"
                        type="radio"
                        name="duration"
                        value="ONE_MONTH"
                        checked={selectedDuration === "ONE_MONTH"}
                        onChange={() => setSelectedDuration("ONE_MONTH")}
                        className="radio radio-primary"
                      />
                      <span className="font-semibold text-lg flex items-center">
                        {t("license.duration.oneMonth")}
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-baseline">
                        <span className="text-primary text-4xl font-bold">
                          $15
                        </span>
                        <span className="text-base-content/70 ml-1">
                          /month
                        </span>
                      </div>
                    </div>
                  </label>

                  {/* Yearly License Option */}
                  <label
                    htmlFor="duration-year"
                    aria-label="Select 1 year license"
                    className={`flex flex-col p-5 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedDuration === "ONE_YEAR"
                        ? "border-primary bg-primary/5 shadow-md"
                        : "border-base-300 hover:border-primary/50"
                    }`}
                  >
                    <div className="flex items-center gap-3 mb-2">
                      <input
                        id="duration-year"
                        type="radio"
                        name="duration"
                        value="ONE_YEAR"
                        checked={selectedDuration === "ONE_YEAR"}
                        onChange={() => setSelectedDuration("ONE_YEAR")}
                        className="radio radio-primary"
                      />
                      <span className="font-semibold text-lg">
                        {t("license.duration.oneYear")}
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-baseline">
                        <span className="text-primary text-4xl font-bold">
                          $120
                        </span>
                        <span className="text-base-content/70 ml-1">
                          /year
                        </span>
                      </div>
                    </div>
                  </label>

                  {/* Lifetime License Option */}
                  <label
                    htmlFor="duration-lifetime"
                    aria-label="Select lifetime license"
                    className={`flex flex-col p-5 rounded-xl border-2 cursor-pointer transition-all ${
                      selectedDuration === "LIFETIME"
                        ? "border-primary bg-primary/5 shadow-md"
                        : "border-base-300 hover:border-primary/50"
                    }`}
                  >
                    <div className="flex items-center gap-3 mb-2">
                      <input
                        id="duration-lifetime"
                        type="radio"
                        name="duration"
                        value="LIFETIME"
                        checked={selectedDuration === "LIFETIME"}
                        onChange={() => setSelectedDuration("LIFETIME")}
                        className="radio radio-primary"
                      />
                      <span className="font-semibold text-lg">
                        {t("license.duration.lifetime")}
                      </span>
                    </div>
                    <div className="mt-2">
                      <div className="flex items-baseline">
                        <span className="text-primary text-4xl font-bold">
                          $300
                        </span>
                        <span className="text-base-content/70 ml-1">
                          one-time
                        </span>
                      </div>
                    </div>
                  </label>
                </div>
              </div>

              {/* Payment info */}
              <div className="bg-base-200/50 rounded-xl p-4">
                <div className="flex items-center justify-center gap-3">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 188 49"
                    className="h-8"
                  >
                    <path
                      fill="#0070E0"
                      d="M164.01 11.446l-4.012 25.207a.643.643 0 0 0 .642.746h4.748a.701.701 0 0 0 .698-.589l4.012-25.207a.643.643 0 0 0-.642-.746h-4.748a.692.692 0 0 0-.698.589zm-5.07 7.356h-4.505a.699.699 0 0 0-.697.588l-.149.928s-3.499-3.794-9.694-1.23c-3.554 1.468-5.26 4.501-5.986 6.723 0 0-2.304 6.753 2.907 10.47 0 0 4.832 3.575 10.273-.22l-.094.592a.644.644 0 0 0 .37.686c.085.04.178.06.272.06h4.508a.692.692 0 0 0 .698-.589l2.742-17.262a.632.632 0 0 0-.149-.521.643.643 0 0 0-.496-.226zm-6.629 9.54a5.005 5.005 0 0 1-1.715 3.095 5.073 5.073 0 0 1-3.345 1.203 4.602 4.602 0 0 1-1.416-.206c-1.945-.62-3.055-2.474-2.736-4.484a5.01 5.01 0 0 1 1.717-3.093 5.08 5.08 0 0 1 3.343-1.207 4.6 4.6 0 0 1 1.416.208c1.957.616 3.062 2.473 2.741 4.485h-.005zm-24.056.477c2.443 0 4.806-.868 6.662-2.446a10.147 10.147 0 0 0 3.456-6.158c.789-4.993-3.14-9.351-8.71-9.351h-8.973a.699.699 0 0 0-.697.589L115.98 36.66a.644.644 0 0 0 .37.686c.086.04.178.06.272.06h4.751a.699.699 0 0 0 .697-.589l1.178-7.402a.692.692 0 0 1 .698-.59l4.309-.006zm3.974-8.831c-.293 1.846-1.731 3.205-4.482 3.205h-3.517l1.068-6.713h3.454c2.844.005 3.77 1.67 3.477 3.513v-.005z"
                    />
                    <path
                      fill="#003087"
                      d="M110.567 19.23l-5.434 9.105-2.758-9.038a.694.694 0 0 0-.672-.495h-4.904a.526.526 0 0 0-.527.446.515.515 0 0 0 .025.247l4.942 15.224-4.47 7.174a.516.516 0 0 0 .18.728.527.527 0 0 0 .269.07h5.282a.876.876 0 0 0 .751-.42l13.804-22.667a.512.512 0 0 0 .011-.53.524.524 0 0 0-.463-.263h-5.28a.877.877 0 0 0-.756.419zm-16.548-.428H89.51a.7.7 0 0 0-.698.59l-.146.927s-3.502-3.794-9.697-1.23c-3.553 1.468-5.26 4.501-5.983 6.723 0 0-2.306 6.753 2.904 10.47 0 0 4.833 3.575 10.274-.22l-.094.592a.642.642 0 0 0 .37.686c.085.04.178.06.272.06h4.508a.701.701 0 0 0 .697-.589l2.743-17.262a.642.642 0 0 0-.37-.687.655.655 0 0 0-.272-.06zm-6.63 9.542a5.011 5.011 0 0 1-1.716 3.091 5.082 5.082 0 0 1-3.343 1.206 4.605 4.605 0 0 1-1.414-.206c-1.944-.62-3.053-2.474-2.734-4.485a5.011 5.011 0 0 1 1.723-3.098 5.082 5.082 0 0 1 3.353-1.201c.48-.005.959.065 1.417.208 1.937.616 3.04 2.472 2.72 4.485h-.005zm-24.055.476a10.284 10.284 0 0 0 6.656-2.449 10.144 10.144 0 0 0 3.452-6.156c.79-4.992-3.14-9.35-8.708-9.35H55.76a.7.7 0 0 0-.698.588l-4 25.2a.642.642 0 0 0 .37.687c.085.039.178.06.272.06h4.748a.7.7 0 0 0 .698-.59l1.176-7.402a.692.692 0 0 1 .698-.589h4.31zm3.974-8.832c-.293 1.846-1.73 3.205-4.481 3.205H59.31l1.066-6.713h3.454c2.845.005 3.77 1.671 3.478 3.513v-.005z"
                    />
                    <path
                      fill="#001C64"
                      d="M32.639 12.16c.107-5.566-4.484-9.836-10.797-9.836H8.784a1.277 1.277 0 0 0-1.262 1.078L2.29 36.095a1.038 1.038 0 0 0 1.025 1.2h7.736l-1.209 7.57a1.038 1.038 0 0 0 1.025 1.2h6.302c.304 0 .575-.109.807-.306.23-.198.268-.471.316-.772l1.85-10.884c.047-.3.2-.69.431-.888.231-.198.433-.306.738-.306h3.856c6.183 0 11.428-4.395 12.387-10.507.679-4.338-1.181-8.286-4.915-10.243z"
                    />
                    <path
                      fill="#0070E0"
                      d="M12.725 25.238l-1.927 12.218-1.21 7.664a1.038 1.038 0 0 0 1.026 1.199h6.67a1.276 1.276 0 0 0 1.26-1.078l1.758-11.139a1.277 1.277 0 0 1 1.261-1.078h3.926c6.183 0 11.428-4.51 12.388-10.622.68-4.338-1.504-8.286-5.238-10.243-.01.462-.05.923-.121 1.38-.959 6.11-6.206 10.621-12.387 10.621h-6.145a1.278 1.278 0 0 0-1.261 1.079"
                    />
                    <path
                      fill="#003087"
                      d="M10.797 37.456h-7.76a1.037 1.037 0 0 1-1.024-1.2L7.245 3.078A1.277 1.277 0 0 1 8.506 2h13.336c6.313 0 10.904 4.594 10.797 10.159-1.571-.824-3.417-1.295-5.439-1.295H16.082a1.277 1.277 0 0 0-1.262 1.078l-2.094 13.296-1.93 12.218z"
                    />
                  </svg>
                  <span className="text-white/80 text-sm">Secure payment via PayPal</span>
                </div>
              </div>

              {/* Submit button */}
              <div className="pt-6">
                <button
                  type="submit"
                  disabled={!isAuthenticated || isSubmitting}
                  className={`btn border-primary border rounded-full py-3 px-6 w-full text-sm
                    bg-gradient-to-b from-transparent to-primary/30
                    hover:to-primary/40 transition-all duration-200
                    ${!isAuthenticated ? "opacity-50 cursor-not-allowed" : ""}`}
                >
                  <span className="inline-block">
                    {isSubmitting
                      ? t("global.processing")
                      : `Proceed to Checkout • ${
                          selectedDuration === "ONE_WEEK"
                            ? "$5"
                            : selectedDuration === "ONE_MONTH"
                              ? "$15"
                              : selectedDuration === "ONE_YEAR"
                                ? "$120"
                                : "$300"
                        }`}
                  </span>
                </button>

                {!isAuthenticated && (
                  <p className="mt-3 text-center text-red-600">
                    {t("auth.loginRequired")}
                  </p>
                )}
              </div>

              {/* Payment details */}
              <div className="mt-4 text-sm text-base-content/70 flex flex-col justify-center h-full">
                <p>{t("payment.securePayment")}</p>

                {/* Refund Policy */}
                <div className="mt-4 p-4 border border-base-300 rounded-lg bg-base-200/50">
                  <p className="font-semibold text-base">{t("purchase.refundPolicy.title")}</p>
                  <p className="mt-2">{t("purchase.refundPolicy.description")}</p>
                </div>
              </div>
            </form>
          </div>
        </div>
      </div>
    </section>
  )
}

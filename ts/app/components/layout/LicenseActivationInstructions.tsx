import { X } from "lucide-react"
import React, { useState } from "react"
import { useTranslation } from "react-i18next"
import AnimateElementIn from "~/components/ui/AnimateElementIn"
import BtnPrimary from "~/components/ui/BtnPrimary"

// This component can be added to the LicensesRoute component in licenses.tsx
function LicenseActivationInstructions() {
  const { t } = useTranslation()
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className="w-full max-w-5xl mx-auto mb-8">
      {!isOpen ? (
        <div className="flex justify-center">
          <BtnPrimary onClick={() => setIsOpen(true)} className="text-sm">
            {t("license.activateTitle")}
          </BtnPrimary>
        </div>
      ) : (
        <AnimateElementIn
          transition="scale"
          className="bg-base-200/90 rounded-xl p-5 shadow-xl border border-base-300/30"
        >
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-2xl font-semibold text-white">
              {t("license.activateTitle")}
            </h3>
            <button onClick={() => setIsOpen(false)} className="p-1 hover:bg-base-300 rounded-full">
              <X className="h-4 w-4" />
            </button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="bg-base-100 p-4 rounded-lg border border-base-300/40">
              <div className="text-center mb-3">
                <h4 className="text-lg font-medium text-white">
                  {t("license.activateStep1")}
                </h4>
              </div>
              <img
                src="/images/activate_license1.png"
                alt={t("license.activateStep1")}
                className="mb-3 mx-auto rounded-md border border-base-300"
              />
              <p
                className="text-sm"
                dangerouslySetInnerHTML={{ __html: t("license.activateStep1Details") }}
              ></p>
            </div>

            <div className="bg-base-100 p-4 rounded-lg border border-base-300/40">
              <div className="text-center mb-3">
                <h4 className="text-lg font-medium text-white">
                  {t("license.activateStep2")}
                </h4>
              </div>
              <img
                src="/images/activate_license2.png"
                alt={t("license.activateStep2")}
                className="mb-3 mx-auto rounded-md border border-base-300"
              />
              <p className="text-sm">{t("license.activateStep2Details")}</p>
              <p className="text-sm mt-2">{t("license.activateStep2Additional")}</p>
            </div>

            <div className="bg-base-100 p-4 rounded-lg border border-base-300/40">
              <div className="text-center mb-3">
                <h4 className="text-lg font-medium text-white">
                  {t("license.activateStep3")}
                </h4>
              </div>
              <img
                src="/images/activate_license3.png"
                alt={t("license.activateStep3")}
                className="mb-3 mx-auto rounded-md border border-base-300"
              />
              <p className="text-sm">{t("license.activateStep3Details")}</p>
            </div>
          </div>

          <div className="mt-6 bg-base-300/50 p-4 rounded-lg">
            <h4 className="text-lg font-medium text-white mb-2">
              {t("license.troubleshootingTitle")}
            </h4>
            <ul className="list-disc pl-5 text-sm space-y-1">
              <li dangerouslySetInnerHTML={{ __html: t("license.troubleInvalidKey") }}></li>
              <li dangerouslySetInnerHTML={{ __html: t("license.troubleAlreadyActivated") }}></li>
              <li dangerouslySetInnerHTML={{ __html: t("license.troubleExpired") }}></li>
            </ul>
          </div>
        </AnimateElementIn>
      )}
    </div>
  )
}

export default LicenseActivationInstructions

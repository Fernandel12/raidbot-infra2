import { useTranslation } from "react-i18next"
import PricingCard from "./PricingCard"
import { theme } from "~/lib/theme"

export default function PricingSection() {
  const { t } = useTranslation()

  const pricingPlans = [
    {
      title: t("pricing.free.title"),
      price: t("pricing.free.price"),
      description: t("pricing.free.description"),
      features: t("pricing.free.features", { returnObjects: true }) as string[],
      buttonText: t("pricing.free.buttonText"),
      buttonLink:
        "https://community.raidbot.app/index.php?/store/product/11-free-license/",
    },
    {
      title: t("pricing.regular.title"),
      price: t("pricing.regular.price"),
      description: t("pricing.regular.description"),
      features: t("pricing.regular.features", { returnObjects: true }) as string[],
      buttonText: t("pricing.regular.buttonText"),
      buttonLink:
        "https://community.raidbot.app/index.php?/store/product/3-regular-license/",
    },
    {
      title: t("pricing.premium.title"),
      price: t("pricing.premium.price"),
      description: t("pricing.premium.description"),
      features: t("pricing.premium.features", { returnObjects: true }) as string[],
      buttonText: t("pricing.premium.buttonText"),
      buttonLink:
        "https://community.raidbot.app/index.php?/store/product/9-premium-license/",
    },
  ]

  return (
    <section className="py-16 px-4 bg-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-left mb-12 ml-10">
          <h2 className="text-4xl font-bold" style={{ color: theme.colors.primary.dark }}>
            {t("pricing.title")}
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 justify-items-center">
          {pricingPlans.map((plan, idx) => (
            <PricingCard key={idx} {...plan} />
          ))}
        </div>

        {/* Payment methods */}
        <div className="mt-12 text-center">
          <p className="text-gray-600 mb-4">{t("pricing.securedPayment")}</p>
          <div className="flex justify-center items-center gap-6">
            <span className="text-sm text-gray-500">
              {t("pricing.paymentMethods")}
            </span>
          </div>
        </div>
      </div>
    </section>
  )
}

import { useTranslation } from "react-i18next"
import PricingCard from "./PricingCard"

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
        "https://community.rslbot.com/index.php?/store/product/11-free-license/",
    },
    {
      title: t("pricing.regular.title"),
      price: t("pricing.regular.price"),
      description: t("pricing.regular.description"),
      features: t("pricing.regular.features", { returnObjects: true }) as string[],
      buttonText: t("pricing.regular.buttonText"),
      buttonLink:
        "https://community.rslbot.com/index.php?/store/product/3-regular-license/",
    },
    {
      title: t("pricing.premium.title"),
      price: t("pricing.premium.price"),
      description: t("pricing.premium.description"),
      features: t("pricing.premium.features", { returnObjects: true }) as string[],
      buttonText: t("pricing.premium.buttonText"),
      buttonLink:
        "https://community.rslbot.com/index.php?/store/product/9-premium-license/",
    },
  ]

  return (
    <section className="py-20 px-4 bg-white">
      <div className="max-w-7xl mx-auto">
        {/* Section Title */}
        <div className="text-center mb-4">
          <h2 className="text-4xl md:text-5xl font-bold text-gray-900">
            {t("pricing.title")}
          </h2>
        </div>

        {/* Subtitle */}
        <div className="text-center mb-16">
          <p className="text-gray-600 text-lg max-w-2xl mx-auto">
            {t("pricing.subtitle")}
          </p>
        </div>

        {/* Pricing Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 justify-items-center mb-12">
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

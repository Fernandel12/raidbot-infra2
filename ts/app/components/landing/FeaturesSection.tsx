import { useTranslation } from "react-i18next"
import FeatureCard from "./FeatureCard"

export default function FeaturesSection() {
  const { t } = useTranslation()

  const features = [
    {
      icon: (
        <svg
          className="w-10 h-10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="white"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <rect x="2" y="4" width="20" height="16" rx="2" />
          <path d="M6 8h.01M10 8h.01M14 8h.01M18 8h.01M8 12h.01M12 12h.01M16 12h.01M6 16h.01M10 16h.01M14 16h.01M18 16h.01" />
        </svg>
      ),
      title: t("features.farmWhileAfk.title"),
      description: t("features.farmWhileAfk.description"),
      bullets: t("features.farmWhileAfk.bullets", { returnObjects: true }) as string[],
    },
    {
      icon: (
        <svg
          className="w-10 h-10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="white"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
          <circle cx="9" cy="7" r="4" />
          <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
          <path d="M16 3.13a4 4 0 0 1 0 7.75" />
        </svg>
      ),
      title: t("features.community.title"),
      description: t("features.community.description"),
      bullets: t("features.community.bullets", { returnObjects: true }) as string[],
    },
    {
      icon: (
        <svg
          className="w-10 h-10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="white"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="M9 12l2 2 4-4" />
          <circle cx="12" cy="12" r="10" />
        </svg>
      ),
      title: t("features.safeAndEasy.title"),
      description: t("features.safeAndEasy.description"),
      bullets: t("features.safeAndEasy.bullets", { returnObjects: true }) as string[],
    },
  ]

  return (
    <section id="features" className="py-20 px-4 bg-base-200">
      <div className="max-w-7xl mx-auto">
        {/* Section Title */}
        <div className="text-center mb-16">
          <h2 className="text-4xl md:text-5xl font-bold text-base-content">{t("features.title")}</h2>
        </div>

        {/* Feature Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {features.map((feature, idx) => (
            <FeatureCard key={idx} {...feature} />
          ))}
        </div>
      </div>
    </section>
  )
}

import { useTranslation } from "react-i18next"
import FeatureCard from "./FeatureCard"
import { theme } from "~/lib/theme"

export default function FeaturesSection() {
  const { t } = useTranslation()

  const features = [
    {
      icon: (
        <svg
          className="w-24 h-24"
          viewBox="0 0 120 120"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M60 110C87.614 110 110 87.614 110 60C110 32.386 87.614 10 60 10C32.386 10 10 32.386 10 60C10 87.614 32.386 110 60 110Z"
            fill="white"
            opacity="0.9"
          />
          <rect x="40" y="50" width="40" height="30" rx="2" fill="#352C71" />
          <rect x="35" y="40" width="50" height="8" rx="1" fill="#352C71" />
        </svg>
      ),
      title: t("features.farmWhileAfk.title"),
      description: t("features.farmWhileAfk.description"),
    },
    {
      icon: (
        <svg
          className="w-24 h-24"
          viewBox="0 0 120 120"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle cx="40" cy="45" r="15" fill="white" opacity="0.9" />
          <circle cx="80" cy="45" r="15" fill="white" opacity="0.9" />
          <circle cx="60" cy="75" r="15" fill="white" opacity="0.9" />
        </svg>
      ),
      title: t("features.community.title"),
      description: t("features.community.description"),
    },
    {
      icon: (
        <svg
          className="w-24 h-24"
          viewBox="0 0 120 120"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M35 55L55 75L85 35"
            stroke="white"
            strokeWidth="8"
            strokeLinecap="round"
            strokeLinejoin="round"
            opacity="0.9"
          />
        </svg>
      ),
      title: t("features.safeAndEasy.title"),
      description: t("features.safeAndEasy.description"),
    },
  ]

  return (
    <section className="py-16 px-4 bg-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-left mb-12 ml-10">
          <h2 className="text-4xl font-bold" style={{ color: theme.colors.primary.dark }}>
            {t("features.title")}
          </h2>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 justify-items-center">
          {features.map((feature, idx) => (
            <FeatureCard key={idx} {...feature} />
          ))}
        </div>
      </div>
    </section>
  )
}

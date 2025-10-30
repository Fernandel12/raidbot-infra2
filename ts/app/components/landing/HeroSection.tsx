import { useTranslation } from "react-i18next"

export default function HeroSection() {
  const { t } = useTranslation()

  return (
    <section className="pt-32 pb-12 px-4 bg-gradient-to-b from-indigo-50 to-white">
      <div className="max-w-7xl mx-auto text-center">
        <img
          src="/logo.png"
          alt="RaidBot Logo"
          className="w-32 h-32 mx-auto mb-6"
        />
        <h1 className="text-5xl md:text-6xl font-bold text-indigo-900 mb-4">
          {t("header.title")}
        </h1>
        <h2 className="text-2xl md:text-3xl text-indigo-700 mb-4">
          {t("header.tagline")}
        </h2>
        <p className="text-gray-600 text-lg">{t("header.description")}</p>
      </div>
    </section>
  )
}

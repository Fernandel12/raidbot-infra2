import { useTranslation } from "react-i18next"
import { Link } from "@remix-run/react"

export default function HeroSection() {
  const { t } = useTranslation()

  return (
    <section className="pt-24 pb-16 px-4 bg-gradient-to-b from-gray-50 to-white">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-12 items-center">
          {/* Left Column - Text Content */}
          <div className="space-y-6">
            <h1 className="text-5xl md:text-6xl font-bold text-gray-900">
              {t("hero.title")}
              <br />
              <span className="text-indigo-600">{t("hero.titleHighlight")}</span>
            </h1>
            <p className="text-lg text-gray-600 leading-relaxed">{t("hero.description")}</p>
            <div className="flex flex-col sm:flex-row gap-4 pt-4">
              <Link
                to="https://community.rslbot.com/index.php?/store/product/11-free-license/"
                className="px-8 py-4 bg-indigo-600 text-white font-semibold rounded-lg hover:bg-indigo-700 transition-colors text-center shadow-lg hover:shadow-xl"
              >
                {t("hero.startFreeButton")}
              </Link>
              <a
                href="#features"
                className="px-8 py-4 bg-white text-indigo-600 font-semibold rounded-lg border-2 border-indigo-600 hover:bg-indigo-50 transition-colors text-center"
              >
                {t("hero.learnMoreButton")}
              </a>
            </div>
          </div>

          {/* Right Column - Feature Card */}
          <div className="relative">
            <div className="rounded-3xl p-8 bg-gradient-to-br from-indigo-600 via-purple-600 to-indigo-800 shadow-2xl">
              {/* Header */}
              <div className="text-center mb-6">
                <div className="flex items-center justify-center gap-2 mb-2">
                  <h2 className="text-3xl font-bold text-white">{t("hero.cardTitle")}</h2>
                  <span className="text-2xl">😊</span>
                </div>
                <p className="text-indigo-200 text-sm font-medium mb-4">{t("hero.cardTagline")}</p>
                <p className="text-white text-sm max-w-md mx-auto">{t("hero.cardDescription")}</p>
              </div>

              {/* Robot Illustration Placeholder */}
              <div className="relative my-8 flex items-center justify-center">
                <div className="w-48 h-48 bg-indigo-400/30 rounded-2xl backdrop-blur-sm flex items-center justify-center">
                  <div className="text-center">
                    <div className="text-6xl mb-2">🤖</div>
                    <div className="w-32 h-2 bg-cyan-400/50 rounded-full mx-auto"></div>
                    <div className="w-24 h-2 bg-cyan-400/30 rounded-full mx-auto mt-2"></div>
                  </div>
                </div>

                {/* Feature Badges */}
                <div className="absolute right-0 top-1/2 -translate-y-1/2 space-y-3">
                  <div className="bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg border border-white/30">
                    <p className="text-white text-xs font-bold whitespace-nowrap">
                      {t("hero.badge1")}
                    </p>
                  </div>
                  <div className="bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg border border-white/30">
                    <p className="text-white text-xs font-bold whitespace-nowrap">
                      {t("hero.badge2")}
                    </p>
                  </div>
                  <div className="bg-white/20 backdrop-blur-sm px-3 py-2 rounded-lg border border-white/30">
                    <p className="text-white text-xs font-bold whitespace-nowrap">
                      {t("hero.badge3")}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}

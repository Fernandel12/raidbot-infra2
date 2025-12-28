import { useTranslation } from "react-i18next"
import { links } from "~/lib/theme"

// TODO: Fix accessibility - replace placeholder # hrefs with proper navigation or buttons
/* eslint-disable jsx-a11y/anchor-is-valid */
export default function Footer() {
  const { t } = useTranslation()
  const currentYear = new Date().getFullYear()

  return (
    <footer className="w-full bg-neutral/90 text-white py-8 md:py-16">
      <div className="max-w-[1200px] mx-auto px-4 sm:px-8">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-6 md:gap-8">
          <div>
            <h6 className="text-lg font-bold text-white">{t("footer.raidBot")}</h6>
            <ul className="mt-3 md:mt-4 space-y-2">
              <li>
                <a href="/" className="text-sm text-neutral-400 hover:text-primary">
                  {t("global.home")}
                </a>
              </li>
              <li>
                <a href="/purchase" className="text-sm text-neutral-400 hover:text-primary">
                  {t("global.purchase")}
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-neutral-400 hover:text-primary">
                  {t("footer.aboutUs")}
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h6 className="text-lg font-bold text-white">{t("footer.resources")}</h6>
            <ul className="mt-3 md:mt-4 space-y-2">
              <li>
                <a
                  href="https://community.rslbot.com/c/guides/beginners-guide"
                  className="text-sm text-neutral-400 hover:text-primary"
                >
                  {t("footer.gettingStarted")}
                </a>
              </li>
              <li>
                <a
                  href="https://community.rslbot.com/c/guides"
                  className="text-sm text-neutral-400 hover:text-primary"
                >
                  {t("footer.guides")}
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-neutral-400 hover:text-primary">
                  {t("footer.faq")}
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h6 className="text-lg font-bold text-white">{t("footer.community")}</h6>
            <ul className="mt-3 md:mt-4 space-y-2">
              <li>
                <a
                  href={links.discord}
                  className="text-sm text-neutral-400 hover:text-primary"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  Discord
                </a>
              </li>
              <li>
                <a
                  href={links.forum}
                  className="text-sm text-neutral-400 hover:text-primary"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {t("global.forum")}
                </a>
              </li>
              <li>
                <a
                  href="https://rslbot.com/"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-sm text-neutral-400 hover:text-primary"
                >
                  {t("footer.raidShadowLegendsBot")}
                </a>
              </li>
            </ul>
          </div>

          <div>
            <h6 className="text-lg font-bold text-white">{t("footer.legal")}</h6>
            <ul className="mt-3 md:mt-4 space-y-2">
              <li>
                <a href="#" className="text-sm text-neutral-400 hover:text-primary">
                  {t("footer.termsOfService")}
                </a>
              </li>
              <li>
                <a href="#" className="text-sm text-neutral-400 hover:text-primary">
                  {t("footer.privacyPolicy")}
                </a>
              </li>
            </ul>
          </div>
        </div>

        <div className="mt-8 md:mt-16 pt-6 md:pt-8 border-t border-neutral-800 text-neutral-400 text-center">
          <p className="text-sm">
            © {currentYear} RaidBot. {t("footer.allRightsReserved")}
          </p>
        </div>
      </div>
    </footer>
  )
}

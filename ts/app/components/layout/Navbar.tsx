import { Link } from "@remix-run/react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"
import LanguageSelector from "~/components/ui/LanguageSelector"
import UserMenu from "~/components/ui/UserMenu"
import { links } from "~/lib/theme"

export default function Navbar() {
  const { t } = useTranslation()
  const [isMobile, setIsMobile] = useState(false)

  // Detect mobile viewport
  useEffect(() => {
    const checkMobile = () => {
      setIsMobile(window.innerWidth < 768)
    }
    checkMobile()
    window.addEventListener("resize", checkMobile)
    return () => window.removeEventListener("resize", checkMobile)
  }, [])

  return (
    <>
      <div className="sticky top-0 left-0 right-0 w-full z-10 bg-neutral/90 backdrop-blur-sm border-b border-neutral-content/10">
        <div className="flex items-center justify-between mx-auto px-4 sm:px-8 py-4">
          {/* Logo/Brand */}
          <Link to="/" className="flex items-center gap-2">
            <span className="text-xl font-bold text-primary">{t("products.raidBot2")}</span>
          </Link>

          {/* Mobile: Only UserMenu */}
          <div className="flex md:hidden items-center">
            <UserMenu />
          </div>

          {/* Desktop Nav items */}
          <div className="hidden md:flex items-center gap-6">
            <LanguageSelector />

            <a
              href="https://community.rslbot.com/c/eb2-releases/8"
              target="_blank"
              rel="noopener noreferrer"
              className="text-neutral-content/80 hover:text-neutral-content transition-colors text-sm"
            >
              {t("global.download")}
            </a>

            <Link
              to="/purchase"
              className="text-neutral-content/80 hover:text-neutral-content transition-colors text-sm"
            >
              {t("global.purchase")}
            </Link>

            <a
              href={links.discord}
              className="text-neutral-content/80 hover:text-neutral-content transition-colors text-sm"
              target="_blank"
              rel="noopener noreferrer"
            >
              {t("global.discordServer")}
            </a>

            <a
              href={links.forum}
              target="_blank"
              className="text-neutral-content/80 hover:text-neutral-content transition-colors text-sm"
            >
              {t("global.forum")}
            </a>

            <UserMenu />
          </div>
        </div>
      </div>

      {/* Mobile Sticky Bottom Navigation */}
      {isMobile && (
        <div className="fixed bottom-0 left-0 right-0 bg-neutral/95 backdrop-blur-sm border-t border-neutral-content/10 z-50 md:hidden">
          <div className="flex items-center justify-around py-2">
            {/* Download */}
            <a
              href="https://community.rslbot.com/c/eb2-releases/8"
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-col items-center p-2 text-neutral-content/80 hover:text-primary transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
                <polyline points="7 10 12 15 17 10"></polyline>
                <line x1="12" x2="12" y1="15" y2="3"></line>
              </svg>
              <span className="text-xs mt-1">{t("global.download")}</span>
            </a>

            {/* Purchase */}
            <Link
              to="/purchase"
              className="flex flex-col items-center p-2 text-neutral-content/80 hover:text-primary transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <rect width="20" height="14" x="2" y="5" rx="2"></rect>
                <line x1="2" x2="22" y1="10" y2="10"></line>
              </svg>
              <span className="text-xs mt-1">{t("global.purchase")}</span>
            </Link>

            {/* Discord Server */}
            <a
              href={links.discord}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-col items-center p-2 text-neutral-content/80 hover:text-primary transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="currentColor"
              >
                <path d="M19.27 5.33C17.94 4.71 16.5 4.26 15 4a.09.09 0 0 0-.07.03c-.18.33-.39.76-.53 1.09a16.09 16.09 0 0 0-4.8 0c-.14-.34-.35-.76-.54-1.09c-.01-.02-.04-.03-.07-.03c-1.5.26-2.93.71-4.27 1.33c-.01 0-.02.01-.03.02c-2.72 4.07-3.47 8.03-3.1 11.95c0 .02.01.04.03.05c1.8 1.32 3.53 2.12 5.24 2.65c.03.01.06 0 .07-.02c.4-.55.76-1.13 1.07-1.74c.02-.04 0-.08-.04-.09c-.57-.22-1.11-.48-1.64-.78c-.04-.02-.04-.08-.01-.11c.11-.08.22-.17.33-.25c.02-.02.05-.02.07-.01c3.44 1.57 7.15 1.57 10.55 0c.02-.01.05-.01.07.01c.11.09.22.17.33.26c.04.03.04.09-.01.11c-.52.31-1.07.56-1.64.78c-.04.01-.05.06-.04.09c.32.61.68 1.19 1.07 1.74c.03.01.06.02.09.01c1.72-.53 3.45-1.33 5.25-2.65c.02-.01.03-.03.03-.05c.44-4.53-.73-8.46-3.1-11.95c-.01-.01-.02-.02-.04-.02zM8.52 14.91c-1.03 0-1.89-.95-1.89-2.12s.84-2.12 1.89-2.12c1.06 0 1.9.96 1.89 2.12c0 1.17-.84 2.12-1.89 2.12zm6.97 0c-1.03 0-1.89-.95-1.89-2.12s.84-2.12 1.89-2.12c1.06 0 1.9.96 1.89 2.12c0 1.17-.83 2.12-1.89 2.12z"></path>
              </svg>
              <span className="text-xs mt-1">Discord</span>
            </a>

            {/* Forum */}
            <a
              href={links.forum}
              target="_blank"
              rel="noopener noreferrer"
              className="flex flex-col items-center p-2 text-neutral-content/80 hover:text-primary transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                width="20"
                height="20"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <path d="M7 10h10"></path>
                <path d="M7 14h5"></path>
                <path d="M19 5H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h8l4 4V7a2 2 0 0 0-2-2Z"></path>
              </svg>
              <span className="text-xs mt-1">{t("global.forum")}</span>
            </a>
          </div>
        </div>
      )}
    </>
  )
}

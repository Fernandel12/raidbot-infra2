import { useNavigate } from "@remix-run/react"
import { useEffect, useRef, useState } from "react"
import { useTranslation } from "react-i18next"
import { getDiscourseLoginURL } from "~/lib/api/discourse"
import { logout } from "~/lib/services/auth"
import { useSessionStore } from "~/store/sessionStore"
import BtnPrimary from "./BtnPrimary"

export default function UserMenu() {
  const { discourseUser, isAuthenticated, loading } = useSessionStore()
  const [menuOpen, setMenuOpen] = useState(false)
  const [loggingOut, setLoggingOut] = useState(false)
  const [isMobile, setIsMobile] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const navigate = useNavigate()
  const { t } = useTranslation()

  // Check if on mobile device
  useEffect(() => {
    if (typeof window === "undefined") return

    const checkMobile = () => {
      setIsMobile(window.innerWidth < 640)
    }

    checkMobile()
    window.addEventListener("resize", checkMobile)

    return () => {
      window.removeEventListener("resize", checkMobile)
    }
  }, [])

  // Handle clicking outside to close menu
  useEffect(() => {
    if (typeof window === "undefined") return

    function handleClickOutside(event: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(event.target as Node)) {
        setMenuOpen(false)
      }
    }

    document.addEventListener("mousedown", handleClickOutside)
    return () => {
      document.removeEventListener("mousedown", handleClickOutside)
    }
  }, [])

  const handleLoginClick = () => {
    const loginURL = getDiscourseLoginURL()
    window.location.href = loginURL
  }

  const handleLicensesClick = () => {
    navigate("/licenses")
    setMenuOpen(false)
  }

  const handleLogoutClick = async () => {
    try {
      setLoggingOut(true)
      await logout()
      setMenuOpen(false)
    } catch (error) {
      console.error("Logout failed:", error)
    } finally {
      setLoggingOut(false)
    }
  }

  if (loading) {
    return (
      <div className="relative h-8 md:h-9 w-16 md:w-24 flex items-center justify-center">
        <div className="animate-pulse bg-primary/30 rounded-full h-2 w-12 md:w-16"></div>
      </div>
    )
  }

  if (!isAuthenticated) {
    return (
      <BtnPrimary
        onClick={handleLoginClick}
        className="text-xs md:text-sm py-1 md:py-1.5 px-3 md:px-4"
      >
        {t("global.login")}
      </BtnPrimary>
    )
  }

  // Ensure avatar URL has proper protocol
  const avatarUrl = discourseUser?.avatarUrl
    ? discourseUser.avatarUrl.replace("http://", "https://")
    : null

  return (
    <div className="relative" ref={menuRef}>
      <button
        ref={buttonRef}
        onClick={() => setMenuOpen(!menuOpen)}
        className="flex items-center gap-2 bg-base-300 hover:bg-base-200 rounded-full px-2 sm:px-3 py-1 sm:py-1.5 transition-colors border border-primary/20 hover:border-primary/40"
      >
        {/* User Avatar */}
        {avatarUrl ? (
          <div className="w-5 h-5 md:w-6 md:h-6 rounded-full overflow-hidden">
            <img
              src={avatarUrl}
              alt={discourseUser?.username}
              className="w-full h-full object-cover"
              onError={(e) => {
                // Fallback if image fails to load
                ;(e.target as HTMLImageElement).style.display = "none"
                ;(e.target as HTMLImageElement).parentElement!.innerHTML =
                  `<div class="w-5 h-5 md:w-6 md:h-6 rounded-full bg-primary flex items-center justify-center text-primary-content font-semibold text-xs">
                    ${discourseUser?.username?.charAt(0).toUpperCase() || "U"}
                  </div>`
              }}
            />
          </div>
        ) : (
          <div className="w-5 h-5 md:w-6 md:h-6 rounded-full bg-primary flex items-center justify-center text-primary-content font-semibold text-xs">
            {discourseUser?.username?.charAt(0).toUpperCase() || "U"}
          </div>
        )}

        <span className="text-xs md:text-sm text-neutral-content/90 max-w-16 md:max-w-32 truncate hidden sm:block">
          {discourseUser?.username || "User"}
        </span>
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="16"
          height="16"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          className={`transition-transform ${menuOpen ? "rotate-180" : ""}`}
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      </button>

      {menuOpen && (
        <div
          className="fixed sm:absolute right-0 left-auto sm:mt-2 w-40 sm:w-48 bg-base-300 rounded-lg shadow-lg border border-primary/20 z-[100]"
          style={{
            top:
              isMobile && buttonRef.current
                ? `${buttonRef.current.getBoundingClientRect().bottom + 8}px`
                : "auto",
            right: isMobile ? "1rem" : "auto",
          }}
        >
          {/* Display admin badge if applicable */}
          {discourseUser?.admin && (
            <div className="p-2 border-b border-primary/10">
              <div className="text-xs text-primary font-medium px-3 py-2">Admin</div>
            </div>
          )}

          {/* Username on mobile */}
          <div className="sm:hidden p-2 border-b border-primary/10">
            <div className="text-xs text-neutral-content/90 font-medium px-3 py-2">
              {discourseUser?.username || "User"}
            </div>
          </div>

          {/* Menu options */}
          <div className="p-2 space-y-1">
            {/* Licenses option */}
            <button
              onClick={handleLicensesClick}
              className="w-full text-left px-4 py-2 text-sm text-neutral-content/80 hover:bg-base-200 rounded hover:text-neutral-content transition-colors"
            >
              {t("global.manageLicenses")}
            </button>
            {/* Logout option */}
            <button
              onClick={() => void handleLogoutClick()}
              disabled={loggingOut}
              className="w-full text-left px-4 py-2 text-xs sm:text-sm text-neutral-content/80 hover:bg-base-200 rounded hover:text-neutral-content transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loggingOut ? t("global.loggingOut") : t("global.logout")}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

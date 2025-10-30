import type { DiscourseSSOUser } from "~/store/sessionStore"
import CryptoJS from "crypto-js"
import { API_CONFIG, getAppUrl } from "~/lib/utils/api"

/**
 * Generates the URL for Discourse SSO authentication
 */
export function getDiscourseLoginURL(): string {
  // Generate a nonce
  const nonce = `nonce-${Date.now()}-${Math.random().toString(36).substring(2, 15)}`

  // Define the return URL with proper protocol and host
  const returnURL = getAppUrl("/auth/callback")

  // Create the payload
  const payload = `return_sso_url=${encodeURIComponent(returnURL)}&nonce=${encodeURIComponent(nonce)}`

  // Base64 encode the payload
  const base64Payload = btoa(payload)

  // Create HMAC-SHA256 signature
  const sig = CryptoJS.HmacSHA256(base64Payload, API_CONFIG.DISCOURSE_SECRET).toString(
    CryptoJS.enc.Hex
  )

  // Return the complete SSO URL
  return `${API_CONFIG.DISCOURSE_URL}/session/sso_provider?sso=${encodeURIComponent(base64Payload)}&sig=${encodeURIComponent(sig)}`
}

/**
 * Extracts and parses SSO payload from Discourse callback
 */
export function parseDiscoursePayload(sso: string): DiscourseSSOUser | null {
  try {
    // Decode base64 SSO payload
    const decodedPayload = atob(sso)
    const params = new URLSearchParams(decodedPayload)

    // Extract user information
    const username = params.get("username") || ""
    const email = params.get("email") || ""
    const externalId = params.get("external_id") || ""
    const name = params.get("name") || undefined
    const avatarUrl = params.get("avatar_url") || undefined
    const groupsString = params.get("groups") || ""
    const groups = groupsString ? groupsString.split(",") : undefined
    const admin = params.get("admin") === "true"

    if (!username || !email || !externalId) {
      return null
    }

    return {
      username,
      email,
      externalId,
      name,
      avatarUrl,
      groups,
      admin,
    }
  } catch (error) {
    console.error("Failed to parse Discourse SSO payload:", error)
    return null
  }
}

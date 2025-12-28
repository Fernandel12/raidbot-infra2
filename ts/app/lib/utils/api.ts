/**
 * Environment configuration constants
 * These can be accessed anywhere in the application
 */
export const API_CONFIG = {
  // Base URLs
  API_BASE_URL:
    process.env.NODE_ENV === "development" ? "http://localhost:8080" : "https://api.rslbot.com",

  DISCOURSE_URL: "https://community.rslbot.com",

  // Discourse secret key (should be in env vars in production)
  DISCOURSE_SECRET: "hHDjDtwn6ADb3Gv",
}

/**
 * Get the application's base URL based on current environment
 */
export function getAppBaseUrl(): string {
  if (typeof window !== "undefined") {
    return `${window.location.protocol}//${window.location.host}`
  }

  // Default fallback for server-side rendering
  return "https://rslbot.com"
}

/**
 * Get a full URL with the current application base
 */
export function getAppUrl(path = "/"): string {
  const baseUrl = getAppBaseUrl()
  return `${baseUrl}${path.startsWith("/") ? path : `/${path}`}`
}

/**
 * Format a token for use in authorization headers
 */
export function formatAuthToken(token: string): string {
  return token.startsWith("SSO_") ? token : `SSO_${token}`
}

/**
 * Create authorization headers with the given token
 */
export function createAuthHeaders(token: string | null): HeadersInit {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  }

  if (token) {
    headers.Authorization = `Bearer ${formatAuthToken(token)}`
  }

  return headers
}

/**
 * Generic API request function with error handling
 */
export async function apiRequest<T>(
  endpoint: string,
  options: RequestInit = {},
  token: string | null = null
): Promise<T> {
  const url = `${API_CONFIG.API_BASE_URL}${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`

  // Add authorization headers if token is provided
  const headers = createAuthHeaders(token)

  try {
    const response = await fetch(url, {
      ...options,
      headers: {
        ...headers,
        ...options.headers,
      },
    })

    if (!response.ok) {
      let errorMessage: string

      try {
        const errorData = await response.json()
        errorMessage =
          (errorData &&
          typeof errorData === "object" &&
          "message" in errorData &&
          typeof errorData.message === "string"
            ? errorData.message
            : null) || `Request failed with status ${response.status}`
      } catch {
        errorMessage = `Request failed with status ${response.status}`
      }

      throw new Error(errorMessage)
    }

    return (await response.json()) as T
  } catch (error) {
    console.error(`API request error (${url}):`, error)
    throw error
  }
}

import { parseDiscoursePayload } from "~/lib/api/discourse"
import { getUserSession, logoutUser as logoutUserApi } from "~/lib/api/user"
import { useSessionStore } from "~/store/sessionStore"
import type { DiscourseSSOUser } from "~/store/sessionStore"

/**
 * Handles the Discourse SSO login process
 */
export async function handleDiscourseLogin(sso: string, sig: string): Promise<boolean> {
  // Create the auth token
  const token = `SSO_${sso}.${sig}`

  // Parse user info from SSO payload
  const discourseUser = parseDiscoursePayload(sso)

  if (!discourseUser) {
    throw new Error("Invalid SSO payload")
  }

  // Store the token and complete user info from Discourse
  const sessionStore = useSessionStore.getState()
  sessionStore.setSession(discourseUser, token)

  // Just validate the session with backend
  try {
    await getUserSession(token)
    return true
  } catch (error) {
    console.error("Error validating session with backend:", error)
    sessionStore.clearSession()
    return false
  }
}

/**
 * Checks if there's a stored session and validates with backend
 */
export async function checkExistingSession(): Promise<boolean> {
  const sessionStore = useSessionStore.getState()
  const { isAuthenticated, ssoToken, discourseUser } = sessionStore

  // Check if we have a session and user data
  if (!isAuthenticated || !ssoToken || !discourseUser) {
    sessionStore.clearSession()
    return false
  }

  try {
    // Just validate that the token is still valid with the backend
    await getUserSession(ssoToken)
    return true
  } catch (error) {
    console.error("Existing session is invalid:", error)
    sessionStore.clearSession()
    return false
  }
}

/**
 * Initialize user session with provided user data
 */
export function initializeUserSession(user: DiscourseSSOUser, token: string): DiscourseSSOUser {
  try {
    const sessionStore = useSessionStore.getState()
    sessionStore.setSession(user, token)
    return user
  } catch (error) {
    let errorMessage = "Failed to initialize user session"
    if (error instanceof Error) {
      errorMessage = error.message
    }
    useSessionStore.getState().setError(errorMessage)
    throw error
  }
}

/**
 * Log out the user using the backend API and clear local state
 */
export async function logout(): Promise<void> {
  try {
    const sessionStore = useSessionStore.getState()
    const { ssoToken, discourseUser } = sessionStore

    // Only call the API if we have a token and user
    if (ssoToken && discourseUser) {
      try {
        await logoutUserApi(ssoToken)
      } catch (apiError) {
        // Log API errors but continue with local logout
        console.error("Error calling logout API:", apiError)
      }
    }
  } catch (error) {
    console.error("Error during logout process:", error)
  } finally {
    // Always clear the local session state
    useSessionStore.getState().clearSession()
  }
}

import { create } from "zustand"
import { createJSONStorage, persist } from "zustand/middleware"

// Type for Discourse user information
export interface DiscourseSSOUser {
  username: string
  email: string
  externalId: string
  name?: string
  avatarUrl?: string
  groups?: string[]
  admin?: boolean
}

interface SessionState {
  discourseUser: DiscourseSSOUser | null
  ssoToken: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null

  // Actions
  setSession: (user: DiscourseSSOUser, token: string) => void
  setToken: (token: string) => void
  setLoading: () => void
  setError: (error: string) => void
  clearSession: () => void
  checkSessionValidity: () => boolean
}

export const useSessionStore = create<SessionState>()(
  persist(
    (set, get) => ({
      discourseUser: null,
      ssoToken: null,
      isAuthenticated: false,
      loading: false,
      error: null,

      // Set user and token after successful authentication
      setSession: (user, token) =>
        set({
          discourseUser: user,
          ssoToken: token,
          isAuthenticated: true,
          loading: false,
          error: null,
        }),

      // Set just the token
      setToken: (token) =>
        set({
          ssoToken: token,
        }),

      // Start loading state
      setLoading: () =>
        set({
          loading: true,
          error: null,
        }),

      // Handle authentication errors
      setError: (error) =>
        set({
          loading: false,
          error,
        }),

      // Clear session state on logout
      clearSession: () =>
        set({
          discourseUser: null,
          ssoToken: null,
          isAuthenticated: false,
          loading: false,
          error: null,
        }),

      // Check if the current session is valid
      checkSessionValidity: () => {
        const { discourseUser, ssoToken, isAuthenticated } = get()
        return isAuthenticated && !!discourseUser && !!ssoToken
      },
    }),
    {
      name: "session-storage",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({
        discourseUser: state.discourseUser,
        ssoToken: state.ssoToken,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)

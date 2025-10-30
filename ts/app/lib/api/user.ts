import { apiRequest } from "~/lib/utils/api"
import type { LicenseDuration } from "~/lib/api/license"

export enum SubscriptionStatus {
  STATUS_UNSPECIFIED = 0,
  STATUS_ACTIVE = 1,
  STATUS_CANCELED = 2,
}

export interface Subscription {
  id: number
  createdAt: string
  updatedAt: string

  stripeSubscriptionId: string
  stripeCustomerId: string
  status: SubscriptionStatus
  duration: LicenseDuration
  currentPeriodStart: string
  currentPeriodEnd: string
  sandboxMode: boolean

  userId: number
  licenseKeyId: number
}

interface GetSubscriptionsResponse {
  subscriptions: Subscription[]
}

interface LogoutResponse {
  success: boolean
}

interface GetSessionResponse {
  user: {
    id: string
    discourseId: string
  }
  userId: number
  licenseKeyId: number
}

/**
 * Get current session information
 */
export async function getUserSession(ssoToken: string): Promise<GetSessionResponse["user"]> {
  const response = await apiRequest<GetSessionResponse>(
    "/user/session",
    { method: "GET" },
    ssoToken
  )

  return response.user
}

/**
 * Get all subscriptions for the authenticated user
 */
export async function getUserSubscriptions(ssoToken: string): Promise<Subscription[]> {
  const response = await apiRequest<GetSubscriptionsResponse>(
    "/user/subscriptions",
    { method: "GET" },
    ssoToken
  )

  return response.subscriptions
}

/**
 * Log out the user using the backend API
 */
export async function logoutUser(ssoToken: string): Promise<void> {
  await apiRequest<LogoutResponse>("/user/logout", { method: "POST" }, ssoToken)
}

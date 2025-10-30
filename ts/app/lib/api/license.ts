import { apiRequest } from "~/lib/utils/api"

export enum LicenseDuration {
  UNSPECIFIED = 0,
  LIFETIME = 1,
  ONE_DAY = 2,
  ONE_WEEK = 3,
  ONE_MONTH = 5,
  THREE_MONTHS = 6,
  ONE_YEAR = 7,
}

// Map string enum values to their numeric counterparts
export const mapStringDurationToEnum = (durationStr: string): LicenseDuration => {
  switch (durationStr) {
    case "LIFETIME":
      return LicenseDuration.LIFETIME
    case "ONE_DAY":
      return LicenseDuration.ONE_DAY
    case "ONE_WEEK":
      return LicenseDuration.ONE_WEEK
    case "ONE_MONTH":
      return LicenseDuration.ONE_MONTH
    case "THREE_MONTHS":
      return LicenseDuration.THREE_MONTHS
    case "ONE_YEAR":
      return LicenseDuration.ONE_YEAR
    case "UNSPECIFIED":
      return LicenseDuration.UNSPECIFIED
    default:
      return LicenseDuration.UNSPECIFIED
  }
}

export interface LicenseKey {
  id: number
  createdAt: string
  updatedAt: string

  effectiveFrom: string
  key: string
  revoked: boolean
  duration: LicenseDuration | string
  activeUsageId: string
  uses: number | string
  sandboxMode: boolean

  userId: number | string

  // Frontend-specific computed fields
  status?: "ACTIVE" | "EXPIRED" | "REVOKED"
  version?: string
}

interface GetLicensesResponse {
  licenses: LicenseKey[]
}

interface SetupAutoRenewalResponse {
  checkoutUrl?: string
}

/**
 * Get all licenses for the authenticated user
 */
export async function getUserLicenses(ssoToken: string): Promise<LicenseKey[]> {
  const response = await apiRequest<GetLicensesResponse>(
    "/user/licenses",
    { method: "GET" },
    ssoToken
  )

  return response.licenses
}

/**
 * Setup auto-renewal for a license
 */
export async function setupAutoRenewal(
  ssoToken: string,
  licenseKeyId: string
): Promise<{ checkoutUrl?: string }> {
  const response = await apiRequest<SetupAutoRenewalResponse>(
    "/license/auto-renewal/setup",
    {
      method: "POST",
      body: JSON.stringify({
        license_key_id: licenseKeyId,
      }),
    },
    ssoToken
  )

  return {
    checkoutUrl: response.checkoutUrl,
  }
}

/**
 * Cancel auto-renewal for a license
 */
export async function cancelAutoRenewal(ssoToken: string, licenseKeyId: string): Promise<void> {
  await apiRequest(
    "/license/auto-renewal/cancel",
    {
      method: "POST",
      body: JSON.stringify({
        license_key_id: licenseKeyId,
      }),
    },
    ssoToken
  )
}

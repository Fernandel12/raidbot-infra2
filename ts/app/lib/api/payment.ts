import { apiRequest } from "~/lib/utils/api"

interface PayPalCheckoutResponse {
  orderId: string
  checkoutUrl: string
}

interface CreateCheckoutParams {
  licenseDuration: string
  renewalKey?: string
}

/**
 * Create a PayPal checkout session via the backend API
 */
export async function createPayPalCheckout(
  ssoToken: string,
  params: CreateCheckoutParams
): Promise<PayPalCheckoutResponse> {
  const response = await apiRequest<PayPalCheckoutResponse>(
    "/payment/paypal/create-checkout",
    {
      method: "POST",
      body: JSON.stringify({
        license_duration: params.licenseDuration,
        renewal_key_id: params.renewalKey || 0,
      }),
    },
    ssoToken
  )

  // Validate response
  if (!response.checkoutUrl) {
    throw new Error("Server returned incomplete data (missing checkout URL)")
  }

  return response
}

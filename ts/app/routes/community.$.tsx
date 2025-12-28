import type { LoaderFunctionArgs } from "@remix-run/cloudflare"
import { redirect } from "@remix-run/cloudflare"

// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function loader({ request, params }: LoaderFunctionArgs) {
  // Get the wildcard part of the route
  const wildcardPath = params["*"] || ""

  // Redirect to the target domain with the same path
  return redirect(`https://community.rslbot.com/${wildcardPath}`, {
    status: 301,
  })
}

export default function CommunityCatchAll() {
  return null // This component won't render as the redirect happens first
}

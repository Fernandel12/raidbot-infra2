import type { LoaderFunction } from "@remix-run/cloudflare"
import { generateSitemap } from "~/sitemap.server"

export const loader: LoaderFunction = ({ request }) => {
  return generateSitemap(request)
}

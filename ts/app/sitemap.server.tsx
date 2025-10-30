export function generateSitemap(request: Request) {
  const hostname = new URL(request.url).origin
  const routes = [
    { url: "/", changefreq: "weekly", priority: 1 },
    { url: "/ebclassic", changefreq: "weekly", priority: 1 },
    { url: "/purchase", changefreq: "weekly", priority: 1 },
    { url: "/licenses", changefreq: "monthly", priority: 1 },
  ]

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
  <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
    ${routes
      .map(
        (route) => `
    <url>
      <loc>${hostname}${route.url}</loc>
      <changefreq>${route.changefreq}</changefreq>
      <priority>${route.priority}</priority>
    </url>
    `
      )
      .join("")}
  </urlset>`

  return new Response(xml, {
    headers: {
      "Content-Type": "application/xml",
    },
  })
}

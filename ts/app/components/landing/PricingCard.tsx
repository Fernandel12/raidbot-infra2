import { theme } from "~/lib/theme"

interface PricingCardProps {
  title: string
  price: string
  description: string
  features: string[]
  buttonText: string
  buttonLink: string
}

export default function PricingCard({
  title,
  price,
  description,
  features,
  buttonText,
  buttonLink,
}: PricingCardProps) {
  return (
    <div
      className="rounded-2xl hover:shadow-xl transition-shadow overflow-hidden max-w-[315px] min-h-[400px] flex flex-col"
      style={{
        background: theme.gradients.pricing,
        boxShadow: theme.shadows.card,
      }}
    >
      {/* Price header */}
      <div className="bg-base-100 py-4 px-6 text-center">
        <h3 className="text-primary text-2xl font-bold">{price}</h3>
      </div>

      {/* Content */}
      <div className="flex-1 p-6 flex flex-col items-center text-center">
        <p className="text-white text-sm uppercase font-semibold mb-3">{title}</p>
        <p className="text-white text-sm mb-6">{description}</p>

        <ul className="space-y-3 mb-8 text-white text-sm">
          {features.map((feature, idx) => (
            <li key={idx}>{feature}</li>
          ))}
        </ul>

        <a
          href={buttonLink}
          target="_blank"
          rel="noopener noreferrer"
          className="mt-auto bg-base-100 hover:bg-base-200 text-base-content font-semibold py-3 px-8 rounded-lg transition-colors"
        >
          {buttonText}
        </a>
      </div>
    </div>
  )
}

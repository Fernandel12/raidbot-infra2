interface FeatureCardProps {
  icon: React.ReactNode
  title: string
  description: string
  bullets?: string[]
}

export default function FeatureCard({ icon, title, description, bullets }: FeatureCardProps) {
  return (
    <div className="bg-base-100 rounded-2xl p-8 shadow-lg hover:shadow-xl transition-shadow border border-base-300">
      {/* Icon */}
      <div className="w-20 h-20 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center mb-6">
        {icon}
      </div>

      {/* Title */}
      <h3 className="text-2xl font-bold text-base-content mb-3">{title}</h3>

      {/* Description */}
      <p className="text-base-content/70 mb-6 leading-relaxed">{description}</p>

      {/* Bullet Points */}
      {bullets && bullets.length > 0 && (
        <ul className="space-y-3">
          {bullets.map((bullet, idx) => (
            <li key={idx} className="flex items-start gap-3">
              <svg
                className="w-5 h-5 text-success mt-0.5 flex-shrink-0"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
              <span className="text-base-content/80 text-sm">{bullet}</span>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}

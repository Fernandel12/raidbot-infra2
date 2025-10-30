import { theme } from "~/lib/theme"

interface FeatureCardProps {
  icon: React.ReactNode
  title: string
  description: string
}

export default function FeatureCard({ icon, title, description }: FeatureCardProps) {
  return (
    <div
      className="rounded-2xl p-8 hover:shadow-xl transition-shadow min-h-[450px] flex flex-col items-center text-center"
      style={{
        background: theme.gradients.feature,
        boxShadow: theme.shadows.card,
      }}
    >
      <div className="h-32 flex items-center justify-center mb-6">
        {icon}
      </div>
      <h3 className="text-white text-xl font-semibold mb-4">{title}</h3>
      <p className="text-white text-sm leading-relaxed">{description}</p>
    </div>
  )
}

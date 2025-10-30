export default function BtnPrimary({
  children,
  onClick,
  className,
}: {
  children: React.ReactNode
  onClick?: () => void
  className?: string
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`btn border-primary border rounded-full py-3 px-6 min-w-[64px]
        bg-gradient-to-b from-transparent to-primary/30
        hover:to-primary/40 transition-all duration-200
        ${className || ""}`}
    >
      {children}
    </button>
  )
}

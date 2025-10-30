export default function IconSvg({
  path,
  size = "7",
  color = "primary",
}: {
  path: string
  size?: string
  color?: string
}) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 256 256"
      focusable="false"
      className={`text-${color} w-${size} h-${size}`}
    >
      <g>
        <path d={path} fill="currentColor" />
      </g>
    </svg>
  )
}

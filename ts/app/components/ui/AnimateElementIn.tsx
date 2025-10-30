import { motion } from "framer-motion"
import { forwardRef } from "react"

const TRANSITIONS = {
  slideDown: {
    initial: { opacity: 0, y: -30 },
    animate: { opacity: 1, y: 0 },
  },
  slideUp: {
    initial: { opacity: 0, y: 30 },
    animate: { opacity: 1, y: 0 },
  },
  scale: {
    initial: { opacity: 0, scale: 0.5 },
    animate: { opacity: 1, scale: 1 },
  },
}

function AnimateElementIn(
  {
    children,
    transition,
    duration = 0.28,
    className,
  }: {
    children: React.ReactNode
    transition: keyof typeof TRANSITIONS
    duration?: number
    className?: string
  },
  ref?: React.ForwardedRef<HTMLDivElement>
) {
  return (
    <motion.div
      className={className}
      initial={TRANSITIONS[transition].initial}
      animate={TRANSITIONS[transition].animate}
      transition={{ duration }}
      ref={ref}
    >
      {children}
    </motion.div>
  )
}

export default forwardRef(AnimateElementIn)

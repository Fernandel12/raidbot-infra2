/**
 * Toast - Simple notification component
 */

import { AlertCircle, CheckCircle, X } from "lucide-react"
import { useEffect } from "react"

export interface ToastProps {
  type: "success" | "error" | "info"
  message: string
  onClose: () => void
  duration?: number // in milliseconds, 0 for manual close only
}

export default function Toast({ type, message, onClose, duration = 5000 }: ToastProps) {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onClose()
      }, duration)

      return () => clearTimeout(timer)
    }
  }, [duration, onClose])

  const bgColor =
    type === "success"
      ? "bg-green-100 dark:bg-green-900/90 border-green-400 dark:border-green-600"
      : type === "error"
        ? "bg-red-100 dark:bg-red-900/90 border-red-400 dark:border-red-600"
        : "bg-blue-100 dark:bg-blue-900/90 border-blue-400 dark:border-blue-600"

  const textColor =
    type === "success"
      ? "text-green-900 dark:text-green-100"
      : type === "error"
        ? "text-red-900 dark:text-red-100"
        : "text-blue-900 dark:text-blue-100"

  const Icon = type === "success" ? CheckCircle : AlertCircle
  const iconColor =
    type === "success"
      ? "text-green-700 dark:text-green-300"
      : type === "error"
        ? "text-red-700 dark:text-red-300"
        : "text-blue-700 dark:text-blue-300"

  return (
    <div
      className={`fixed top-20 right-4 z-[9999] flex items-center gap-3 p-4 border-2 rounded-lg shadow-xl backdrop-blur-sm ${bgColor} ${textColor} max-w-md`}
      role="alert"
      aria-live="polite"
    >
      <Icon className={`w-5 h-5 flex-shrink-0 ${iconColor}`} />
      <p className="flex-1 text-sm">{message}</p>
      <button
        type="button"
        onClick={onClose}
        className="flex-shrink-0 p-1 rounded hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        aria-label="Close notification"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  )
}

/**
 * ToastContainer - Manages multiple toasts
 */
interface ToastContainerProps {
  toasts: Array<{
    id: string
    type: "success" | "error" | "info"
    message: string
  }>
  onRemove: (id: string) => void
}

export function ToastContainer({ toasts, onRemove }: ToastContainerProps) {
  return (
    <div className="fixed top-20 right-4 z-[9999] flex flex-col gap-2">
      {toasts.map((toast) => (
        <Toast
          key={toast.id}
          type={toast.type}
          message={toast.message}
          onClose={() => onRemove(toast.id)}
        />
      ))}
    </div>
  )
}

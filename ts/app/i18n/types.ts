/**
 * Defines the TypeScript types for complex i18next translation objects.
 */

/**
 * Represents the structure of a single FAQ item from the translation files.
 * This type accommodates various formats (simple answer, lists of steps, etc.).
 */
export interface FaqItem {
  question: string
  answer?: string
  setupSteps?: string[]
  limitationsList?: string[]
  controlsList?: { key: string; action: string }[]
  pickitSteps?: string[]
  logDisableSteps?: string[]
  logDisableNote?: string
}

/**
 * Represents the 'faq.items' object returned by i18next, which is a dictionary
 * of FAQ items, keyed by their unique identifiers (e.g., "setup", "limitations").
 */
export type FaqItemsObject = Record<string, FaqItem>

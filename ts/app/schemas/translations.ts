import { z } from "zod"

/**
 * Schema for FAQ item with optional special fields for different types
 */
export const FaqItemSchema = z.object({
  question: z.string(),
  answer: z.string(),
  setupSteps: z.array(z.string()).optional(),
  limitationsList: z.array(z.string()).optional(),
  controlsList: z
    .array(
      z.object({
        key: z.string(),
        action: z.string(),
      })
    )
    .optional(),
  pickitSteps: z.array(z.string()).optional(),
  logDisableSteps: z.array(z.string()).optional(),
  logDisableNote: z.string().optional(),
})

/**
 * Schema for the entire FAQ items object (keyed by id)
 */
export const FaqItemsSchema = z.record(z.string(), FaqItemSchema)

/**
 * Schema for translation meta information
 */
export const TranslationMetaSchema = z.object({
  title: z.string(),
  description: z.string(),
})

export type FaqItem = z.infer<typeof FaqItemSchema>
export type FaqItems = z.infer<typeof FaqItemsSchema>
export type TranslationMeta = z.infer<typeof TranslationMetaSchema>

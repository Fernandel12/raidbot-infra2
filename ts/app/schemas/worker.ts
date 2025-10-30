import { z } from "zod"

/**
 * Schema for BaseItemType from game data JSON
 */
export const BaseItemTypeSchema = z.object({
  Name: z.string(),
}).passthrough() // Allow additional unknown properties

/**
 * Schema for ItemClass from game data JSON
 */
export const ItemClassSchema = z.object({
  Name: z.string(),
}).passthrough() // Allow additional unknown properties

/**
 * Schema for Stat from game data JSON
 */
export const StatSchema = z.object({
  Id: z.string(),
}).passthrough() // Allow additional unknown properties

/**
 * Schema for the combined game data
 */
export const GameDataSchema = z.object({
  baseTypes: z.array(BaseItemTypeSchema),
  itemClasses: z.array(ItemClassSchema),
  stats: z.array(StatSchema),
})

/**
 * Schema for successful worker response
 */
export const WorkerSuccessResponseSchema = z.object({
  success: z.literal(true),
  data: GameDataSchema,
})

/**
 * Schema for error worker response
 */
export const WorkerErrorResponseSchema = z.object({
  success: z.literal(false),
  error: z.string(),
})

/**
 * Union schema for any worker response
 */
export const WorkerResponseSchema = z.union([
  WorkerSuccessResponseSchema,
  WorkerErrorResponseSchema,
])

export type WorkerResponse = z.infer<typeof WorkerResponseSchema>

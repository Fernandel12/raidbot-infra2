#!/usr/bin/env node

import fs from "fs"
import path from "path"
import { fileURLToPath } from "url"
import { parseArgs } from "util"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// Parse command-line arguments
const options = {
  verbose: {
    type: "boolean",
    short: "v",
    default: false,
    description: "Show verbose output including all unused keys",
  },
  namespace: {
    type: "string",
    short: "n",
    description: "Filter results by namespace (e.g., common, eb2, classic)",
  },
  output: {
    type: "string",
    short: "o",
    default: "console",
    description: "Output format: console, json, or csv",
  },
  threshold: {
    type: "string",
    short: "t",
    default: "0",
    description: "Only show namespaces with usage below this percentage",
  },
  help: {
    type: "boolean",
    short: "h",
    description: "Show help message",
  },
  debug: {
    type: "boolean",
    short: "d",
    description: "Show debug output including file scanning details",
  },
}

let args
try {
  const { values } = parseArgs({ options, allowPositionals: false })
  args = values
} catch (error) {
  console.error("Error parsing arguments:", error.message)
  process.exit(1)
}

if (args.help) {
  console.log("Usage: node find-unused-translations.mjs [options]\n")
  console.log("Options:")
  Object.entries(options).forEach(([key, opt]) => {
    const shortFlag = opt.short ? `-${opt.short}, ` : "    "
    console.log(`  ${shortFlag}--${key.padEnd(12)} ${opt.description}`)
  })
  console.log("\nExamples:")
  console.log("  node find-unused-translations.mjs -v                    # Verbose output")
  console.log("  node find-unused-translations.mjs -n eb2                # Only show eb2 namespace")
  console.log("  node find-unused-translations.mjs -o json               # Output as JSON")
  console.log(
    "  node find-unused-translations.mjs -t 50                 # Show namespaces with <50% usage"
  )
  process.exit(0)
}

const LOCALES_DIR = path.join(__dirname, "../app/i18n/locales")
const SOURCE_DIR = path.join(__dirname, "..")
const SCAN_DIRS = ["app", "functions", "plugins"]

function getAllTranslationKeys() {
  const allKeys = new Map() // namespace -> Set of keys

  // Get English translations as the baseline
  const enDir = path.join(LOCALES_DIR, "en")
  const files = fs.readdirSync(enDir).filter((f) => f.endsWith(".json"))

  files.forEach((file) => {
    const namespace = path.basename(file, ".json")
    const content = JSON.parse(fs.readFileSync(path.join(enDir, file), "utf-8"))
    const keys = new Set()

    function extractKeys(obj, prefix = "") {
      Object.keys(obj).forEach((key) => {
        const fullKey = prefix ? `${prefix}.${key}` : key
        if (typeof obj[key] === "object" && obj[key] !== null) {
          extractKeys(obj[key], fullKey)
        } else {
          keys.add(fullKey)
        }
      })
    }

    extractKeys(content)
    allKeys.set(namespace, keys)
  })

  return allKeys
}

function findUsedKeys(allTranslationKeys) {
  const usedKeys = new Map() // namespace -> Set of keys
  const dynamicKeys = new Map() // namespace -> Set of partial keys that are dynamic
  const fileStats = { total: 0, scanned: 0, skipped: 0 }
  const scannedFiles = []
  const detectedKeys = [] // Track where keys are found for debug

  // Initialize with empty sets for all known namespaces
  const knownNamespaces = ["common", "eb2", "classic"]
  knownNamespaces.forEach((ns) => {
    usedKeys.set(ns, new Set())
    dynamicKeys.set(ns, new Set())
  })

  function scanDirectory(dir, depth = 0) {
    const files = fs.readdirSync(dir)

    files.forEach((file) => {
      const fullPath = path.join(dir, file)
      const stat = fs.statSync(fullPath)

      fileStats.total++

      if (stat.isDirectory()) {
        // Skip node_modules, build directories, and hidden directories
        if (
          !file.includes("node_modules") &&
          !file.includes("build") &&
          !file.includes("dist") &&
          !file.startsWith(".")
        ) {
          if (args.debug && depth < 3) {
            console.log(`${"  ".repeat(depth)}📁 Scanning directory: ${file}/`)
          }
          scanDirectory(fullPath, depth + 1)
        } else {
          fileStats.skipped++
          if (args.debug && depth < 3) {
            console.log(`${"  ".repeat(depth)}⏩ Skipping directory: ${file}/`)
          }
        }
      } else if (
        file.endsWith(".ts") ||
        file.endsWith(".tsx") ||
        file.endsWith(".js") ||
        file.endsWith(".jsx")
      ) {
        fileStats.scanned++
        scannedFiles.push(fullPath.replace(SOURCE_DIR + "/", ""))
        if (args.debug && depth < 3) {
          console.log(`${"  ".repeat(depth)}📄 Scanning file: ${file}`)
        }
        const content = fs.readFileSync(fullPath, "utf-8")

        // Find namespace from useTranslation hook and i18n.use
        const namespaceMatches = [
          ...content.matchAll(/useTranslation\s*\(\s*["']([^"']+)["']\s*\)/g),
          ...content.matchAll(/i18n\.use\s*\(\s*["']([^"']+)["']\s*\)/g),
          ...content.matchAll(/withTranslation\s*\(\s*["']([^"']+)["']\s*\)/g),
        ]
        const namespaces = new Map()
        const conditionalNamespaces = new Set() // Track namespaces used in conditions

        // Check for conditional namespace selection
        const conditionalMatches = [
          ...content.matchAll(
            /useTranslation\s*\(\s*[^)]*\?\s*["']([^"']+)["']\s*:\s*["']([^"']+)["']/g
          ),
        ]
        conditionalMatches.forEach((match) => {
          conditionalNamespaces.add(match[1])
          conditionalNamespaces.add(match[2])
        })

        // Map variable names to namespaces
        namespaceMatches.forEach((match) => {
          const namespace = match[1]
          // Find the variable name this is assigned to
          const lineMatch = content.slice(0, match.index).lastIndexOf("\n")
          const line = content.slice(lineMatch, match.index + match[0].length + 100)

          // Match patterns like: const { t } = useTranslation("namespace")
          const varMatch = line.match(
            /(?:const|let|var)\s+\{\s*t(?:\s*:\s*(\w+))?\s*\}\s*=\s*useTranslation/
          )
          if (varMatch) {
            const varName = varMatch[1] || "t"
            namespaces.set(varName, namespace)
          }

          // Also match patterns like: const t = useTranslation("namespace")
          const directMatch = line.match(/(?:const|let|var)\s+(\w+)\s*=\s*useTranslation/)
          if (directMatch) {
            namespaces.set(directMatch[1], namespace)
          }

          // Match i18n.t patterns
          if (line.includes("i18n")) {
            namespaces.set("i18n.t", namespace)
          }
        })

        // Default namespace for 't' if not explicitly set
        if (!namespaces.has("t") && namespaces.size > 0) {
          // If there's only one namespace and no alias, 't' probably refers to it
          const entries = Array.from(namespaces.entries())
          if (entries.length === 1 && !entries[0][0].includes("t")) {
            namespaces.set("t", entries[0][1])
          }
        }

        // Find all translation key usages with improved patterns
        const patterns = [
          // Basic t() calls
          /\bt\s*\(\s*["']([^"']+)["']/g,
          /\bt\s*\(\s*`([^`]+)`/g,
          // Named variables like t_eb2, commonT, etc.
          /\b(\w+T|t_\w+)\s*\(\s*["']([^"']+)["']/g,
          /\b(\w+T|t_\w+)\s*\(\s*`([^`]+)`/g,
          // Trans component from react-i18next
          /<Trans[^>]*i18nKey=["']([^"']+)["']/g,
          // t function with options object
          /\bt\s*\(\s*["']([^"']+)["']\s*,\s*\{/g,
          // i18n.t() calls
          /i18n\.t\s*\(\s*["']([^"']+)["']/g,
          /i18next\.t\s*\(\s*["']([^"']+)["']/g,
          // Translation in JSX attributes
          /(?:label|title|placeholder|alt|aria-label|description|helpText|errorMessage)=\{t\s*\(\s*["']([^"']+)["']\s*\)\}/g,
          // Translation with namespace prefix
          /\bt\s*\(\s*["']([\w]+):([^"']+)["']/g,
          // Server-side translations (Remix loaders/actions)
          /i18next\.getFixedT\s*\([^,]+,\s*["']([^"']+)["']/g,
          /t\.get\s*\(\s*["']([^"']+)["']/g,
          // Object property translations
          /["']?(?:label|title|name|description|placeholder)["']?\s*:\s*t\s*\(\s*["']([^"']+)["']/g,
          // Server-side translations via parentData.translations
          /translations\??\.\w+\??\.([\w.]+)/g,
          // getFixedT patterns
          /getFixedT\s*\([^,)]+(?:,\s*["']([^"']+)["'])?\)/g,
        ]

        // Special handling for returnObjects: true pattern
        const returnObjectsMatches = content.matchAll(
          /t\s*\(\s*["']([^"']+)["']\s*,\s*\{[^}]*returnObjects\s*:\s*true/g
        )
        for (const match of returnObjectsMatches) {
          const parentKey = match[1]
          // Find the namespace for this usage
          let namespace = namespaces.get("t") || namespaces.values().next().value

          if (namespace && usedKeys.has(namespace)) {
            // Mark all child keys of this parent as used
            const allNamespaceKeys = allTranslationKeys.get(namespace) || new Set()
            for (const key of allNamespaceKeys) {
              if (key.startsWith(parentKey + ".")) {
                usedKeys.get(namespace).add(key)
                if (args.debug) {
                  detectedKeys.push({
                    namespace,
                    key,
                    file: fullPath.replace(SOURCE_DIR + "/", ""),
                    note: "via returnObjects",
                  })
                }
              }
            }
            // Also add the parent key itself
            usedKeys.get(namespace).add(parentKey)
          }
        }

        patterns.forEach((pattern) => {
          const matches = [...content.matchAll(pattern)]
          matches.forEach((match) => {
            let key, funcName

            if (match.length === 2) {
              // Simple t() pattern
              funcName = "t"
              key = match[1]
            } else {
              // Named function pattern
              funcName = match[1]
              key = match[2]
            }

            // Handle dynamic keys and template literals
            if (key) {
              // Determine namespace first
              let namespace = namespaces.get(funcName)

              if (!namespace) {
                // Try to infer from function name
                if (funcName === "t_eb2" || funcName === "tEb2" || funcName === "eb2T") {
                  namespace = "eb2"
                } else if (
                  funcName === "commonT" ||
                  funcName === "t_common" ||
                  funcName === "tCommon"
                ) {
                  namespace = "common"
                } else if (
                  funcName === "classicT" ||
                  funcName === "t_classic" ||
                  funcName === "tClassic"
                ) {
                  namespace = "classic"
                } else if (funcName === "t" && namespaces.size === 1) {
                  namespace = namespaces.values().next().value
                } else if (funcName === "i18n.t" && namespaces.has("i18n.t")) {
                  namespace = namespaces.get("i18n.t")
                }

                // Handle namespace:key pattern
                if (key && key.includes(":")) {
                  const [ns, actualKey] = key.split(":")
                  const knownNamespaces = ["common", "eb2", "classic"]
                  if (knownNamespaces.includes(ns)) {
                    namespace = ns
                    key = actualKey
                  }
                }
              }

              // Check if it's a dynamic key
              if (key.includes("${")) {
                // Extract the static part of dynamic keys
                const staticPart = key.split("${")[0]
                if (staticPart && namespace) {
                  dynamicKeys.get(namespace)?.add(staticPart)
                }
                return // Skip further processing for this key
              }

              if (namespace && usedKeys.has(namespace)) {
                usedKeys.get(namespace).add(key)
                if (args.debug) {
                  detectedKeys.push({
                    namespace,
                    key,
                    file: fullPath.replace(SOURCE_DIR + "/", ""),
                  })
                }
              }

              // For conditional namespaces, add key to all possible namespaces
              if (!namespace && conditionalNamespaces.size > 0 && key) {
                conditionalNamespaces.forEach((ns) => {
                  if (usedKeys.has(ns)) {
                    usedKeys.get(ns).add(key)
                    if (args.debug) {
                      detectedKeys.push({
                        namespace: ns,
                        key,
                        file: fullPath.replace(SOURCE_DIR + "/", ""),
                        note: "conditional namespace",
                      })
                    }
                  }
                })
              }
            }
          })
        })

        // Special handling for server-side meta function patterns
        if (content.includes("translations") && content.includes("meta")) {
          // Look for patterns like translations?.eb2?.getStarted?.meta
          const metaMatches = content.matchAll(/translations\??\.([\w]+)\??\.([\w]+)\??\.meta/g)
          for (const match of metaMatches) {
            const namespace = match[1] // e.g., 'eb2'
            const parentKey = match[2] // e.g., 'getStarted'

            if (usedKeys.has(namespace)) {
              // Mark meta.title and meta.description as used
              usedKeys.get(namespace).add(`${parentKey}.meta.title`)
              usedKeys.get(namespace).add(`${parentKey}.meta.description`)
              if (args.debug) {
                detectedKeys.push({
                  namespace,
                  key: `${parentKey}.meta.title`,
                  file: fullPath.replace(SOURCE_DIR + "/", ""),
                  note: "server meta function",
                })
                detectedKeys.push({
                  namespace,
                  key: `${parentKey}.meta.description`,
                  file: fullPath.replace(SOURCE_DIR + "/", ""),
                  note: "server meta function",
                })
              }
            }
          }
        }
      } else {
        fileStats.skipped++
      }
    })
  }

  if (args.debug) {
    console.log("\n🔍 Starting file scan...")
    console.log(`Base directory: ${SOURCE_DIR}`)
    console.log(`Scanning directories: ${SCAN_DIRS.join(", ")}\n`)
  }

  // Scan multiple directories
  SCAN_DIRS.forEach((dir) => {
    const fullDir = path.join(SOURCE_DIR, dir)
    if (fs.existsSync(fullDir)) {
      if (args.debug) {
        console.log(`\n📂 Scanning ${dir}/...`)
      }
      scanDirectory(fullDir)
    }
  })

  if (args.debug) {
    console.log("\n📊 File Scanning Summary:")
    console.log(`  Total items found: ${fileStats.total}`)
    console.log(`  Files scanned: ${fileStats.scanned}`)
    console.log(`  Items skipped: ${fileStats.skipped}`)
    console.log("\n📝 Sample of scanned files (first 10):")
    scannedFiles.slice(0, 10).forEach((file) => {
      console.log(`    - ${file}`)
    })
    if (scannedFiles.length > 10) {
      console.log(`    ... and ${scannedFiles.length - 10} more files`)
    }
  }

  // Log some statistics for debugging
  console.log("\n🔑 Detection Statistics:")
  usedKeys.forEach((keys, namespace) => {
    console.log(`  ${namespace}: ${keys.size} unique keys detected`)
  })

  if (args.debug && detectedKeys.length > 0) {
    console.log("\n📍 Sample of detected keys (first 20):")
    detectedKeys.slice(0, 20).forEach(({ namespace, key, file }) => {
      console.log(`  [${namespace}] ${key} in ${file}`)
    })
    if (detectedKeys.length > 20) {
      console.log(`  ... and ${detectedKeys.length - 20} more detections`)
    }
  }

  if (Array.from(dynamicKeys.values()).some((set) => set.size > 0)) {
    console.log("\nDynamic key patterns found:")
    dynamicKeys.forEach((patterns, namespace) => {
      if (patterns.size > 0) {
        console.log(`  ${namespace}: ${patterns.size} patterns`)
      }
    })
  }

  return { usedKeys, dynamicKeys }
}

function analyzeUnusedKeys() {
  const allKeys = getAllTranslationKeys()
  const { usedKeys, dynamicKeys } = findUsedKeys(allKeys)

  const threshold = parseFloat(args.threshold) || 0
  const filterNamespace = args.namespace

  console.log("=".repeat(60))
  console.log("Translation Keys Analysis Report")
  console.log("=".repeat(60))

  let totalKeys = 0
  let totalUsed = 0
  let totalUnused = 0

  const unusedByNamespace = new Map()

  allKeys.forEach((keys, namespace) => {
    // Skip if filtering by namespace
    if (filterNamespace && namespace !== filterNamespace) {
      return
    }

    const used = usedKeys.get(namespace) || new Set()
    const unused = new Set()

    keys.forEach((key) => {
      if (!used.has(key)) {
        // Check if this might be a dynamic key
        let isDynamic = false
        const dynamicPatterns = dynamicKeys.get(namespace) || new Set()
        for (const pattern of dynamicPatterns) {
          if (key.startsWith(pattern)) {
            isDynamic = true
            break
          }
        }

        if (!isDynamic) {
          unused.add(key)
        }
      }
    })

    totalKeys += keys.size
    totalUsed += used.size
    totalUnused += unused.size

    const usagePercent = Math.round((used.size / keys.size) * 100)

    // Skip if above threshold
    if (threshold > 0 && usagePercent > threshold) {
      return
    }

    if (args.output === "console") {
      console.log(`\nNamespace: ${namespace}`)
      console.log(`  Total keys: ${keys.size}`)
      console.log(`  Used keys: ${used.size}`)
      console.log(`  Unused keys: ${unused.size}`)
      console.log(`  Usage: ${usagePercent}%`)
    }

    if (unused.size > 0) {
      unusedByNamespace.set(namespace, unused)
    }
  })

  if (args.output === "console") {
    console.log("\n" + "=".repeat(60))
    console.log("Summary:")
    console.log("=".repeat(60))
    console.log(`Total translation keys: ${totalKeys}`)
    console.log(`Used keys: ${totalUsed}`)
    console.log(`Unused keys: ${totalUnused}`)
    console.log(`Overall usage: ${Math.round((totalUsed / totalKeys) * 100)}%`)
  }

  // Show dynamic key patterns if any
  let hasDynamicKeys = false
  dynamicKeys.forEach((patterns, namespace) => {
    if (patterns.size > 0) {
      if (!hasDynamicKeys) {
        console.log("\n" + "=".repeat(60))
        console.log("Dynamic Key Patterns Detected:")
        console.log("=".repeat(60))
        hasDynamicKeys = true
      }
      console.log(`\n${namespace}:`)
      patterns.forEach((pattern) => {
        console.log(`  - ${pattern}*`)
      })
    }
  })

  if (totalUnused > 0 && args.output === "console") {
    console.log("\n" + "=".repeat(60))
    console.log("Unused Translation Keys by Namespace:")
    console.log("=".repeat(60))

    unusedByNamespace.forEach((keys, namespace) => {
      console.log(`\n${namespace}: (${keys.size} unused keys)`)
      const sortedKeys = Array.from(keys).sort()

      if (args.verbose) {
        // Show all keys in verbose mode
        sortedKeys.forEach((key) => {
          console.log(`  - ${key}`)
        })
      } else {
        // Show first 10 keys as examples
        const displayKeys = sortedKeys.slice(0, 10)
        displayKeys.forEach((key) => {
          console.log(`  - ${key}`)
        })

        if (sortedKeys.length > 10) {
          console.log(`  ... and ${sortedKeys.length - 10} more`)
        }
      }
    })
  }

  // Save full report to file
  const report = {
    timestamp: new Date().toISOString(),
    summary: {
      totalKeys,
      usedKeys: totalUsed,
      unusedKeys: totalUnused,
      usagePercentage: Math.round((totalUsed / totalKeys) * 100),
    },
    byNamespace: {},
  }

  unusedByNamespace.forEach((keys, namespace) => {
    report.byNamespace[namespace] = {
      total: allKeys.get(namespace).size,
      used: usedKeys.get(namespace).size,
      unused: keys.size,
      unusedKeys: Array.from(keys).sort(),
    }
  })

  // Handle different output formats
  if (args.output === "json") {
    console.log(JSON.stringify(report, null, 2))
  } else if (args.output === "csv") {
    console.log("Namespace,Key,Status")
    allKeys.forEach((keys, namespace) => {
      const used = usedKeys.get(namespace) || new Set()
      keys.forEach((key) => {
        const status = used.has(key) ? "used" : "unused"
        console.log(`${namespace},${key},${status}`)
      })
    })
  }
  // File creation disabled - script will only output to console
}

// Run the analysis
analyzeUnusedKeys()

#!/usr/bin/env node

import fs from "fs"
import path from "path"
import { fileURLToPath } from "url"

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const LOCALES_DIR = path.join(__dirname, "../app/i18n/locales")
const SOURCE_DIR = path.join(__dirname, "../app")

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

function findUsedKeys() {
  const usedKeys = new Map() // namespace -> Set of keys

  // Initialize with empty sets
  usedKeys.set("common", new Set())
  usedKeys.set("eb2", new Set())
  usedKeys.set("classic", new Set())

  function scanDirectory(dir) {
    const files = fs.readdirSync(dir)

    files.forEach((file) => {
      const fullPath = path.join(dir, file)
      const stat = fs.statSync(fullPath)

      if (stat.isDirectory()) {
        // Skip node_modules and build directories
        if (!file.includes("node_modules") && !file.includes("build") && !file.includes(".")) {
          scanDirectory(fullPath)
        }
      } else if (
        file.endsWith(".ts") ||
        file.endsWith(".tsx") ||
        file.endsWith(".js") ||
        file.endsWith(".jsx")
      ) {
        const content = fs.readFileSync(fullPath, "utf-8")

        // Find namespace from useTranslation hook
        const namespaceMatches = [
          ...content.matchAll(/useTranslation\s*\(\s*["']([^"']+)["']\s*\)/g),
        ]
        const namespaces = new Map()

        // Map variable names to namespaces
        namespaceMatches.forEach((match) => {
          const namespace = match[1]
          // Find the variable name this is assigned to
          const lineMatch = content.slice(0, match.index).lastIndexOf("\n")
          const line = content.slice(lineMatch, match.index + match[0].length + 100)

          // Match patterns like: const { t } = useTranslation("namespace")
          const varMatch = line.match(/const\s+\{\s*t(?:\s*:\s*(\w+))?\s*\}\s*=\s*useTranslation/)
          if (varMatch) {
            const varName = varMatch[1] || "t"
            namespaces.set(varName, namespace)
          }

          // Also match patterns like: const t = useTranslation("namespace")
          const directMatch = line.match(/const\s+(\w+)\s*=\s*useTranslation/)
          if (directMatch) {
            namespaces.set(directMatch[1], namespace)
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

        // Find all translation key usages
        // Pattern 1: t("key") or t('key') or t(`key`)
        const patterns = [
          /\bt\s*\(\s*["']([^"']+)["']/g,
          /\bt\s*\(\s*`([^`]+)`/g,
          // Named variables like t_eb2, commonT, etc.
          /\b(\w+T|\w+_\w+)\s*\(\s*["']([^"']+)["']/g,
          /\b(\w+T|\w+_\w+)\s*\(\s*`([^`]+)`/g,
        ]

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

            // Skip if key contains ${} (dynamic keys)
            if (key && !key.includes("${")) {
              // Determine namespace
              let namespace = namespaces.get(funcName)

              if (!namespace) {
                // Try to infer from function name
                if (funcName === "t_eb2" || funcName === "tEb2") {
                  namespace = "eb2"
                } else if (funcName === "commonT" || funcName === "t_common") {
                  namespace = "common"
                } else if (funcName === "classicT" || funcName === "t_classic") {
                  namespace = "classic"
                } else if (funcName === "t" && namespaces.size === 1) {
                  namespace = namespaces.values().next().value
                }
              }

              if (namespace && usedKeys.has(namespace)) {
                usedKeys.get(namespace).add(key)
              }
            }
          })
        })
      }
    })
  }

  scanDirectory(SOURCE_DIR)

  return usedKeys
}

function analyzeUnusedKeys() {
  const allKeys = getAllTranslationKeys()
  const usedKeys = findUsedKeys()

  console.log("=".repeat(60))
  console.log("Translation Keys Analysis Report")
  console.log("=".repeat(60))

  let totalKeys = 0
  let totalUsed = 0
  let totalUnused = 0

  const unusedByNamespace = new Map()

  allKeys.forEach((keys, namespace) => {
    const used = usedKeys.get(namespace) || new Set()
    const unused = new Set()

    keys.forEach((key) => {
      if (!used.has(key)) {
        unused.add(key)
      }
    })

    totalKeys += keys.size
    totalUsed += used.size
    totalUnused += unused.size

    console.log(`\nNamespace: ${namespace}`)
    console.log(`  Total keys: ${keys.size}`)
    console.log(`  Used keys: ${used.size}`)
    console.log(`  Unused keys: ${unused.size}`)
    console.log(`  Usage: ${Math.round((used.size / keys.size) * 100)}%`)

    if (unused.size > 0) {
      unusedByNamespace.set(namespace, unused)
    }
  })

  console.log("\n" + "=".repeat(60))
  console.log("Summary:")
  console.log("=".repeat(60))
  console.log(`Total translation keys: ${totalKeys}`)
  console.log(`Used keys: ${totalUsed}`)
  console.log(`Unused keys: ${totalUnused}`)
  console.log(`Overall usage: ${Math.round((totalUsed / totalKeys) * 100)}%`)

  if (totalUnused > 0) {
    console.log("\n" + "=".repeat(60))
    console.log("Unused Translation Keys by Namespace:")
    console.log("=".repeat(60))

    unusedByNamespace.forEach((keys, namespace) => {
      console.log(`\n${namespace}: (${keys.size} unused keys)`)
      const sortedKeys = Array.from(keys).sort()

      // Show first 10 keys as examples
      const displayKeys = sortedKeys.slice(0, 10)
      displayKeys.forEach((key) => {
        console.log(`  - ${key}`)
      })

      if (sortedKeys.length > 10) {
        console.log(`  ... and ${sortedKeys.length - 10} more`)
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

  const reportPath = path.join(__dirname, "unused-translations-report.json")
  fs.writeFileSync(reportPath, JSON.stringify(report, null, 2))
  console.log(`\nFull report saved to: ${reportPath}`)

  // Also create a file with just the unused keys for easy removal
  const unusedKeysPath = path.join(__dirname, "unused-keys.txt")
  let unusedContent = ""
  unusedByNamespace.forEach((keys, namespace) => {
    unusedContent += `# ${namespace}\n`
    Array.from(keys)
      .sort()
      .forEach((key) => {
        unusedContent += `${key}\n`
      })
    unusedContent += "\n"
  })
  fs.writeFileSync(unusedKeysPath, unusedContent)
  console.log(`Unused keys list saved to: ${unusedKeysPath}`)
}

// Run the analysis
analyzeUnusedKeys()

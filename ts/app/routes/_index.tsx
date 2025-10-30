import type { LoaderFunctionArgs, MetaFunction } from "@remix-run/cloudflare"
import HeroSection from "~/components/landing/HeroSection"
import FeaturesSection from "~/components/landing/FeaturesSection"
import PricingSection from "~/components/landing/PricingSection"
import { i18nLoader } from "~/i18n/root-loader"

import translationsEN from "~/i18n/locales/en/translations.json"
import translationsRU from "~/i18n/locales/ru/translations.json"
import translationsTW from "~/i18n/locales/tw/translations.json"
import translationsKO from "~/i18n/locales/ko/translations.json"
import translationsPTBR from "~/i18n/locales/pt-BR/translations.json"

interface TranslationsMeta {
  meta: {
    title: string
    description: string
    keywords?: string
  }
}

const translationsMap: Record<string, TranslationsMeta> = {
  en: translationsEN as TranslationsMeta,
  ru: translationsRU as TranslationsMeta,
  tw: translationsTW as TranslationsMeta,
  ko: translationsKO as TranslationsMeta,
  "pt-BR": translationsPTBR as TranslationsMeta,
}

export async function loader({ request, params, context }: LoaderFunctionArgs) {
  const i18nData = await i18nLoader({ request, params, context })
  const locale = i18nData.locale || "en"

  // Get translation data for meta tags
  const translations = translationsMap[locale] || (translationsEN as TranslationsMeta)

  return {
    locale,
    meta: translations.meta,
  }
}

export const meta: MetaFunction<typeof loader> = ({ data }) => {
  const metaData = data?.meta || {
    title: "RaidBot - Best bot for RAID: Shadow Legends",
    description: "RaidBot is your ultimate assistant for RAID: Shadow Legends. Automate farming, arena battles, and more with human-like behavior simulation.",
  }

  return [
    { title: metaData.title },
    {
      name: "description",
      content: metaData.description,
    },
    { property: "og:title", content: metaData.title },
    {
      property: "og:description",
      content: metaData.description,
    },
    { property: "og:type", content: "website" },
    {
      name: "keywords",
      content: metaData.keywords || "",
    },
  ]
}

export default function Index() {
  return (
    <>
      <HeroSection />
      <FeaturesSection />
      <PricingSection />
    </>
  )
}

import plugin from "tailwindcss/plugin"

const FONT_WEIGHTS = {
  thin: 100,
  extralight: 200,
  light: 300,
  normal: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
  extrabold: 800,
  black: 900,
}

// eslint-disable-next-line @typescript-eslint/unbound-method -- TailwindCSS plugin API pattern
export const fontVariationSettings = plugin(({ addUtilities }) => {
  const utilities = Object.entries(FONT_WEIGHTS).reduce(
    (utils, [key, value]) => {
      return {
        ...utils,
        [`.font-${key}`]: {
          fontWeight: value,
          fontVariationSettings: `"wght" ${value}`,
        },
        [`.font-${key}.italic`]: {
          fontWeight: value,
          fontVariationSettings: `"slnt" 1, "wght" ${value}`,
          fontStyle: "italic",
        },
      }
    },
    {
      ".italic": {
        fontStyle: "italic",
        fontVariationSettings: '"slnt" 1',
      },
    }
  )

  addUtilities(utilities)
})

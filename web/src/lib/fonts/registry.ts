// Self-hosted latin subsets (OFL-1.1, Fontsource). Offline builds do not call Google.
import localFont from "next/font/local";

import { GeistPixelSquare } from "geist/font/pixel";

const fontOpts = { display: "swap", adjustFontFallback: false, weight: "100 900" } as const;

const geist = localFont({
  src: "../../../node_modules/geist/dist/fonts/geist-sans/Geist-Variable.woff2",
  variable: "--font-geist",
  ...fontOpts,
});
const geistMono = localFont({
  src: "../../../node_modules/geist/dist/fonts/geist-mono/GeistMono-Variable.woff2",
  variable: "--font-geist-mono",
  ...fontOpts,
});
const inter = localFont({ src: "./files/inter-latin-wght-normal.woff2", variable: "--font-inter", ...fontOpts });
const notoSans = localFont({
  src: "./files/noto-sans-latin-wght-normal.woff2",
  variable: "--font-noto-sans",
  ...fontOpts,
});
const nunitoSans = localFont({
  src: "./files/nunito-sans-latin-wght-normal.woff2",
  variable: "--font-nunito-sans",
  ...fontOpts,
});
const figtree = localFont({ src: "./files/figtree-latin-wght-normal.woff2", variable: "--font-figtree", ...fontOpts });
const roboto = localFont({ src: "./files/roboto-latin-wght-normal.woff2", variable: "--font-roboto", ...fontOpts });
const raleway = localFont({ src: "./files/raleway-latin-wght-normal.woff2", variable: "--font-raleway", ...fontOpts });
const dmSans = localFont({ src: "./files/dm-sans-latin-wght-normal.woff2", variable: "--font-dm-sans", ...fontOpts });
const publicSans = localFont({
  src: "./files/public-sans-latin-wght-normal.woff2",
  variable: "--font-public-sans",
  ...fontOpts,
});
const outfit = localFont({ src: "./files/outfit-latin-wght-normal.woff2", variable: "--font-outfit", ...fontOpts });
const jetBrainsMono = localFont({
  src: "./files/jetbrains-mono-latin-wght-normal.woff2",
  variable: "--font-jetbrains-mono",
  ...fontOpts,
});
const notoSerif = localFont({
  src: "./files/noto-serif-latin-wght-normal.woff2",
  variable: "--font-noto-serif",
  ...fontOpts,
});
const robotoSlab = localFont({
  src: "./files/roboto-slab-latin-wght-normal.woff2",
  variable: "--font-roboto-slab",
  ...fontOpts,
});
const merriweather = localFont({
  src: "./files/merriweather-latin-wght-normal.woff2",
  variable: "--font-merriweather",
  ...fontOpts,
});
const lora = localFont({ src: "./files/lora-latin-wght-normal.woff2", variable: "--font-lora", ...fontOpts });
const playfairDisplay = localFont({
  src: "./files/playfair-display-latin-wght-normal.woff2",
  variable: "--font-playfair-display",
  ...fontOpts,
});

export const fontRegistry = {
  geist: { label: "Geist", font: geist },
  inter: { label: "Inter", font: inter },
  notoSans: { label: "Noto Sans", font: notoSans },
  nunitoSans: { label: "Nunito Sans", font: nunitoSans },
  figtree: { label: "Figtree", font: figtree },
  roboto: { label: "Roboto", font: roboto },
  raleway: { label: "Raleway", font: raleway },
  dmSans: { label: "DM Sans", font: dmSans },
  publicSans: { label: "Public Sans", font: publicSans },
  outfit: { label: "Outfit", font: outfit },
  geistMono: { label: "Geist Mono", font: geistMono },
  geistPixelSquare: { label: "Geist Pixel Square", font: GeistPixelSquare },
  jetBrainsMono: { label: "JetBrains Mono", font: jetBrainsMono },
  notoSerif: { label: "Noto Serif", font: notoSerif },
  robotoSlab: { label: "Roboto Slab", font: robotoSlab },
  merriweather: { label: "Merriweather", font: merriweather },
  lora: { label: "Lora", font: lora },
  playfairDisplay: { label: "Playfair Display", font: playfairDisplay },
} as const;

export type FontKey = keyof typeof fontRegistry;

export const fontVars = (Object.values(fontRegistry) as Array<(typeof fontRegistry)[FontKey]>)
  .map((f) => f.font.variable)
  .join(" ");

export const fontOptions = (
  Object.entries(fontRegistry) as [FontKey, { label: string; font: { variable: string } }][]
).map(([key, f]) => ({
  key,
  label: f.label,
  variable: f.font.variable,
}));

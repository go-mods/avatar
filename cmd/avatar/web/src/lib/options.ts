// An empty number or text means that the server chooses the value.
export type NumberOption = number | ''

export const shapes = ['circle', 'square', 'diamond', 'octagon', 'heart', 'cloud'] as const

export interface AvatarOptions {
  name: string
  separator: string
  length: NumberOption
  lowercase: boolean
  uppercase: boolean
  camelcase: boolean
  sensitive: boolean
  wordLength: boolean
  width: NumberOption
  height: NumberOption
  shape: (typeof shapes)[number]
  backgroundColor: string
  fontColor: string
  borderColor: string
  randomColor: boolean
  randomBackgroundColor: boolean
  randomFontColor: boolean
  randomBorderColor: boolean
  fontFamily: string
  fontWeight: NumberOption
  fontSize: NumberOption
  borderWidth: NumberOption
  borderRadius: NumberOption
  borderDash: string
  padding: NumberOption
}

export function defaultOptions(): AvatarOptions {
  return {
    name: 'Ada Lovelace',
    separator: '',
    length: '',
    lowercase: false,
    uppercase: false,
    camelcase: false,
    sensitive: false,
    wordLength: false,
    width: '',
    height: '',
    shape: 'circle',
    backgroundColor: '',
    fontColor: '',
    borderColor: '',
    randomColor: false,
    randomBackgroundColor: false,
    randomFontColor: false,
    randomBorderColor: false,
    fontFamily: '',
    fontWeight: '',
    fontSize: '',
    borderWidth: '',
    borderRadius: '',
    borderDash: '',
    padding: '',
  }
}

const textKeys = ['name', 'separator', 'backgroundColor', 'fontColor', 'borderColor', 'fontFamily', 'borderDash'] as const
const numberKeys = ['length', 'width', 'height', 'fontWeight', 'fontSize', 'borderWidth', 'borderRadius', 'padding'] as const
const booleanKeys = [
  'lowercase',
  'uppercase',
  'camelcase',
  'sensitive',
  'wordLength',
  'randomColor',
  'randomBackgroundColor',
  'randomFontColor',
  'randomBorderColor',
] as const

// buildQuery returns the query of the request, with the options which are not the default ones.
export function buildQuery(options: AvatarOptions): string {
  const query = new URLSearchParams()
  for (const key of textKeys) {
    if (options[key] !== '') query.set(key, options[key])
  }
  for (const key of numberKeys) {
    if (options[key] !== '') query.set(key, String(options[key]))
  }
  for (const key of booleanKeys) {
    if (options[key]) query.set(key, 'true')
  }
  if (options.shape !== 'circle') query.set('shape', options.shape)
  if (options.shape !== 'square') query.delete('borderRadius')
  return query.toString()
}

// parseQuery returns the options of a shared link: what cannot be read keeps its default value.
export function parseQuery(search: string): AvatarOptions {
  const query = new URLSearchParams(search)
  const options = defaultOptions()
  for (const key of textKeys) {
    const value = query.get(key)
    if (value !== null) options[key] = value
  }
  for (const key of numberKeys) {
    const value = Number(query.get(key) ?? '')
    if (query.get(key) && Number.isInteger(value)) options[key] = value
  }
  for (const key of booleanKeys) {
    options[key] = query.get(key) === 'true'
  }
  options.shape = shapes.find((shape) => shape === query.get('shape')) ?? options.shape
  if (options.fontWeight === 400) options.fontWeight = ''
  return options
}

// goCode returns the Go code which generates the same avatar.
export function goCode(options: AvatarOptions): string {
  const text = (value: string) => JSON.stringify(value)

  const initials: string[] = []
  if (options.separator !== '') initials.push(`initials.WithSeparator(${text(options.separator)})`)
  if (options.length !== '') initials.push(`initials.WithLength(${options.length})`)
  if (options.lowercase) initials.push('initials.WithLowercase()')
  if (options.uppercase) initials.push('initials.WithUppercase()')
  if (options.camelcase) initials.push('initials.WithCamelCase()')
  if (options.sensitive) initials.push('initials.WithSensitive()')
  if (options.wordLength) initials.push('initials.WithWordLength()')

  const avatar: string[] = []
  if (options.shape !== 'circle') avatar.push(`avatar.WithShape(${text(options.shape)})`)
  if (options.width !== '') avatar.push(`avatar.WithWidth(${options.width})`)
  if (options.height !== '') avatar.push(`avatar.WithHeight(${options.height})`)
  if (options.backgroundColor !== '') avatar.push(`avatar.WithBackgroundColor(${text(options.backgroundColor)})`)
  if (options.fontColor !== '') avatar.push(`avatar.WithFontColor(${text(options.fontColor)})`)
  if (options.randomColor) avatar.push('avatar.WithRandomColor()')
  if (options.randomBackgroundColor) avatar.push('avatar.WithRandomBackgroundColor()')
  if (options.randomFontColor) avatar.push('avatar.WithRandomFontColor()')
  if (options.randomBorderColor) avatar.push('avatar.WithRandomBorderColor()')
  if (options.fontFamily !== '' || options.fontWeight !== '') avatar.push(`avatar.WithFont(${text(options.fontFamily)}, ${options.fontWeight === '' ? 400 : options.fontWeight})`)
  if (options.fontSize !== '') avatar.push(`avatar.WithFontSize(${options.fontSize})`)
  if (options.borderColor !== '') avatar.push(`avatar.WithBorderColor(${text(options.borderColor)})`)
  if (options.borderWidth !== '') avatar.push(`avatar.WithBorderWidth(${options.borderWidth})`)
  if (options.shape === 'square' && options.borderRadius !== '') avatar.push(`avatar.WithBorderRadius(${options.borderRadius})`)
  if (options.borderDash !== '') avatar.push(`avatar.WithBorderDash(${text(options.borderDash)})`)
  if (options.padding !== '') avatar.push(`avatar.WithPadding(${options.padding})`)

  const lines = (values: string[], indent: string) => values.map((value) => `${indent}${value},\n`).join('')
  const nameCall =
    initials.length === 0
      ? `initials.GetInitials(${text(options.name)})`
      : `initials.GetInitials(${text(options.name)},\n${lines(initials, '\t\t')}\t)`

  if (avatar.length === 0 && initials.length === 0) return `svg, err := avatar.GetSVG(${nameCall})`
  return `svg, err := avatar.GetSVG(\n\t${nameCall},\n${lines(avatar, '\t')})`
}

// svgToDataUrl returns the url which displays an svg in an image, the way an application without
// server does: an svg displayed as an image loads nothing from outside.
export function svgToDataUrl(svg: string): string {
  const bytes = new TextEncoder().encode(svg)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return 'data:image/svg+xml;base64,' + btoa(binary)
}

// readBackgroundColor returns the color which fills the shape of an svg.
export function readBackgroundColor(svg: string): string | null {
  return /<(?:circle|rect|polygon|path)\b[^>]*\sfill="(#[0-9a-fA-F]{6})"/.exec(svg)?.[1] ?? null
}

function luminance(hex: string): number {
  const channel = (start: number) => {
    const value = parseInt(hex.slice(start, start + 2), 16) / 255
    return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4
  }
  return 0.2126 * channel(1) + 0.7152 * channel(3) + 0.0722 * channel(5)
}

// readableInkOn returns the text color with the highest contrast on a background.
export function readableInkOn(background: string): string {
  const light = luminance(background)
  return (light + 0.05) / 0.0632 > 1.05 / (light + 0.05) ? '#14161a' : '#ffffff'
}

function mix(hex: string, target: number, amount: number): string {
  const channel = (start: number) => {
    const value = parseInt(hex.slice(start, start + 2), 16)
    return Math.round(value + (target - value) * amount)
      .toString(16)
      .padStart(2, '0')
  }
  return `#${channel(1)}${channel(3)}${channel(5)}`
}

// panelColor returns a color of the same family as the background of an avatar, far enough from it
// for the shape to stand out on the panel.
export function panelColor(background: string): string {
  return luminance(background) > 0.75 ? mix(background, 0, 0.18) : mix(background, 255, 0.6)
}

// usesRandomColors tells whether two requests with these options give different avatars.
export function usesRandomColors(options: AvatarOptions): boolean {
  return options.randomColor || options.randomBackgroundColor || options.randomFontColor || options.randomBorderColor
}

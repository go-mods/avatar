import { describe, expect, it } from 'vitest'
import {
  buildQuery,
  defaultOptions,
  goCode,
  panelColor,
  parseQuery,
  readableInkOn,
  readBackgroundColor,
  shapes,
  svgToDataUrl,
  usesRandomColors,
  type AvatarOptions,
} from './options'

function options(changes: Partial<AvatarOptions>): AvatarOptions {
  return { ...defaultOptions(), ...changes }
}

describe('buildQuery', () => {
  it('only sends what differs from the defaults', () => {
    expect(buildQuery(options({}))).toBe('name=Ada+Lovelace')
    expect(buildQuery(options({ name: '' }))).toBe('')
  })

  it('sends every kind of option', () => {
    const query = new URLSearchParams(
      buildQuery(options({ shape: 'square', width: 96, borderRadius: 8, randomColor: true, backgroundColor: '#ff0000', padding: 0 })),
    )
    expect(query.get('shape')).toBe('square')
    expect(query.get('width')).toBe('96')
    expect(query.get('borderRadius')).toBe('8')
    expect(query.get('randomColor')).toBe('true')
    expect(query.get('backgroundColor')).toBe('#ff0000')
    expect(query.get('padding')).toBe('0')
  })

  it('leaves out the options without effect', () => {
    expect(buildQuery(options({ borderRadius: 8 }))).not.toContain('borderRadius')
    expect(buildQuery(options({ shape: 'heart', borderRadius: 8 }))).not.toContain('borderRadius')
    expect(goCode(options({ shape: 'heart', borderRadius: 8 }))).not.toContain('WithBorderRadius')
  })

  it('sends the weight with or without family', () => {
    expect(buildQuery(options({ fontWeight: 900 }))).toContain('fontWeight=900')
    expect(buildQuery(options({ fontFamily: 'Inter', fontWeight: 900 }))).toContain('fontWeight=900')
  })
})

describe('parseQuery', () => {
  it('restores the options of a shared link', () => {
    const shared = options({ name: 'Grace Hopper', shape: 'square', width: 96, borderRadius: 8, lowercase: true, fontFamily: 'Roboto Mono', fontWeight: 700 })
    expect(parseQuery(buildQuery(shared))).toEqual(shared)
  })

  it('restores every shape', () => {
    for (const shape of shapes) {
      expect(parseQuery(buildQuery(options({ shape }))).shape).toBe(shape)
    }
  })

  it('keeps the default of what cannot be read', () => {
    const parsed = parseQuery('width=abc&shape=star&lowercase=maybe&unknown=1')
    expect(parsed.width).toBe('')
    expect(parsed.shape).toBe('circle')
    expect(parsed.lowercase).toBe(false)
  })

  it('reads the regular weight as the default one', () => {
    expect(parseQuery('fontWeight=400').fontWeight).toBe('')
    expect(parseQuery('fontWeight=700').fontWeight).toBe(700)
  })

  it('keeps an empty name', () => {
    expect(parseQuery('name=').name).toBe('')
  })

  it('uses the defaults without query', () => {
    expect(parseQuery('')).toEqual(defaultOptions())
  })
})

describe('goCode', () => {
  it('stays on one line without option', () => {
    expect(goCode(options({}))).toBe('svg, err := avatar.GetSVG(initials.GetInitials("Ada Lovelace"))')
  })

  it('lists the options', () => {
    expect(goCode(options({ shape: 'square', width: 96, fontFamily: 'Inter' }))).toBe(
      'svg, err := avatar.GetSVG(\n\tinitials.GetInitials("Ada Lovelace"),\n\tavatar.WithShape("square"),\n\tavatar.WithWidth(96),\n\tavatar.WithFont("Inter", 400),\n)',
    )
  })

  it('writes the weight without family', () => {
    expect(goCode(options({ fontWeight: 700 }))).toContain('avatar.WithFont("", 700)')
  })

  it('lists the options of the initials', () => {
    expect(goCode(options({ length: 3 }))).toContain('initials.GetInitials("Ada Lovelace",\n\t\tinitials.WithLength(3),\n\t)')
  })

  it('quotes the text', () => {
    expect(goCode(options({ name: 'a"b' }))).toContain('"a\\"b"')
  })
})

describe('svgToDataUrl', () => {
  it('encodes the svg as UTF-8, like the applications do', () => {
    expect(svgToDataUrl('<svg>é你</svg>')).toBe('data:image/svg+xml;base64,PHN2Zz7DqeS9oDwvc3ZnPg==')
  })
})

describe('readBackgroundColor', () => {
  it('reads the fill of the shape, not of the text', () => {
    expect(readBackgroundColor('<circle cx="1" fill="#c4b3a2" stroke="#000000" /><text fill="#8d4806">A</text>')).toBe('#c4b3a2')
    expect(readBackgroundColor('<rect x="0" fill="#123456" />')).toBe('#123456')
    expect(readBackgroundColor('<polygon points="0,0" fill="#123456" />')).toBe('#123456')
    expect(readBackgroundColor('<path d="M0,0" stroke-linejoin="round" fill="#123456" />')).toBe('#123456')
  })

  it('does not take the stroke for a fill', () => {
    expect(readBackgroundColor('<text fill="#8d4806">A</text>')).toBeNull()
  })
})

describe('readableInkOn', () => {
  it('chooses the color with the highest contrast', () => {
    expect(readableInkOn('#ffffff')).toBe('#14161a')
    expect(readableInkOn('#c4b3a2')).toBe('#14161a')
    expect(readableInkOn('#000000')).toBe('#ffffff')
    expect(readableInkOn('#000080')).toBe('#ffffff')
  })
})

describe('panelColor', () => {
  it('lightens a background', () => {
    expect(panelColor('#c4b3a2')).toBe('#e7e1da')
    expect(panelColor('#000000')).toBe('#999999')
  })

  it('darkens a background too light to stand out once lightened', () => {
    expect(panelColor('#ffffff')).toBe('#d1d1d1')
  })

  it('always leaves a difference with the background', () => {
    for (const background of ['#000000', '#191970', '#ff0000', '#00ff00', '#808080', '#ffff00', '#ffffff']) {
      expect(panelColor(background)).not.toBe(background)
    }
  })
})

describe('usesRandomColors', () => {
  it('sees each random option', () => {
    expect(usesRandomColors(options({}))).toBe(false)
    for (const key of ['randomColor', 'randomBackgroundColor', 'randomFontColor', 'randomBorderColor'] as const) {
      expect(usesRandomColors(options({ [key]: true }))).toBe(true)
    }
  })
})

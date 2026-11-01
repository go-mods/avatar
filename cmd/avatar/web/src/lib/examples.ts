import type { AvatarOptions } from './options'

export interface Example {
  caption: string
  options: Partial<AvatarOptions>
}

// The examples of the gallery: each one shows one thing the options can do.
export const examples: Example[] = [
  { caption: 'Colors from the name', options: { name: 'Grace Hopper' } },
  { caption: 'Rounded square', options: { name: 'Alan Turing', shape: 'square', borderRadius: 10 } },
  { caption: 'Dashed border', options: { name: 'Linus Torvalds', borderWidth: 2, borderDash: '4 2', borderColor: '#334155' } },
  { caption: 'Your own colors', options: { name: 'Katherine Johnson', backgroundColor: 'midnightblue', fontColor: 'gold' } },
  { caption: 'A Google font', options: { name: 'Margaret Hamilton', fontFamily: 'Pacifico' } },
  { caption: 'A heavy weight', options: { name: 'Barbara Liskov', fontFamily: 'Inter', fontWeight: 900 } },
  { caption: 'Three letters', options: { name: 'Dennis Ritchie', length: 3, fontSize: 16 } },
  { caption: 'Separated initials', options: { name: 'Edsger Dijkstra', separator: '.', fontSize: 16 } },
]

// The fonts proposed by the form; any other Google Fonts family can be typed.
export const fontFamilies = [
  'Inter',
  'Roboto',
  'Roboto Mono',
  'Lato',
  'Montserrat',
  'Playfair Display',
  'Merriweather',
  'Caprasimo',
  'Pacifico',
  'Rancho',
  'Tangerine',
  'M PLUS 1p',
]

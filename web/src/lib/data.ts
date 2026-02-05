import type { Meta, About, Work, Cv, Links } from '@data/types'

import metaJson from '@data/content/meta.json'
import aboutJson from '@data/content/about.json'
import workJson from '@data/content/work.json'
import cvJson from '@data/content/cv.json'
import linksJson from '@data/content/links.json'

/** Load site metadata. */
export function getMeta(): Meta {
  return metaJson as Meta
}

/** Load about/bio data. */
export function getAbout(): About {
  return aboutJson as About
}

/** Load work/projects data. */
export function getWork(): Work {
  return workJson as Work
}

/** Load CV/resume data. */
export function getCv(): Cv {
  return cvJson as Cv
}

/** Load external links data. */
export function getLinks(): Links {
  return linksJson as Links
}

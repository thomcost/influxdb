import {TimeRange} from 'src/types'

export const CUSTOM_TIME_RANGE = 'Custom Time Range'
export const TIME_RANGE_FORMAT = 'YYYY-MM-DD HH:mm'

export const TIME_RANGES: TimeRange[] = [
  {
    lower: '',
    label: CUSTOM_TIME_RANGE,
  },
  {
    seconds: 300,
    lower: 'now() - 5m',
    upper: null,
    label: 'Past 5m',
    duration: '5m',
  },
  {
    seconds: 900,
    lower: 'now() - 15m',
    upper: null,
    label: 'Past 15m',
    duration: '15m',
  },
  {
    seconds: 3600,
    lower: 'now() - 1h',
    upper: null,
    label: 'Past 1h',
    duration: '1h',
  },
  {
    seconds: 21600,
    lower: 'now() - 6h',
    upper: null,
    label: 'Past 6h',
    duration: '6h',
  },
  {
    seconds: 43200,
    lower: 'now() - 12h',
    upper: null,
    label: 'Past 12h',
    duration: '12h',
  },
  {
    seconds: 86400,
    lower: 'now() - 24h',
    upper: null,
    label: 'Past 24h',
    duration: '24h',
  },
  {
    seconds: 172800,
    lower: 'now() - 2d',
    upper: null,
    label: 'Past 2d',
    duration: '2d',
  },
  {
    seconds: 604800,
    lower: 'now() - 7d',
    upper: null,
    label: 'Past 7d',
    duration: '7d',
  },
  {
    seconds: 2592000,
    lower: 'now() - 30d',
    upper: null,
    label: 'Past 30d',
    duration: '30d',
  },
]

export const DEFAULT_TIME_RANGE: TimeRange = TIME_RANGES[1]

export const ABSOLUTE = 'absolute'
export const INVALID = 'invalid'
export const RELATIVE_LOWER = 'relative lower'
export const RELATIVE_UPPER = 'relative upper'
export const INFLUXQL = 'influxql'

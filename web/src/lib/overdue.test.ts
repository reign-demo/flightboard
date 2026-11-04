import { describe, expect, it } from 'vitest'
import type { Status } from '../api/types'
import { makeWorkOrder } from '../test/fixtures'
import { isOverdue } from './overdue'

const now = new Date('2026-10-22T19:00:00Z')
const activeStatuses: Status[] = ['Published', 'Routed', 'Assigned']
const excludedStatuses: Status[] = ['Draft', 'Work Done', 'Approved', 'Paid', 'Cancelled']

describe('isOverdue', () => {
  it.each(activeStatuses)('returns true for %s with a past end time', (status) => {
    expect(isOverdue(makeWorkOrder({ status }), now)).toBe(true)
  })

  it.each(excludedStatuses)('returns false for %s even with a past end time', (status) => {
    expect(isOverdue(makeWorkOrder({ status }), now)).toBe(false)
  })

  describe.each(activeStatuses)('%s', (status) => {
    it.each(['2026-10-22T19:00:00Z', '2026-10-22T20:00:00Z'])(
      'returns false for an end time at or after now: %s',
      (end) => {
        const wo = makeWorkOrder({ status, schedule: { start: '2026-10-22T14:00:00Z', end } })
        expect(isOverdue(wo, now)).toBe(false)
      },
    )

    it('returns false without a schedule', () => {
      expect(isOverdue(makeWorkOrder({ status, schedule: undefined }), now)).toBe(false)
    })
  })
})

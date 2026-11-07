import { describe, expect, it } from 'vitest'
import { STATUSES } from '../api/types'
import { makeWorkOrder } from '../test/fixtures'
import { DEFAULT_TAB_ID, findTab, tabCounts, TABS } from './tabs'

describe('tabs', () => {
  const workOrders = [
    makeWorkOrder({ id: 1, status: 'Draft', health: 'Alert' }),
    makeWorkOrder({ id: 2, status: 'Published', health: 'Review' }),
    makeWorkOrder({ id: 3, status: 'Routed' }),
    makeWorkOrder({ id: 4, status: 'Assigned' }),
    makeWorkOrder({ id: 5, status: 'Work Done' }),
    makeWorkOrder({ id: 6, status: 'Approved' }),
    makeWorkOrder({ id: 7, status: 'Paid', health: 'Alert' }),
    makeWorkOrder({ id: 8, status: 'Cancelled', health: 'Review' }),
  ]

  it('counts work orders per tab', () => {
    expect(tabCounts(workOrders)).toEqual({
      'in-flight': 4,
      alerts: 2,
      draft: 1,
      'published-routed': 2,
      assigned: 1,
      done: 1,
      approved: 2,
      all: 8,
    })
  })

  it('has unique tab ids', () => {
    expect(new Set(TABS.map((t) => t.id)).size).toBe(TABS.length)
  })

  it('places Alerts immediately after the default In-Flight tab', () => {
    expect(DEFAULT_TAB_ID).toBe('in-flight')
    expect(TABS[0].id).toBe(DEFAULT_TAB_ID)
    expect(TABS[1]).toEqual(expect.objectContaining({ id: 'alerts', label: 'Alerts' }))
    expect(findTab('alerts')).toBe(TABS[1])
  })

  it.each(STATUSES)('matches Alerts by health for %s work orders', (status) => {
    const tab = findTab('alerts')
    const needsAttention = status !== 'Paid' && status !== 'Cancelled'

    expect(tab.matches(makeWorkOrder({ status, health: 'Alert' }))).toBe(needsAttention)
    expect(tab.matches(makeWorkOrder({ status, health: 'Review' }))).toBe(needsAttention)
    expect(tab.matches(makeWorkOrder({ status, health: 'On Track' }))).toBe(false)
  })

  it('falls back to the first tab for an unknown id', () => {
    expect(findTab('nope').id).toBe(TABS[0].id)
    expect(findTab(DEFAULT_TAB_ID).label).toBe('In-Flight')
  })
})

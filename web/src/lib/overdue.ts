import type { WorkOrder } from '../api/types'

export function isOverdue(wo: WorkOrder, now: Date): boolean {
  return (wo.status === 'Published' || wo.status === 'Routed' || wo.status === 'Assigned')
    && wo.schedule !== undefined
    && new Date(wo.schedule.end).getTime() < now.getTime()
}

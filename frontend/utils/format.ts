export function money(cents: number | null | undefined, currency = 'RUB'): string {
  if (cents == null) return '—'
  const value = cents / 100
  try {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency,
      maximumFractionDigits: value % 1 === 0 ? 0 : 2,
    }).format(value)
  } catch {
    return `${value.toLocaleString('ru-RU')} ${currency}`
  }
}

export function toCents(value: string | number): number {
  const n = typeof value === 'number' ? value : parseFloat(String(value).replace(',', '.'))
  return Math.round((Number.isFinite(n) ? n : 0) * 100)
}

export function date(iso: string | null | undefined): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })
}

export function relTime(iso: string | null | undefined): string {
  if (!iso) return ''
  const diff = Date.now() - new Date(iso).getTime()
  const day = 86_400_000
  if (diff < day) return 'сегодня'
  if (diff < 2 * day) return 'вчера'
  if (diff < 7 * day) return `${Math.floor(diff / day)} дн. назад`
  return date(iso)
}

const STATUS_RU: Record<string, string> = {
  draft: 'черновик', open: 'открыт', in_progress: 'в работе', completed: 'завершён',
  cancelled: 'отменён', disputed: 'спор',
  pending: 'ожидает', shortlisted: 'в шорт-листе', accepted: 'принят',
  declined: 'отклонён', withdrawn: 'отозван',
  funded: 'оплачен в escrow', submitted: 'на проверке', approved: 'принят', released: 'выплачен',
  active: 'активен', paused: 'приостановлен',
  requested: 'запрошена', processing: 'в обработке', paid: 'выплачена', rejected: 'отклонена',
  under_review: 'на рассмотрении', resolved_client: 'в пользу заказчика',
  resolved_freelancer: 'в пользу фрилансера', resolved_split: 'частично',
  available: 'доступен', limited: 'ограниченно', unavailable: 'занят', unknown: 'не указано',
}

export function statusRu(s: string): string {
  return STATUS_RU[s] || s
}

export function statusColor(s: string): string {
  if (['open', 'active', 'approved', 'released', 'paid', 'accepted', 'completed', 'available'].includes(s))
    return 'bg-emerald-100 text-emerald-800'
  if (['pending', 'requested', 'processing', 'submitted', 'shortlisted', 'funded', 'draft', 'limited'].includes(s))
    return 'bg-amber-100 text-amber-800'
  if (['cancelled', 'declined', 'rejected', 'withdrawn', 'disputed', 'unavailable'].includes(s))
    return 'bg-rose-100 text-rose-800'
  return 'bg-slate-100 text-slate-700'
}

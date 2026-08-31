export function money(cents: number | null | undefined, currency = 'RUB'): string {
  if (cents == null) return '—'
  const v = cents / 100
  try {
    return new Intl.NumberFormat('ru-RU', {
      style: 'currency',
      currency,
      maximumFractionDigits: v % 1 === 0 ? 0 : 2,
    }).format(v)
  } catch {
    return `${v.toLocaleString('ru-RU')} ${currency}`
  }
}

export function toCents(value: string | number): number {
  const n = typeof value === 'number' ? value : parseFloat(String(value).replace(',', '.'))
  return Math.round((Number.isFinite(n) ? n : 0) * 100)
}

export function fmtDate(iso?: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
}

const RU: Record<string, string> = {
  draft: 'черновик', open: 'открыт', in_progress: 'в работе', completed: 'завершён',
  cancelled: 'отменён', disputed: 'спор',
  pending: 'ожидает', shortlisted: 'в шорт-листе', accepted: 'принят',
  declined: 'отклонён', withdrawn: 'отозван',
  funded: 'в escrow', submitted: 'на проверке', approved: 'принят', released: 'выплачен',
  active: 'активен', paused: 'на паузе',
  requested: 'запрошена', processing: 'в обработке', paid: 'выплачена', rejected: 'отклонена',
  under_review: 'на рассмотрении',
  resolved_client: 'в пользу заказчика', resolved_freelancer: 'в пользу фрилансера',
  available: 'открыт к работе', limited: 'частично занят', unavailable: 'не ищет', unknown: '—',
  any: 'любой', entry: 'junior', intermediate: 'middle', expert: 'senior',
  fixed: 'фиксированная цена', hourly: 'почасовая оплата',
}

export function ru(s?: string): string {
  return s ? RU[s] || s : ''
}

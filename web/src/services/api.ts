import { z } from 'zod'

export type ApiError = { code: string; message: string; fields?: { field: string; message: string }[]; request_id: string }

const apiErrorSchema = z.object({
  code: z.string().min(1),
  message: z.string().min(1),
  request_id: z.string(),
  fields: z.array(z.object({
    field: z.string(),
    message: z.string(),
  })).optional(),
})

const headers = {
  'Content-Type': 'application/json',
  'X-Actor-ID': 'demo-user',
  'X-Actor-Role': 'owner'
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api/v1${path}`, { ...init, headers: { ...headers, ...(init.headers ?? {}) } })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok) {
    const parsed = apiErrorSchema.safeParse(payload.error)
    throw (parsed.success
      ? parsed.data
      : { code: 'INVALID_ERROR_RESPONSE', message: '服务返回了无法识别的错误格式', request_id: response.headers.get('X-Request-ID') ?? '' }) satisfies ApiError
  }
  return payload.data as T
}

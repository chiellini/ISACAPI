import { hasUnsafeUrlPathCharacters } from '@/utils/url'

const DEFAULT_AUTH_REDIRECT = '/dashboard'
const MAX_AUTH_REDIRECT_LENGTH = 2048

/**
 * Keep post-authentication navigation on this application.
 *
 * The backend applies the same shape of checks to OAuth redirects. Keeping the
 * frontend rule in one place also protects password, 2FA, and email-verification
 * flows from protocol-relative or external redirect targets.
 */
export function sanitizeAuthRedirect(
  value: unknown,
  fallback = DEFAULT_AUTH_REDIRECT
): string {
  if (typeof value !== 'string') return fallback

  const redirect = value.trim()
  if (
    !redirect.startsWith('/') ||
    /^\/[\\/]/.test(redirect) ||
    redirect.includes('://') ||
    hasUnsafeUrlPathCharacters(redirect) ||
    redirect.length > MAX_AUTH_REDIRECT_LENGTH
  ) {
    return fallback
  }

  return redirect
}

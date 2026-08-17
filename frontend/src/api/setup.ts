/**
 * Setup API endpoints
 */
import axios from 'axios'
import { buildGatewayUrl } from './url'

// Create a separate client for setup endpoints (not under /api/v1)
const setupClient = axios.create({
  baseURL: buildGatewayUrl('/').replace(/\/+$/, ''),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

const SETUP_TOKEN_SESSION_KEY = 'isac_setup_bootstrap_token'

export function setSetupBootstrapToken(token: string): void {
  if (typeof sessionStorage === 'undefined') return
  const normalized = token.trim()
  if (normalized) {
    sessionStorage.setItem(SETUP_TOKEN_SESSION_KEY, normalized)
  } else {
    sessionStorage.removeItem(SETUP_TOKEN_SESSION_KEY)
  }
}

function setupBootstrapHeaders(): Record<string, string> {
  if (typeof sessionStorage === 'undefined') return {}
  const token = sessionStorage.getItem(SETUP_TOKEN_SESSION_KEY)?.trim()
  return token ? { 'X-Setup-Token': token } : {}
}

export interface SetupStatus {
  needs_setup: boolean
  step: string
}

export interface DatabaseConfig {
  host: string
  port: number
  user: string
  password: string
  dbname: string
  sslmode: string
}

export interface RedisConfig {
  host: string
  port: number
  username: string
  password: string
  db: number
  enable_tls: boolean
}

export interface AdminConfig {
  email: string
  password: string
}

export interface ServerConfig {
  host: string
  port: number
  mode: string
}

export interface InstallRequest {
  database: DatabaseConfig
  redis: RedisConfig
  admin: AdminConfig
  server: ServerConfig
}

export interface InstallResponse {
  message: string
  restart: boolean
}

/**
 * Get setup status
 */
export async function getSetupStatus(): Promise<SetupStatus> {
  const response = await setupClient.get('/setup/status')
  return response.data.data
}

/**
 * Test database connection
 */
export async function testDatabase(config: DatabaseConfig): Promise<void> {
  await setupClient.post('/setup/test-db', config, { headers: setupBootstrapHeaders() })
}

/**
 * Test Redis connection
 */
export async function testRedis(config: RedisConfig): Promise<void> {
  await setupClient.post('/setup/test-redis', config, { headers: setupBootstrapHeaders() })
}

/**
 * Perform installation
 */
export async function install(config: InstallRequest): Promise<InstallResponse> {
  const response = await setupClient.post('/setup/install', config, { headers: setupBootstrapHeaders() })
  return response.data.data
}

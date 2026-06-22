/**
 * 注册接口防刷配置。
 *
 * 为阻断照搬上游开源项目接口格式的批量注册脚本，注册与发送验证码接口：
 *  1. 使用自定义、非默认的路径（偏离开源的 /auth/register、/auth/send-verify-code）；
 *  2. 要求请求携带自定义 X-Reg-Guard 头，值为约定密钥。
 *
 * 注意：这是「提高门槛」的纵深防御，不能替代 Turnstile / 邮箱验证 / 限流。
 *
 * 若要修改，请保持前后端一致：
 *  - 路径：同步修改后端 backend/internal/server/routes/auth.go 中的 registerPath / sendVerifyCodePath；
 *  - 密钥：构建时设置 VITE_REG_GUARD_TOKEN，并将后端 SERVER_REGISTER_GUARD_TOKEN 设为相同值。
 */

export const REG_GUARD_HEADER = 'X-Reg-Guard'

// 默认值仅用于让 fork 出来的实例开箱可用；生产部署请通过 VITE_REG_GUARD_TOKEN 覆盖。
export const REG_GUARD_TOKEN: string = import.meta.env.VITE_REG_GUARD_TOKEN || 's2a-rg-7Kq2xZ9m'

// 自定义注册接口路径（需与后端保持一致）。
export const REGISTER_PATH = '/auth/r/x7k2q'
export const SEND_VERIFY_CODE_PATH = '/auth/r/x7k2q/code'

// 注册类请求需要附带的请求头，便于各处复用。
export const regGuardHeaders = (): Record<string, string> => ({
  [REG_GUARD_HEADER]: REG_GUARD_TOKEN,
})

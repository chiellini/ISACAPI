package middleware

import (
	"crypto/subtle"

	"github.com/gin-gonic/gin"
)

// RegistrationGuardHeader 是注册相关接口要求携带的自定义请求头名称。
// 前端在调用注册/发送验证码接口时会自动带上该头，值为部署侧配置的密钥。
const RegistrationGuardHeader = "X-Reg-Guard"

// NewRegistrationGuard 构造注册接口防刷守卫。
//
// 目的：阻断照搬上游开源项目接口格式的批量注册脚本。请求必须携带 RegistrationGuardHeader
// 头且值与配置的 token 完全一致，否则直接按「路由不存在」返回 404——既不泄露接口真实存在，
// 也让通用脚本无从判断失败原因。
//
// 这是一道「提高门槛」的纵深防御，不能替代 Turnstile / 邮箱验证 / 限流等手段。
// 若 token 为空，则不启用守卫（放行），便于未配置时回退到原有行为。
func NewRegistrationGuard(token string) gin.HandlerFunc {
	expected := []byte(token)
	return func(c *gin.Context) {
		if len(expected) == 0 {
			c.Next()
			return
		}
		got := []byte(c.GetHeader(RegistrationGuardHeader))
		// 使用常量时间比较，避免通过响应时延侧信道猜测 token。
		if len(got) != len(expected) || subtle.ConstantTimeCompare(got, expected) != 1 {
			c.AbortWithStatus(404)
			return
		}
		c.Next()
	}
}

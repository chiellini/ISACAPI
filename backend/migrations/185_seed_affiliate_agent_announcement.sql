-- Announce the affiliate agent workflow once after it becomes available.
-- The title guard keeps the seed idempotent if the SQL is replayed manually.
INSERT INTO announcements (
    title,
    content,
    status,
    notify_mode,
    targeting,
    starts_at,
    ends_at,
    created_by,
    updated_by,
    created_at,
    updated_at
)
SELECT
    '代理合作与佣金提现功能上线',
    $announcement$
## 代理合作功能现已上线

邀请新用户注册后，受邀用户后续的真实充值可按您的代理佣金比例持续产生佣金。

- 从左侧菜单进入 **邀请返利（代理中心）**，复制邀请码或邀请链接
- 查看可提现、冻结中、提现预留及欠款金额
- 绑定支付宝、银行卡或 USDT 收款账户并申请提现
- 提现审核通过后由平台线下转账，完成后状态显示 **已经转账**
- 如受邀用户的充值发生退款，对应佣金会同步冲正

> 代理资格需管理员开通；佣金比例、冻结期和最低提现金额以平台配置为准。若左侧未显示“邀请返利”，请联系管理员确认功能开关和界面模式。

[立即前往代理中心](/affiliate)
$announcement$,
    'active',
    'popup',
    '{"any_of":[]}'::jsonb,
    NOW(),
    NULL,
    NULL,
    NULL,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1
    FROM announcements
    WHERE title = '代理合作与佣金提现功能上线'
);

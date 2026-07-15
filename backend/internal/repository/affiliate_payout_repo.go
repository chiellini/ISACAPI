package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

type affiliatePayoutRepository struct{ client *dbent.Client }

func NewAffiliatePayoutRepository(client *dbent.Client) service.AffiliatePayoutRepository {
	return &affiliatePayoutRepository{client: client}
}

func (r *affiliatePayoutRepository) withTx(ctx context.Context, fn func(context.Context, *dbent.Client) error) error {
	if tx := dbent.TxFromContext(ctx); tx != nil { return fn(ctx, tx.Client()) }
	tx, err := r.client.Tx(ctx); if err != nil { return fmt.Errorf("begin affiliate payout transaction: %w", err) }
	defer func(){ _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)
	if err := fn(txCtx, tx.Client()); err != nil { return err }
	if err := tx.Commit(); err != nil { return fmt.Errorf("commit affiliate payout transaction: %w", err) }
	return nil
}

func (r *affiliatePayoutRepository) GetAgentStatus(ctx context.Context, userID int64) (string, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `SELECT agent_status FROM user_affiliates WHERE user_id=$1`, userID)
	if err != nil { return "", err }; defer rows.Close()
	if !rows.Next() { return service.AffiliateAgentStatusInactive, nil }
	var status string; if err := rows.Scan(&status); err != nil { return "", err }; return status, rows.Err()
}

func (r *affiliatePayoutRepository) ListPaymentAccounts(ctx context.Context, userID int64) ([]service.AffiliatePaymentAccount, error) {
	client := clientFromContext(ctx, r.client)
	rows, err := client.QueryContext(ctx, `SELECT id,user_id,type,masked_summary,is_default,created_at,updated_at FROM user_affiliate_payment_accounts WHERE user_id=$1 ORDER BY is_default DESC,id`, userID)
	if err != nil { return nil, err }; defer rows.Close()
	items := make([]service.AffiliatePaymentAccount,0)
	for rows.Next() { var v service.AffiliatePaymentAccount; if err := rows.Scan(&v.ID,&v.UserID,&v.Type,&v.MaskedSummary,&v.IsDefault,&v.CreatedAt,&v.UpdatedAt); err != nil { return nil,err }; items=append(items,v) }
	return items, rows.Err()
}

func (r *affiliatePayoutRepository) CreatePaymentAccount(ctx context.Context, userID int64, typ, encrypted, summary string, requestedDefault bool) (*service.AffiliatePaymentAccount, error) {
	var out service.AffiliatePaymentAccount
	err := r.withTx(ctx, func(txCtx context.Context, c *dbent.Client) error {
		var count int
		rows, err := c.QueryContext(txCtx, `SELECT COUNT(*) FROM user_affiliate_payment_accounts WHERE user_id=$1`,userID); if err != nil{return err}
		if !rows.Next(){rows.Close();return errors.New("count payment accounts returned no row")}; if err=rows.Scan(&count);err!=nil{rows.Close();return err}; rows.Close()
		isDefault := requestedDefault || count==0
		if isDefault { if _,err=c.ExecContext(txCtx,`UPDATE user_affiliate_payment_accounts SET is_default=FALSE,updated_at=NOW() WHERE user_id=$1 AND is_default`,userID);err!=nil{return err} }
		rows,err=c.QueryContext(txCtx,`INSERT INTO user_affiliate_payment_accounts(user_id,type,details_encrypted,masked_summary,is_default,created_at,updated_at) VALUES($1,$2,$3,$4,$5,NOW(),NOW()) RETURNING id,user_id,type,masked_summary,is_default,created_at,updated_at`,userID,typ,encrypted,summary,isDefault)
		if err!=nil{return err}; defer rows.Close(); if !rows.Next(){return errors.New("create payment account returned no row")}; return rows.Scan(&out.ID,&out.UserID,&out.Type,&out.MaskedSummary,&out.IsDefault,&out.CreatedAt,&out.UpdatedAt)
	}); if err!=nil{return nil,err}; return &out,nil
}

func (r *affiliatePayoutRepository) UpdatePaymentAccount(ctx context.Context, userID, accountID int64, typ, encrypted, summary string, requestedDefault bool) (*service.AffiliatePaymentAccount, error) {
	var out service.AffiliatePaymentAccount
	err:=r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		if requestedDefault { if _,err:=c.ExecContext(txCtx,`UPDATE user_affiliate_payment_accounts SET is_default=FALSE,updated_at=NOW() WHERE user_id=$1 AND is_default`,userID);err!=nil{return err} }
		rows,err:=c.QueryContext(txCtx,`UPDATE user_affiliate_payment_accounts SET type=$1,details_encrypted=$2,masked_summary=$3,is_default=$4,updated_at=NOW() WHERE id=$5 AND user_id=$6 RETURNING id,user_id,type,masked_summary,is_default,created_at,updated_at`,typ,encrypted,summary,requestedDefault,accountID,userID)
		if err!=nil{return err}; if !rows.Next(){rows.Close();return service.ErrAffiliatePaymentAccountNotFound}; if err=rows.Scan(&out.ID,&out.UserID,&out.Type,&out.MaskedSummary,&out.IsDefault,&out.CreatedAt,&out.UpdatedAt);err!=nil{rows.Close();return err}; rows.Close()
		if !requestedDefault { // Never leave an account set without a default destination.
			if _,err=c.ExecContext(txCtx,`UPDATE user_affiliate_payment_accounts SET is_default=TRUE,updated_at=NOW() WHERE id=(SELECT id FROM user_affiliate_payment_accounts WHERE user_id=$1 ORDER BY id LIMIT 1) AND NOT EXISTS(SELECT 1 FROM user_affiliate_payment_accounts WHERE user_id=$1 AND is_default)`,userID);err!=nil{return err}
			rows,err=c.QueryContext(txCtx,`SELECT is_default,updated_at FROM user_affiliate_payment_accounts WHERE id=$1`,accountID);if err!=nil{return err};if rows.Next(){err=rows.Scan(&out.IsDefault,&out.UpdatedAt)};rows.Close();return err
		}; return nil
	});if err!=nil{return nil,err};return &out,nil
}

func (r *affiliatePayoutRepository) DeletePaymentAccount(ctx context.Context,userID,accountID int64)error{
	return r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		res,err:=c.ExecContext(txCtx,`DELETE FROM user_affiliate_payment_accounts WHERE id=$1 AND user_id=$2`,accountID,userID);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return service.ErrAffiliatePaymentAccountNotFound}
		_,err=c.ExecContext(txCtx,`UPDATE user_affiliate_payment_accounts SET is_default=TRUE,updated_at=NOW() WHERE id=(SELECT id FROM user_affiliate_payment_accounts WHERE user_id=$1 ORDER BY id LIMIT 1) AND NOT EXISTS(SELECT 1 FROM user_affiliate_payment_accounts WHERE user_id=$1 AND is_default)`,userID);return err
	})
}

func (r *affiliatePayoutRepository) CreateWithdrawal(ctx context.Context,userID,paymentAccountID int64,amount,minimum float64,idempotencyKey string)(*service.AffiliateWithdrawal,error){
	var out *service.AffiliateWithdrawal
	err:=r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		rows,err:=c.QueryContext(txCtx,`SELECT agent_status,aff_quota::double precision,aff_debt::double precision FROM user_affiliates WHERE user_id=$1 FOR UPDATE`,userID);if err!=nil{return err};if !rows.Next(){rows.Close();return service.ErrAffiliateAgentInactive}
		var status string;var available,debt float64;if err=rows.Scan(&status,&available,&debt);err!=nil{rows.Close();return err};rows.Close()
		thawed,err:=thawFrozenQuotaTx(txCtx,c,userID);if err!=nil{return err};available+=thawed
		if idempotencyKey!="" { existing,err:=queryAffiliateWithdrawalByUserKey(txCtx,c,userID,idempotencyKey);if err==nil{if !sameAffiliateWithdrawalRequest(existing,paymentAccountID,amount){return service.ErrAffiliateIdempotencyKeyConflict};out=existing;return nil};if !errors.Is(err,service.ErrAffiliateWithdrawalNotFound){return err} }
		if status!=service.AffiliateAgentStatusActive{return service.ErrAffiliateAgentInactive};if debt>0{return service.ErrAffiliateWithdrawalDebt};if amount<minimum{return service.ErrAffiliateWithdrawalTooSmall}
		if available+1e-9<amount{return service.ErrAffiliateWithdrawalInsufficient}
		rows,err=c.QueryContext(txCtx,`SELECT type,details_encrypted,masked_summary FROM user_affiliate_payment_accounts WHERE id=$1 AND user_id=$2 FOR SHARE`,paymentAccountID,userID);if err!=nil{return err};if !rows.Next(){rows.Close();return service.ErrAffiliatePaymentAccountNotFound}
		var typ,encrypted,summary string;if err=rows.Scan(&typ,&encrypted,&summary);err!=nil{rows.Close();return err};rows.Close()
		res,err:=c.ExecContext(txCtx,`UPDATE user_affiliates SET aff_quota=aff_quota-$1,aff_withdrawal_pending=aff_withdrawal_pending+$1,updated_at=NOW() WHERE user_id=$2 AND aff_quota >= $1`,amount,userID);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return service.ErrAffiliateWithdrawalInsufficient}
		rows,err=c.QueryContext(txCtx,`INSERT INTO user_affiliate_withdrawals(user_id,amount,status,payment_account_id,payment_account_type,payment_details_encrypted,payment_account_summary,idempotency_key,submitted_at,created_at,updated_at) VALUES($1,$2,'submitted',$3,$4,$5,$6,$7,NOW(),NOW(),NOW()) RETURNING id`,userID,amount,paymentAccountID,typ,encrypted,summary,nullString(idempotencyKey));if err!=nil{return err};var id int64;if !rows.Next(){rows.Close();return errors.New("create withdrawal returned no row")};if err=rows.Scan(&id);err!=nil{rows.Close();return err};rows.Close()
		if _,err=c.ExecContext(txCtx,`INSERT INTO user_affiliate_ledger(user_id,action,amount,source_user_id,aff_quota_after,created_at,updated_at) SELECT user_id,'withdrawal_hold',$1,NULL,aff_quota,NOW(),NOW() FROM user_affiliates WHERE user_id=$2`,amount,userID);err!=nil{return err}
		out,err=queryAffiliateWithdrawalByID(txCtx,c,id);return err
	});if err!=nil{return nil,err};return out,nil
}

func (r *affiliatePayoutRepository) ListUserWithdrawals(ctx context.Context,userID int64,page,pageSize int)([]service.AffiliateWithdrawal,int64,error){
	client:=clientFromContext(ctx,r.client);return listAffiliateWithdrawals(ctx,client,`w.user_id=$1`,[]any{userID},page,pageSize)
}

func (r *affiliatePayoutRepository) CancelWithdrawal(ctx context.Context,userID,withdrawalID int64)(*service.AffiliateWithdrawal,error){
	var out *service.AffiliateWithdrawal
	err:=r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		if err:=lockAffiliate(txCtx,c,userID);err!=nil{return err};w,err:=queryAffiliateWithdrawalForUpdate(txCtx,c,withdrawalID);if err!=nil{return err};if w.UserID!=userID{return service.ErrAffiliateWithdrawalNotFound};if w.Status!=service.AffiliateWithdrawalSubmitted{return service.ErrAffiliateWithdrawalState}
		if _,err=c.ExecContext(txCtx,`UPDATE user_affiliate_withdrawals SET status='canceled',canceled_at=NOW(),cancel_reason='canceled by affiliate',updated_at=NOW() WHERE id=$1 AND status='submitted'`,withdrawalID);err!=nil{return err}
		if err=releaseWithdrawalHold(txCtx,c,userID,w.Amount,"withdrawal_release");err!=nil{return err};out,err=queryAffiliateWithdrawalByID(txCtx,c,withdrawalID);return err
	});return out,err
}

func (r *affiliatePayoutRepository) ListAgents(ctx context.Context,filter service.AffiliatePayoutListFilter)([]service.AffiliateAgentAdminEntry,int64,error){
	client:=clientFromContext(ctx,r.client);where,args:=[]string{"1=1"},[]any{};if s:=strings.TrimSpace(filter.Search);s!=""{args=append(args,"%"+s+"%");p:=fmt.Sprintf("$%d",len(args));where=append(where,"(u.email ILIKE "+p+" OR u.username ILIKE "+p+" OR ua.aff_code ILIKE "+p+")")};if st:=strings.TrimSpace(filter.Status);st!=""{args=append(args,st);where=append(where,fmt.Sprintf("ua.agent_status=$%d",len(args)))}
	wsql:=strings.Join(where," AND ");count,err:=queryCount(ctx,client,`SELECT COUNT(*) FROM user_affiliates ua JOIN users u ON u.id=ua.user_id WHERE `+wsql,args...);if err!=nil{return nil,0,err};args=append(args,filter.PageSize,(filter.Page-1)*filter.PageSize)
	rows,err:=client.QueryContext(ctx,`SELECT ua.user_id,COALESCE(u.email,''),COALESCE(u.username,''),ua.agent_status,ua.aff_code,COALESCE(ua.aff_rebate_rate_percent,0)::double precision,ua.aff_count,ua.aff_quota::double precision,ua.aff_frozen_quota::double precision,ua.aff_withdrawal_pending::double precision,ua.aff_debt::double precision FROM user_affiliates ua JOIN users u ON u.id=ua.user_id WHERE `+wsql+fmt.Sprintf(" ORDER BY ua.updated_at DESC LIMIT $%d OFFSET $%d",len(args)-1,len(args)),args...);if err!=nil{return nil,0,err};defer rows.Close();items:=make([]service.AffiliateAgentAdminEntry,0)
	for rows.Next(){var v service.AffiliateAgentAdminEntry;if err=rows.Scan(&v.UserID,&v.Email,&v.Username,&v.Status,&v.AffCode,&v.RebateRatePercent,&v.InvitedCount,&v.AvailableCommission,&v.FrozenCommission,&v.WithdrawalReserved,&v.Debt);err!=nil{return nil,0,err};items=append(items,v)};return items,count,rows.Err()
}

func (r *affiliatePayoutRepository) GetAgent(ctx context.Context,userID int64)(*service.AffiliateAgentAdminEntry,error){
	client:=clientFromContext(ctx,r.client);rows,err:=client.QueryContext(ctx,`SELECT ua.user_id,COALESCE(u.email,''),COALESCE(u.username,''),ua.agent_status,ua.aff_code,COALESCE(ua.aff_rebate_rate_percent,0)::double precision,ua.aff_count,ua.aff_quota::double precision,ua.aff_frozen_quota::double precision,ua.aff_withdrawal_pending::double precision,ua.aff_debt::double precision FROM user_affiliates ua JOIN users u ON u.id=ua.user_id WHERE ua.user_id=$1`,userID);if err!=nil{return nil,err};defer rows.Close();if !rows.Next(){return nil,service.ErrAffiliateProfileNotFound};var v service.AffiliateAgentAdminEntry;if err=rows.Scan(&v.UserID,&v.Email,&v.Username,&v.Status,&v.AffCode,&v.RebateRatePercent,&v.InvitedCount,&v.AvailableCommission,&v.FrozenCommission,&v.WithdrawalReserved,&v.Debt);err!=nil{return nil,err};return &v,nil
}

func (r *affiliatePayoutRepository) SetAgentStatus(ctx context.Context,userID int64,status string,operatorID int64,key string)error{
	return r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		if existingID,after,ok,err:=affiliateAdminAuditTarget(txCtx,c,"agent_status_changed",key);err!=nil{return err}else if ok{if sameAffiliateAgentStatusRequest(existingID,after,userID,status){return nil};return service.ErrAffiliateIdempotencyKeyConflict}
		if _,err:=ensureUserAffiliateWithClient(txCtx,c,userID);err!=nil{return err};if err:=lockAffiliate(txCtx,c,userID);err!=nil{return err}
		var old string;rows,err:=c.QueryContext(txCtx,`SELECT agent_status FROM user_affiliates WHERE user_id=$1`,userID);if err!=nil{return err};if rows.Next(){err=rows.Scan(&old)};rows.Close();if err!=nil{return err}
		if old!=status{if _,err=c.ExecContext(txCtx,`UPDATE user_affiliates SET agent_status=$1,agent_status_updated_by=$2,agent_status_updated_at=NOW(),updated_at=NOW() WHERE user_id=$3`,status,operatorID,userID);err!=nil{return err}}
		detail,_:=json.Marshal(map[string]any{"before":old,"after":status});return insertAffiliateAdminAudit(txCtx,c,operatorID,&userID,nil,"agent_status_changed",key,string(detail))
	})
}

func (r *affiliatePayoutRepository) ListWithdrawals(ctx context.Context,filter service.AffiliatePayoutListFilter)([]service.AffiliateWithdrawal,int64,error){
	client:=clientFromContext(ctx,r.client);where,args:=[]string{"1=1"},[]any{};if s:=strings.TrimSpace(filter.Search);s!=""{args=append(args,"%"+s+"%");p:=fmt.Sprintf("$%d",len(args));where=append(where,"(u.email ILIKE "+p+" OR u.username ILIKE "+p+" OR w.external_reference ILIKE "+p+")")};if st:=strings.TrimSpace(filter.Status);st!=""{args=append(args,st);where=append(where,fmt.Sprintf("w.status=$%d",len(args)))};return listAffiliateWithdrawals(ctx,client,strings.Join(where," AND "),args,filter.Page,filter.PageSize)
}
func (r *affiliatePayoutRepository) GetWithdrawal(ctx context.Context,id int64)(*service.AffiliateWithdrawal,error){return queryAffiliateWithdrawalByID(ctx,clientFromContext(ctx,r.client),id)}

func (r *affiliatePayoutRepository) ApproveWithdrawal(ctx context.Context,id,operatorID int64,key string)(*service.AffiliateWithdrawal,error){return r.adminTransition(ctx,id,operatorID,key,"withdrawal_approved",func(txCtx context.Context,c *dbent.Client,w *service.AffiliateWithdrawal)error{if w.Status!=service.AffiliateWithdrawalSubmitted{return service.ErrAffiliateWithdrawalState};_,err:=c.ExecContext(txCtx,`UPDATE user_affiliate_withdrawals SET status='approved',reviewed_at=NOW(),reviewed_by=$1,updated_at=NOW() WHERE id=$2 AND status='submitted'`,operatorID,id);return err})}

func (r *affiliatePayoutRepository) RejectWithdrawal(ctx context.Context,id,operatorID int64,reason,key string)(*service.AffiliateWithdrawal,error){return r.adminTransition(ctx,id,operatorID,key,"withdrawal_rejected",func(txCtx context.Context,c *dbent.Client,w *service.AffiliateWithdrawal)error{if w.Status!=service.AffiliateWithdrawalSubmitted&&w.Status!=service.AffiliateWithdrawalApproved{return service.ErrAffiliateWithdrawalState};res,err:=c.ExecContext(txCtx,`UPDATE user_affiliate_withdrawals SET status='rejected',reviewed_at=NOW(),reviewed_by=$1,reject_reason=$2,updated_at=NOW() WHERE id=$3 AND status IN ('submitted','approved')`,operatorID,reason,id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return service.ErrAffiliateWithdrawalState};return releaseWithdrawalHold(txCtx,c,w.UserID,w.Amount,"withdrawal_release")})}

func (r *affiliatePayoutRepository) MarkWithdrawalPaid(ctx context.Context,id,operatorID int64,input service.AffiliateMarkPaidInput,key string)(*service.AffiliateWithdrawal,error){return r.adminTransition(ctx,id,operatorID,key,"withdrawal_paid",func(txCtx context.Context,c *dbent.Client,w *service.AffiliateWithdrawal)error{if w.Status!=service.AffiliateWithdrawalApproved{return service.ErrAffiliateWithdrawalState};res,err:=c.ExecContext(txCtx,`UPDATE user_affiliate_withdrawals SET status='paid',paid_at=NOW(),paid_by=$1,actual_currency=$2,actual_amount=$3,exchange_rate=$4,external_reference=$5,updated_at=NOW() WHERE id=$6 AND status='approved'`,operatorID,input.ActualCurrency,input.ActualAmount,input.ExchangeRate,input.ExternalReference,id);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return service.ErrAffiliateWithdrawalState};res,err=c.ExecContext(txCtx,`UPDATE user_affiliates SET aff_withdrawal_pending=aff_withdrawal_pending-$1,updated_at=NOW() WHERE user_id=$2 AND aff_withdrawal_pending >= $1`,w.Amount,w.UserID);if err!=nil{return err};n,_=res.RowsAffected();if n==0{return errors.New("affiliate withdrawal pending balance invariant violated")};_,err=c.ExecContext(txCtx,`INSERT INTO user_affiliate_ledger(user_id,action,amount,source_user_id,created_at,updated_at) VALUES($1,'withdrawal_paid',$2,NULL,NOW(),NOW())`,w.UserID,w.Amount);return err})}

func (r *affiliatePayoutRepository) adminTransition(ctx context.Context,id,operatorID int64,key,action string,apply func(context.Context,*dbent.Client,*service.AffiliateWithdrawal)error)(*service.AffiliateWithdrawal,error){
	var out *service.AffiliateWithdrawal;err:=r.withTx(ctx,func(txCtx context.Context,c *dbent.Client)error{
		if existingID,ok,err:=affiliateAdminAuditWithdrawal(txCtx,c,action,key);err!=nil{return err}else if ok{if existingID!=id{return service.ErrAffiliateWithdrawalState};out,err=queryAffiliateWithdrawalByID(txCtx,c,id);return err}
		userID,err:=queryWithdrawalUserID(txCtx,c,id);if err!=nil{return err};if err=lockAffiliate(txCtx,c,userID);err!=nil{return err};w,err:=queryAffiliateWithdrawalForUpdate(txCtx,c,id);if err!=nil{return err};if err=apply(txCtx,c,w);err!=nil{return err}
		detail,_:=json.Marshal(map[string]any{"from":w.Status,"withdrawal_id":id});if err=insertAffiliateAdminAudit(txCtx,c,operatorID,&userID,&id,action,key,string(detail));err!=nil{return err};out,err=queryAffiliateWithdrawalByID(txCtx,c,id);return err
	});return out,err
}

const affiliateWithdrawalSelect=`SELECT w.id,w.user_id,COALESCE(u.email,''),COALESCE(u.username,''),w.amount::double precision,w.status,w.payment_account_id,w.payment_account_type,w.payment_account_summary,w.payment_details_encrypted,w.submitted_at,w.reviewed_at,w.paid_at,w.reject_reason,w.actual_currency,w.actual_amount::double precision,w.exchange_rate::double precision,w.external_reference,w.created_at,w.updated_at FROM user_affiliate_withdrawals w JOIN users u ON u.id=w.user_id`

func listAffiliateWithdrawals(ctx context.Context,c affiliateQueryExecer,where string,args []any,page,pageSize int)([]service.AffiliateWithdrawal,int64,error){count,err:=queryCount(ctx,c,`SELECT COUNT(*) FROM user_affiliate_withdrawals w JOIN users u ON u.id=w.user_id WHERE `+where,args...);if err!=nil{return nil,0,err};args=append(args,pageSize,(page-1)*pageSize);rows,err:=c.QueryContext(ctx,affiliateWithdrawalSelect+` WHERE `+where+fmt.Sprintf(" ORDER BY w.created_at DESC LIMIT $%d OFFSET $%d",len(args)-1,len(args)),args...);if err!=nil{return nil,0,err};defer rows.Close();items:=make([]service.AffiliateWithdrawal,0);for rows.Next(){v,err:=scanAffiliateWithdrawal(rows);if err!=nil{return nil,0,err};items=append(items,*v)};return items,count,rows.Err()}
func queryAffiliateWithdrawalByID(ctx context.Context,c affiliateQueryExecer,id int64)(*service.AffiliateWithdrawal,error){rows,err:=c.QueryContext(ctx,affiliateWithdrawalSelect+` WHERE w.id=$1`,id);if err!=nil{return nil,err};defer rows.Close();if !rows.Next(){return nil,service.ErrAffiliateWithdrawalNotFound};return scanAffiliateWithdrawal(rows)}
func queryAffiliateWithdrawalByUserKey(ctx context.Context,c affiliateQueryExecer,userID int64,key string)(*service.AffiliateWithdrawal,error){rows,err:=c.QueryContext(ctx,affiliateWithdrawalSelect+` WHERE w.user_id=$1 AND w.idempotency_key=$2`,userID,key);if err!=nil{return nil,err};defer rows.Close();if !rows.Next(){return nil,service.ErrAffiliateWithdrawalNotFound};return scanAffiliateWithdrawal(rows)}
func queryAffiliateWithdrawalForUpdate(ctx context.Context,c affiliateQueryExecer,id int64)(*service.AffiliateWithdrawal,error){rows,err:=c.QueryContext(ctx,affiliateWithdrawalSelect+` WHERE w.id=$1 FOR UPDATE OF w`,id);if err!=nil{return nil,err};defer rows.Close();if !rows.Next(){return nil,service.ErrAffiliateWithdrawalNotFound};return scanAffiliateWithdrawal(rows)}
func scanAffiliateWithdrawal(rows *sql.Rows)(*service.AffiliateWithdrawal,error){var v service.AffiliateWithdrawal;var accountID sql.NullInt64;var reviewed,paid sql.NullTime;var reject,currency,ref sql.NullString;var actual,rate sql.NullFloat64;if err:=rows.Scan(&v.ID,&v.UserID,&v.UserEmail,&v.Username,&v.Amount,&v.Status,&accountID,&v.PaymentAccountType,&v.PaymentAccountSummary,&v.PaymentDetailsEncrypted,&v.SubmittedAt,&reviewed,&paid,&reject,&currency,&actual,&rate,&ref,&v.CreatedAt,&v.UpdatedAt);err!=nil{return nil,err};v.StatusLabel=withdrawalStatusLabel(v.Status);if accountID.Valid{v.PaymentAccountID=&accountID.Int64};if reviewed.Valid{v.ReviewedAt=&reviewed.Time};if paid.Valid{v.PaidAt=&paid.Time};if reject.Valid{v.RejectReason=&reject.String};if currency.Valid{v.ActualCurrency=&currency.String};if actual.Valid{v.ActualAmount=&actual.Float64};if rate.Valid{v.ExchangeRate=&rate.Float64};if ref.Valid{v.ExternalReference=&ref.String};return &v,nil}
func withdrawalStatusLabel(status string)string{switch status{case"submitted":return"待审核";case"approved":return"待转账";case"paid":return"已经转账";case"rejected":return"已拒绝";case"canceled":return"已取消"};return status}
func lockAffiliate(ctx context.Context,c affiliateQueryExecer,userID int64)error{rows,err:=c.QueryContext(ctx,`SELECT user_id FROM user_affiliates WHERE user_id=$1 FOR UPDATE`,userID);if err!=nil{return err};defer rows.Close();if !rows.Next(){return service.ErrAffiliateProfileNotFound};return nil}
func queryWithdrawalUserID(ctx context.Context,c affiliateQueryExecer,id int64)(int64,error){rows,err:=c.QueryContext(ctx,`SELECT user_id FROM user_affiliate_withdrawals WHERE id=$1`,id);if err!=nil{return 0,err};defer rows.Close();if !rows.Next(){return 0,service.ErrAffiliateWithdrawalNotFound};var userID int64;if err=rows.Scan(&userID);err!=nil{return 0,err};return userID,nil}
func releaseWithdrawalHold(ctx context.Context,c affiliateQueryExecer,userID int64,amount float64,action string)error{res,err:=c.ExecContext(ctx,`UPDATE user_affiliates SET aff_quota=aff_quota+$1,aff_withdrawal_pending=aff_withdrawal_pending-$1,updated_at=NOW() WHERE user_id=$2 AND aff_withdrawal_pending >= $1`,amount,userID);if err!=nil{return err};n,_:=res.RowsAffected();if n==0{return errors.New("affiliate withdrawal pending balance invariant violated")};_,err=c.ExecContext(ctx,`INSERT INTO user_affiliate_ledger(user_id,action,amount,source_user_id,created_at,updated_at) VALUES($1,$2,$3,NULL,NOW(),NOW())`,userID,action,amount);return err}
func insertAffiliateAdminAudit(ctx context.Context,c affiliateQueryExecer,operatorID int64,targetID,withdrawalID *int64,action,key,detail string)error{_,err:=c.ExecContext(ctx,`INSERT INTO user_affiliate_admin_audits(operator_user_id,target_user_id,withdrawal_id,action,idempotency_key,detail,created_at) VALUES($1,$2,$3,$4,$5,$6::jsonb,NOW())`,operatorID,nullableInt64(targetID),nullableInt64(withdrawalID),action,nullString(key),detail);return err}
func affiliateAdminAuditWithdrawal(ctx context.Context,c affiliateQueryExecer,action,key string)(int64,bool,error){rows,err:=c.QueryContext(ctx,`SELECT withdrawal_id FROM user_affiliate_admin_audits WHERE action=$1 AND idempotency_key=$2`,action,key);if err!=nil{return 0,false,err};defer rows.Close();if !rows.Next(){return 0,false,nil};var id sql.NullInt64;if err=rows.Scan(&id);err!=nil{return 0,false,err};return id.Int64,id.Valid,nil}
func affiliateAdminAuditTarget(ctx context.Context,c affiliateQueryExecer,action,key string)(int64,string,bool,error){rows,err:=c.QueryContext(ctx,`SELECT target_user_id,COALESCE(detail->>'after','') FROM user_affiliate_admin_audits WHERE action=$1 AND idempotency_key=$2`,action,key);if err!=nil{return 0,"",false,err};defer rows.Close();if !rows.Next(){return 0,"",false,nil};var id sql.NullInt64;var after string;if err=rows.Scan(&id,&after);err!=nil{return 0,"",false,err};return id.Int64,after,id.Valid,nil}
func queryCount(ctx context.Context,c affiliateQueryExecer,q string,args ...any)(int64,error){rows,err:=c.QueryContext(ctx,q,args...);if err!=nil{return 0,err};defer rows.Close();if !rows.Next(){return 0,errors.New("count returned no row")};var n int64;return n,rows.Scan(&n)}
func nullableInt64(v *int64)any{if v==nil{return nil};return *v}
func nullString(v string)any{if strings.TrimSpace(v)==""{return nil};return strings.TrimSpace(v)}
func sameAffiliateWithdrawalRequest(existing *service.AffiliateWithdrawal,accountID int64,amount float64)bool{return existing!=nil&&existing.PaymentAccountID!=nil&&*existing.PaymentAccountID==accountID&&existing.Amount-amount<=0.00000001&&amount-existing.Amount<=0.00000001}
func sameAffiliateAgentStatusRequest(existingID int64,after string,userID int64,status string)bool{return existingID==userID&&after==status}

var _ service.AffiliatePayoutRepository=(*affiliatePayoutRepository)(nil)

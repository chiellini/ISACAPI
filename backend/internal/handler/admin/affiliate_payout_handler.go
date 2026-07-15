package admin

import (
	"strconv"
	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *AffiliateHandler) SetPayoutService(s *service.AffiliatePayoutService){h.affiliatePayoutService=s}
func (h *AffiliateHandler) ListAgents(c *gin.Context){page,pageSize:=response.ParsePagination(c);items,total,err:=h.affiliatePayoutService.AdminListAgents(c.Request.Context(),service.AffiliatePayoutListFilter{Search:c.Query("search"),Status:c.Query("status"),Page:page,PageSize:pageSize});if err!=nil{response.ErrorFrom(c,err);return};response.Paginated(c,items,total,page,pageSize)}
func (h *AffiliateHandler) UpdateAgentStatus(c *gin.Context){userID,ok:=affiliatePathID(c,"userId");if !ok{return};var req dto.UpdateAffiliateAgentStatusRequest;if err:=c.ShouldBindJSON(&req);err!=nil{response.BadRequest(c,"Invalid request: "+err.Error());return};subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"Admin not authenticated");return};if err:=h.affiliatePayoutService.AdminSetAgentStatus(c.Request.Context(),userID,req.Status,subject.UserID,c.GetHeader("Idempotency-Key"));err!=nil{response.ErrorFrom(c,err);return};item,err:=h.affiliatePayoutService.AdminGetAgent(c.Request.Context(),userID);if err!=nil{response.ErrorFrom(c,err);return};response.Success(c,item)}
func (h *AffiliateHandler) ListWithdrawals(c *gin.Context){page,pageSize:=response.ParsePagination(c);items,total,err:=h.affiliatePayoutService.AdminListWithdrawals(c.Request.Context(),service.AffiliatePayoutListFilter{Search:c.Query("search"),Status:c.Query("status"),Page:page,PageSize:pageSize});if err!=nil{response.ErrorFrom(c,err);return};response.Paginated(c,items,total,page,pageSize)}
func (h *AffiliateHandler) GetWithdrawal(c *gin.Context){id,ok:=affiliatePathID(c,"id");if !ok{return};item,err:=h.affiliatePayoutService.AdminGetWithdrawal(c.Request.Context(),id);if err!=nil{response.ErrorFrom(c,err);return};c.Header("Cache-Control","no-store");c.Header("Pragma","no-cache");response.Success(c,item)}
func (h *AffiliateHandler) ApproveWithdrawal(c *gin.Context){h.runWithdrawalAdminAction(c,func(id,operator int64)(any,error){return h.affiliatePayoutService.AdminApproveWithdrawal(c.Request.Context(),id,operator,c.GetHeader("Idempotency-Key"))})}
func (h *AffiliateHandler) RejectWithdrawal(c *gin.Context){var req dto.RejectAffiliateWithdrawalRequest;if err:=c.ShouldBindJSON(&req);err!=nil{response.BadRequest(c,"Invalid request: "+err.Error());return};h.runWithdrawalAdminAction(c,func(id,operator int64)(any,error){return h.affiliatePayoutService.AdminRejectWithdrawal(c.Request.Context(),id,operator,req.Reason,c.GetHeader("Idempotency-Key"))})}
func (h *AffiliateHandler) MarkWithdrawalPaid(c *gin.Context){var req dto.MarkAffiliateWithdrawalPaidRequest;if err:=c.ShouldBindJSON(&req);err!=nil{response.BadRequest(c,"Invalid request: "+err.Error());return};h.runWithdrawalAdminAction(c,func(id,operator int64)(any,error){return h.affiliatePayoutService.AdminMarkWithdrawalPaid(c.Request.Context(),id,operator,service.AffiliateMarkPaidInput{ActualCurrency:req.ActualCurrency,ActualAmount:req.ActualAmount,ExchangeRate:req.ExchangeRate,ExternalReference:req.ExternalReference},c.GetHeader("Idempotency-Key"))})}
func (h *AffiliateHandler) runWithdrawalAdminAction(c *gin.Context,fn func(int64,int64)(any,error)){id,ok:=affiliatePathID(c,"id");if !ok{return};subject,ok:=middleware.GetAuthSubjectFromContext(c);if !ok{response.Unauthorized(c,"Admin not authenticated");return};item,err:=fn(id,subject.UserID);if err!=nil{response.ErrorFrom(c,err);return};response.Success(c,item)}
func affiliatePathID(c *gin.Context,name string)(int64,bool){id,err:=strconv.ParseInt(c.Param(name),10,64);if err!=nil||id<=0{response.BadRequest(c,"Invalid id");return 0,false};return id,true}

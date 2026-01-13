package futu

import (
	"errors"
	"fmt"
	"github.com/santsai/futu-go/pb"
	"strings"

	"github.com/rs/zerolog/log"
)

var (
	ErrChannelClosed = errors.New("channel is closed")
	ErrInterrupted   = errors.New("process is interrupted")
	ErrTimeout       = errors.New("timeout")

	errSHA1Mismatch = errors.New("sha1 mismatch")
)

type responseErrorCode int

const (
	err_Unmatched responseErrorCode = iota
	err_UnknownWatchlist
	err_NotSupportedInSimEnv
	err_ModifyingSysSecGroup
	err_FilterMinMaxRequired
	err_RateLimited
	err_SubQuotaExceed
	err_SubTimeTooShort
)

var (
	ErrUnknownWatchlist     = &responseError{Code: err_UnknownWatchlist}
	ErrNotSupportedInSimEnv = &responseError{Code: err_NotSupportedInSimEnv}
	ErrModifyingSysSecGroup = &responseError{Code: err_ModifyingSysSecGroup}
	ErrFilterMinMaxRequired = &responseError{Code: err_FilterMinMaxRequired}
	ErrRateLimited          = &responseError{Code: err_RateLimited}
	ErrSubQuotaExceed       = &responseError{Code: err_SubQuotaExceed}
	ErrSubTimeTooShort      = &responseError{Code: err_SubTimeTooShort}
)

type retMsgMatchType int

const (
	matchAll retMsgMatchType = iota
	matchPrefix
	matchSuffix
)

type retMsgMapping struct {
	Id        pb.ProtoId
	RetType   pb.RetType
	Msgs      []string
	Code      responseErrorCode
	MatchType retMsgMatchType
}

var retMsgMappings = []retMsgMapping{
	{Code: err_UnknownWatchlist,
		Id:      pb.ProtoId_QotGetUserSecurity,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"Unknown watchlists",
			"未知自选股分组",
		}},
	{Code: err_NotSupportedInSimEnv,
		Id:      pb.ProtoId_TrdFlowSummary,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"Paper trading does not support requesting cash flow data.",
			"模拟账户不支持查询现金流水",
		}},
	{Code: err_NotSupportedInSimEnv,
		Id:      pb.ProtoId_TrdGetOrderFee,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"Simulated trade is not supported",
			"暂时不支持模拟交易",
		}},
	{Code: err_NotSupportedInSimEnv,
		Id:      pb.ProtoId_TrdGetHistoryOrderFillList,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"Simulated trade does not support deal list",
			"模拟交易不支持成交数据",
		}},
	{Code: err_NotSupportedInSimEnv,
		Id:      pb.ProtoId_TrdGetOrderFillList,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"Simulated trade does not support deal list",
			"模拟交易不支持成交数据",
		}},
	{Code: err_ModifyingSysSecGroup,
		Id:      pb.ProtoId_QotModifyUserSecurity,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"The System grouping is not supported",
			"不支持系统分组",
		}},
	{Code: err_FilterMinMaxRequired,
		Id:      pb.ProtoId_QotStockFilter,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"The filter field needs to set the range",
			"没有给需要筛选的字段进行区间赋值",
		}},
	{Code: err_RateLimited,
		Id:      pb.ProtoId_QotStockFilter,
		RetType: pb.RetType_Failed,
		Msgs: []string{
			"条件选股频率太高，请求失败，每30秒最多10次。",
		}},
	{Code: err_SubQuotaExceed,
		Id:        pb.ProtoId_QotSub,
		RetType:   pb.RetType_Failed,
		MatchType: matchPrefix,
		Msgs: []string{
			//"订阅额度不足，订阅失败，已用订阅额度：1000/1000"
			"订阅额度不足，订阅失败，已用订阅额度：",
		}},
	{Code: err_SubTimeTooShort,
		Id:        pb.ProtoId_QotSub,
		RetType:   pb.RetType_Failed,
		MatchType: matchSuffix,
		Msgs: []string{
			//"HK.00002的KL_Day订阅时间过短，至少需要订阅1分钟"
			"订阅时间过短，至少需要订阅1分钟",
		}},
	// &{Id:ProtoId_QotStockFilter RetType:RetType_Failed Msgs:[请求个数超过限制] Code:0}
}

type responseError struct {
	Code    responseErrorCode
	ProtoId pb.ProtoId
	// below from s2c resp
	RetType pb.RetType
	RetMsg  string
	RetCode int32
}

func (err *responseError) fillErrCode() {
	if err.Code != err_Unmatched {
		return
	}

mapping_loop:
	for _, m := range retMsgMappings {

		if m.Id != pb.ProtoId_Unknown &&
			m.Id != err.ProtoId {
			continue
		}

		if m.RetType != pb.RetType_Unknown &&
			m.RetType != err.RetType {
			continue
		}

		for _, msg := range m.Msgs {
			switch m.MatchType {
			case matchAll:
				if msg == err.RetMsg {
					err.Code = m.Code
					break mapping_loop
				}
			case matchPrefix:
				if strings.HasPrefix(err.RetMsg, msg) {
					err.Code = m.Code
					break mapping_loop
				}
			case matchSuffix:
				if strings.HasSuffix(err.RetMsg, msg) {
					err.Code = m.Code
					break mapping_loop
				}
			}
		}
	}

	if err.Code == err_Unmatched {
		msg := fmt.Sprintf("%+v", err.toMapping())
		log.Error().Str("err", msg).Msg("Unmatched Error")
	}
}

func (err *responseError) toMapping() *retMsgMapping {
	return &retMsgMapping{
		Code:    err.Code,
		Id:      err.ProtoId,
		RetType: err.RetType,
		Msgs:    []string{err.RetMsg},
	}
}

func (err *responseError) Is(target error) bool {
	tgt, ok := target.(*responseError)
	if !ok {
		return false
	}

	return tgt.Code == err.Code
}

func (err *responseError) Error() string {
	return err.RetMsg
}

func (err *responseError) String() string {
	return fmt.Sprintf("%#v", err)
}

func ResponseError(id pb.ProtoId, r pb.Response) error {

	if r.GetRetType() == pb.RetType_Succeed {
		return nil
	}

	err := &responseError{
		ProtoId: id,
		RetType: r.GetRetType(),
		RetMsg:  r.GetRetMsg(),
		RetCode: r.GetErrCode(),
	}

	err.fillErrCode()

	return err
}

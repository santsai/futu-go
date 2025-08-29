package futu

import (
	"context"

	"github.com/santsai/futu-go/adapt"
	"github.com/santsai/futu-go/pb"
)

// GetGlobalState 1002 - gets the global state.
func (client *Client) GetGlobalState() (*pb.GetGlobalStateResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetGlobalStateWithContext(ctx)
}

// GetAccList 2001 - gets the trading account list.
func (client *Client) GetAccList(req *pb.TrdGetAccListRequest) ([]*pb.TrdAcc, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetAccListWithContext(ctx, req)
}

// UnlockTrade 2005 - unlocks or locks the trade.
//
// pwdMD5: MD5 of the password
//
// securityFirm: security firm
func (client *Client) UnlockTrade(pwdMD5 string, secFirm pb.SecurityFirm) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.UnlockTradeWithContext(ctx, pwdMD5, secFirm)
}

func (c *Client) LockTrade(secFirm pb.SecurityFirm) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return c.LockTradeWithContext(ctx, secFirm)
}

// SubscribeAccPush 2008 - subscribes the trading account push data.
//
// accIDList: account ID list
func (client *Client) SubscribeAccPush(accIDList []uint64) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.SubscribeAccPushWithContext(ctx, accIDList)
}

// GetFunds 2101 - gets the funds.
func (client *Client) GetFunds(req *pb.TrdGetFundsRequest) (*pb.Funds, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetFundsWithContext(ctx, req)
}

// GetPositionList 2102 - gets the position list.
func (client *Client) GetPositionList(req *pb.TrdGetPositionListRequest) ([]*pb.Position, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetPositionListWithContext(ctx, req)
}

// GetOrderList 2111 - gets the maximum available trading quantities.
//
// header: trading header
//
// orderType: order type
//
// code: security code, e.g. US.AAPL
//
// price: price
func (client *Client) GetMaxTrdQtys(req *pb.TrdGetMaxTrdQtysRequest) (*pb.MaxTrdQtys, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetMaxTrdQtysWithContext(ctx, req)
}

// GetOpenOrderList 2201 - gets the open order list.
func (client *Client) GetOpenOrderList(req *pb.TrdGetOrderListRequest) ([]*pb.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOpenOrderListWithContext(ctx, req)
}

// PlaceOrder 2202 - places an order.
//
// header: trading header
//
// trdSide: trading side
//
// orderType: order type
//
// code: security code, e.g. US.AAPL
//
// qty: quantity
//
// price: price
func (client *Client) PlaceOrder(req *pb.TrdPlaceOrderRequest) (*pb.TrdPlaceOrderResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.PlaceOrderWithContext(ctx, req)
}

// ModifyOrder 2205 - modifies an order with context.
//
// header: trading header
//
// orderID: order ID, use 0 if forAll=true
//
// modifyOrderOp: modify order operation
func (client *Client) ModifyOrder(req *pb.TrdModifyOrderRequest) (*pb.TrdModifyOrderResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.ModifyOrderWithContext(ctx, req)
}

// GetHistoryOrderList 2211 - gets the filled order list.
func (client *Client) GetOrderFillList(header *pb.TrdHeader, opts ...adapt.Option) ([]*pb.OrderFill, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOrderFillListWithContext(ctx, header, opts...)
}

// GetHistoryOrderList 2221 - gets the history order list.
func (client *Client) GetHistoryOrderList(header *pb.TrdHeader, fc *pb.TrdFilterConditions, opts ...adapt.Option) ([]*pb.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetHistoryOrderListWithContext(ctx, header, fc, opts...)
}

// GetHistoryOrderFillList 2222 - gets the history filled order list.
func (client *Client) GetHistoryOrderFillList(header *pb.TrdHeader, fc *pb.TrdFilterConditions, opts ...adapt.Option) ([]*pb.OrderFill, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetHistoryOrderFillListWithContext(ctx, header, fc, opts...)
}

// GetMarginRatio 2223 - gets the margin ratio.
func (client *Client) GetMarginRatio(header *pb.TrdHeader, codes []string) ([]*pb.MarginRatioInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetMarginRatioWithContext(ctx, header, codes)
}

// GetOrderFee 2225 - gets the order fee.
func (client *Client) GetOrderFee(header *pb.TrdHeader, orderIdExList []string) ([]*pb.OrderFee, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOrderFeeWithContext(ctx, header, orderIdExList)
}

// TrdFlowSummary 2226 - gets the trading flow summary.
func (client *Client) TrdFlowSummary(header *pb.TrdHeader, clearingDate string) ([]*pb.FlowSummaryInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.TrdFlowSummaryWithContext(ctx, header, clearingDate)
}

// Subscribe 3001 - subscribes or unsubscribes.
//
// codes: security codes
//
// subTypes: subscription types
//
// isSub: true for subscribe, false for unsubscribe
func (client *Client) Subscribe(req *pb.QotSubRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.SubscribeWithContext(ctx, req)
}

// GetSubInfo 3003 - gets the subscription information.
func (client *Client) GetSubInfo(opts ...adapt.Option) (*pb.QotGetSubInfoResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetSubInfoWithContext(ctx, opts...)
}

// GetBasicQot 3004 - gets the basic quotes of given securities.
func (client *Client) GetBasicQot(codes []string) ([]*pb.BasicQot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetBasicQotWithContext(ctx, codes)
}

// GetKL 3006 - gets K-line data.
//
// code: security code
//
// klType: K-line type
func (client *Client) GetKL(code string, klType int32, opts ...adapt.Option) (*pb.QotGetKLResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetKLWithContext(ctx, code, klType, opts...)
}

// GetRT 3008 - gets real-time data.
//
// code: security code
func (client *Client) GetRT(code string) (*pb.QotGetRTResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetRTWithContext(ctx, code)
}

// GetTicker 3010 - gets ticker data.
//
// code: security code
func (client *Client) GetTicker(code string, opts ...adapt.Option) (*pb.QotGetTickerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetTickerWithContext(ctx, code, opts...)
}

// GetOrderBook 3012 - gets order book data.
//
// code: security code
func (client *Client) GetOrderBook(code string, opts ...adapt.Option) (*pb.QotGetOrderBookResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOrderBookWithContext(ctx, code, opts...)
}

// GetBroker 3014 - gets broker data.
//
// code: security code
func (client *Client) GetBroker(code string) (*pb.QotGetBrokerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetBrokerWithContext(ctx, code)
}

// RequestHistoryKL 3103 - requests the history K-line data.
//
// code: security code
//
// klType: K-line type
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) RequestHistoryKL(code string, klType pb.KLType, beginTime string, endTime string, opts ...adapt.Option) (*pb.QotRequestHistoryKLResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.RequestHistoryKLWithContext(ctx, code, klType, beginTime, endTime, opts...)
}

// RequestHistoryKLQuota 3104 - requests the history K-line quota.
func (client *Client) RequestHistoryKLQuota(opts ...adapt.Option) (*pb.QotRequestHistoryKLQuotaResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.RequestHistoryKLQuotaWithContext(ctx, opts...)
}

// RequestRehab 3105 - requests the rehab data.
func (client *Client) RequestRehab(code string) (*pb.QotRequestRehabResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.RequestRehabWithContext(ctx, code)
}

// GetStaticInfo 3202 - gets the static information.
func (client *Client) GetStaticInfo(opts ...adapt.Option) ([]*pb.SecurityStaticInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetStaticInfoWithContext(ctx, opts...)
}

// GetSecuritySnapshot 3203 - gets the security snapshot.
//
// codes: security codes
func (client *Client) GetSecuritySnapshot(codes []string) ([]*pb.Snapshot, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetSecuritySnapshotWithContext(ctx, codes)
}

// GetPlateSet 3204 - gets the plate set.
//
// market: market
//
// plateSetType: plate set type
func (client *Client) GetPlateSet(market pb.QotMarket, plateSetType pb.PlateSetType) ([]*pb.PlateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetPlateSetWithContext(ctx, market, plateSetType)
}

// GetPlateSecurity 3205 - gets the plate securities.
//
// plateCode: plate code
func (client *Client) GetPlateSecurity(plateCode string, opts ...adapt.Option) ([]*pb.SecurityStaticInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetPlateSecurityWithContext(ctx, plateCode, opts...)
}

// GetReference 3206 - gets the reference data.
//
// code: security code
//
// refType: reference type
func (client *Client) GetReference(code string, refType pb.ReferenceType) ([]*pb.SecurityStaticInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetReferenceWithContext(ctx, code, refType)
}

// GetOwnerPlate 3207 - gets the owner plate.
//
// codes: security codes
func (client *Client) GetOwnerPlate(codes []string) ([]*pb.SecurityOwnerPlate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOwnerPlateWithContext(ctx, codes)
}

// GetOptionChain 3209 - gets the option chain with context.
//
// code: security code
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) GetOptionChain(code string, beginTime string, endTime string, opts ...adapt.Option) ([]*pb.OptionChain, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOptionChainWithContext(ctx, code, beginTime, endTime, opts...)
}

// GetWarrant 3210 - gets the warrant, only available in Hong Kong market.
// Sort by score in descending order by default.
//
// begin: begin index
//
// num: number of warrants
func (client *Client) GetWarrant(begin int32, num int32, opts ...adapt.Option) (*pb.QotGetWarrantResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetWarrantWithContext(ctx, begin, num, opts...)
}

// GetCapitalFlow 3211 - gets the capital flow.
//
// code: security code
func (client *Client) GetCapitalFlow(code string, opts ...adapt.Option) (*pb.QotGetCapitalFlowResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetCapitalFlowWithContext(ctx, code, opts...)
}

// GetCapitalDistribution 3212 - gets the capital distribution.
//
// code: security code
func (client *Client) GetCapitalDistribution(code string) (*pb.QotGetCapitalDistributionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetCapitalDistributionWithContext(ctx, code)
}

// GetUserSecurity 3213 - gets the user security.
//
// groupName: group name
func (client *Client) GetUserSecurity(groupName string) ([]*pb.SecurityStaticInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetUserSecurityWithContext(ctx, groupName)
}

// ModifyUserSecurity 3214 - modifies the user security.
//
// groupName: group name
//
// codes: security codes
//
// op: operation, 1 for add, 2 for delete
func (client *Client) ModifyUserSecurity(groupName string, codes []string, op pb.ModifyUserSecurityOp) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.ModifyUserSecurityWithContext(ctx, groupName, codes, op)
}

// StockFilter 3215 - filters the stocks.
//
// market: market
func (client *Client) StockFilter(market int32, opts ...adapt.Option) (*pb.QotStockFilterResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.StockFilterWithContext(ctx, market, opts...)
}

// GetIpoList 3217 - gets the IPO list.
//
// market: market
func (client *Client) GetIpoList(market pb.QotMarket) ([]*pb.IpoData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetIpoListWithContext(ctx, market)
}

// GetFutureInfo 3218 - gets the future information.
//
// codes: security codes
func (client *Client) GetFutureInfo(codes []string) ([]*pb.FutureInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetFutureInfoWithContext(ctx, codes)
}

// RequestTradeDate 3219 - requests the trade date.
//
// market: market
//
// code: security code
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) RequestTradeDate(market pb.TradeDateMarket, code string, beginTime, endTime string) ([]*pb.TradeDate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.RequestTradeDateWithContext(ctx, market, code, beginTime, endTime)
}

// SetPriceReminder 3220 - sets the price reminder.
//
// code: security code
//
// op: operation, 1 for add, 2 for delete
func (client *Client) SetPriceReminder(code string, op pb.SetPriceReminderOp, opts ...adapt.Option) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.SetPriceReminderWithContext(ctx, code, op, opts...)
}

// GetPriceReminder 3221 - gets the price reminder.
//
// code: security code
//
// market: market, if security is set, this param is ignored
func (client *Client) GetPriceReminder(code string, market pb.QotMarket) ([]*pb.PriceReminder, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetPriceReminderWithContext(ctx, code, market)
}

// GetUserSecurityGroup 3222 - gets the user security group.
//
// groupType: group type
func (client *Client) GetUserSecurityGroup(groupType pb.GroupType) ([]*pb.GroupData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetUserSecurityGroupWithContext(ctx, groupType)
}

// GetMarketState 3223 - gets the market state.
//
// codes: security codes
func (client *Client) GetMarketState(codes []string) ([]*pb.MarketInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetMarketStateWithContext(ctx, codes)
}

// GetOptionExpirationDate 3224 - gets the option expiration date.
//
// code: security code
func (client *Client) GetOptionExpirationDate(code string, opts ...adapt.Option) ([]*pb.OptionExpirationDate, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	return client.GetOptionExpirationDateWithContext(ctx, code, opts...)
}

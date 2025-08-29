package futu

import (
	"context"

	"github.com/santsai/futu-go/adapt"
	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

// GetGlobalStateWithContext 1002 - gets the global state with context.
func (client *Client) GetGlobalStateWithContext(ctx context.Context) (*pb.GetGlobalStateResponse, error) {
	req := &pb.GetGlobalStateRequest{
		//		UserID: proto.Uint64(0),
	}

	return req.Dispatch(ctx, client)
}

// GetAccListWithContext 2001 - gets the account list with context.
func (client *Client) GetAccListWithContext(ctx context.Context, req *pb.TrdGetAccListRequest) ([]*pb.TrdAcc, error) {

	//	req.UserID = proto.Uint64(0)
	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetAccList(), nil
}

// UnlockTrade 2005 - unlocks or locks the trade.
//
// pwdMD5: MD5 of the password
//
// securityFirm: security firm
func (client *Client) UnlockTradeWithContext(ctx context.Context, pwdMD5 string, secFirm pb.SecurityFirm) error {
	req := &pb.TrdUnlockTradeRequest{
		Unlock:       proto.Bool(true),
		PwdMD5:       proto.String(pwdMD5),
		SecurityFirm: secFirm.Enum(),
	}

	_, err := req.Dispatch(ctx, client)
	return err
}

func (c *Client) LockTradeWithContext(ctx context.Context, secFirm pb.SecurityFirm) error {
	req := &pb.TrdUnlockTradeRequest{
		Unlock:       proto.Bool(false),
		SecurityFirm: secFirm.Enum(),
	}

	_, err := req.Dispatch(ctx, c)
	return err
}

// SubscribeAccPushWithContext 2008 - subscribes the trading account push data.
//
// accIDList: account ID list
func (client *Client) SubscribeAccPushWithContext(ctx context.Context, accIDList []uint64) error {
	req := &pb.TrdSubAccPushRequest{
		AccIDList: accIDList,
	}

	_, err := req.Dispatch(ctx, client)
	return err
}

// GetFundsWithContext 2101 - gets the funds with context.
func (client *Client) GetFundsWithContext(ctx context.Context, req *pb.TrdGetFundsRequest) (*pb.Funds, error) {

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetFunds(), nil
}

// GetPositionListWithContext 2102 - gets the position list with context.
func (client *Client) GetPositionListWithContext(ctx context.Context, req *pb.TrdGetPositionListRequest) ([]*pb.Position, error) {

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetPositionList(), nil
}

// GetMaxTrdQtysWithContext 2111 - gets the maximum available trading quantities with context.
//
// header: trading header
//
// orderType: order type
//
// code: security code, e.g. AAPL
//
// price: price
func (client *Client) GetMaxTrdQtysWithContext(ctx context.Context, req *pb.TrdGetMaxTrdQtysRequest) (*pb.MaxTrdQtys, error) {

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetMaxTrdQtys(), nil
}

// GetOpenOrderListWithContext 2201 - gets the open order list with context.
func (client *Client) GetOpenOrderListWithContext(ctx context.Context, req *pb.TrdGetOrderListRequest) ([]*pb.Order, error) {

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetOrderList(), nil
	}
}

// PlaceOrderWithContext 2202 - places an order with context.
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
func (client *Client) PlaceOrderWithContext(ctx context.Context, req *pb.TrdPlaceOrderRequest) (*pb.TrdPlaceOrderResponse, error) {

	//	req.PacketID = client.nextTradePacketId()
	return req.Dispatch(ctx, client)
}

// ModifyOrderWithContext 2205 - modifies an order with context.
//
// header: trading header
//
// orderID: order ID, use 0 if forAll=true
//
// modifyOrderOp: modify order operation
func (client *Client) ModifyOrderWithContext(ctx context.Context, req *pb.TrdModifyOrderRequest) (*pb.TrdModifyOrderResponse, error) {

	//	req.PacketID = client.nextTradePacketId()
	return req.Dispatch(ctx, client)
}

// GetOrderFillListWithContext 2211 - gets the filled order list with context.
func (client *Client) GetOrderFillListWithContext(ctx context.Context, header *pb.TrdHeader, opts ...adapt.Option) ([]*pb.OrderFill, error) {
	o := adapt.NewOptions(opts...)
	o["header"] = header

	var req pb.TrdGetOrderFillListRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetOrderFillList(), nil
	}
}

// GetHistoryOrderListWithContext 2221 - gets the history order list with context.
func (client *Client) GetHistoryOrderListWithContext(ctx context.Context, header *pb.TrdHeader, fc *pb.TrdFilterConditions, opts ...adapt.Option) ([]*pb.Order, error) {
	o := adapt.NewOptions(opts...)
	o["header"] = header
	o["filterConditions"] = fc

	var req pb.TrdGetHistoryOrderListRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetOrderList(), nil
	}
}

// GetHistoryOrderFillListWithContext 2222 - gets the history filled order list with context.
func (client *Client) GetHistoryOrderFillListWithContext(ctx context.Context, header *pb.TrdHeader, fc *pb.TrdFilterConditions, opts ...adapt.Option) ([]*pb.OrderFill, error) {
	o := adapt.NewOptions(opts...)
	o["header"] = header
	o["filterConditions"] = fc

	var req pb.TrdGetHistoryOrderFillListRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetOrderFillList(), nil
	}
}

// GetMarginRatioWithContext 2223 - gets the margin ratio with context.
func (client *Client) GetMarginRatioWithContext(ctx context.Context, header *pb.TrdHeader, codes []string) ([]*pb.MarginRatioInfo, error) {
	req := &pb.TrdGetMarginRatioRequest{
		Header:       header,
		SecurityList: adapt.NewSecurities(codes),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetMarginRatioInfoList(), nil
	}
}

// GetOrderFeeWithContext 2225 - gets the order fee with context.
func (client *Client) GetOrderFeeWithContext(ctx context.Context, header *pb.TrdHeader, orderIdExList []string) ([]*pb.OrderFee, error) {
	req := &pb.TrdGetOrderFeeRequest{
		Header:        header,
		OrderIdExList: orderIdExList,
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetOrderFeeList(), nil
	}
}

// TrdFlowSummaryWithContext 2226 - gets the trading flow summary with context.
func (client *Client) TrdFlowSummaryWithContext(ctx context.Context, header *pb.TrdHeader, clearingDate string) ([]*pb.FlowSummaryInfo, error) {
	req := &pb.TrdFlowSummaryRequest{
		Header:       header,
		ClearingDate: proto.String(clearingDate),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetFlowSummaryInfoList(), nil
	}
}

// SubscribeWithContext 3001 - subscribes or unsubscribes with context.
func (client *Client) SubscribeWithContext(ctx context.Context, req *pb.QotSubRequest) error {

	_, err := req.Dispatch(ctx, client)
	return err
}

// GetSubInfoWithContext 3003 - gets the subscription information with context.
func (client *Client) GetSubInfoWithContext(ctx context.Context, opts ...adapt.Option) (*pb.QotGetSubInfoResponse, error) {
	o := adapt.NewOptions(opts...)
	var req pb.QotGetSubInfoRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetBasicQotWithContext 3004 - gets the basic quotes of given securities with context.
func (client *Client) GetBasicQotWithContext(ctx context.Context, codes []string) ([]*pb.BasicQot, error) {
	req := &pb.QotGetBasicQotRequest{
		SecurityList: adapt.NewSecurities(codes),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetBasicQotList(), nil
	}
}

// GetKLWithContext 3006 - gets K-line data with context.
func (client *Client) GetKLWithContext(ctx context.Context, code string, klType int32, opts ...adapt.Option) (*pb.QotGetKLResponse, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)
	o["klType"] = klType

	if _, ok := o["rehabType"]; !ok {
		o["rehabType"] = pb.RehabType_None
	}

	if _, ok := o["reqNum"]; !ok {
		o["reqNum"] = 1000
	}

	var req pb.QotGetKLRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetRTWithContext 3008 - gets real-time data with context.
//
// code: security code
func (client *Client) GetRTWithContext(ctx context.Context, code string) (*pb.QotGetRTResponse, error) {
	req := &pb.QotGetRTRequest{
		Security: adapt.NewSecurity(code),
	}

	return req.Dispatch(ctx, client)
}

// GetTickerWithContext 3010 - gets the ticker data with context.
//
// code: security code
func (client *Client) GetTickerWithContext(ctx context.Context, code string, opts ...adapt.Option) (*pb.QotGetTickerResponse, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)

	if _, ok := o["maxRetNum"]; !ok {
		o["maxRetNum"] = 1000
	}

	var req pb.QotGetTickerRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetOrderBookWithContext 3012 - gets the order book with context.
//
// code: security code
func (client *Client) GetOrderBookWithContext(ctx context.Context, code string, opts ...adapt.Option) (*pb.QotGetOrderBookResponse, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)

	if _, ok := o["num"]; !ok {
		o["num"] = 100
	}

	var req pb.QotGetOrderBookRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetBrokerWithContext 3014 - gets the broker with context.
//
// code: security code
func (client *Client) GetBrokerWithContext(ctx context.Context, code string) (*pb.QotGetBrokerResponse, error) {
	req := &pb.QotGetBrokerRequest{
		Security: adapt.NewSecurity(code),
	}

	return req.Dispatch(ctx, client)
}

// RequestHistoryKLWithContext 3103 - requests the history K-line data with context.
//
// code: security code
//
// klType: K-line type
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) RequestHistoryKLWithContext(ctx context.Context, code string, klType pb.KLType, beginTime string, endTime string, opts ...adapt.Option) (*pb.QotRequestHistoryKLResponse, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)
	o["klType"] = klType
	o["beginTime"] = beginTime
	o["endTime"] = endTime

	if _, ok := o["rehabType"]; !ok {
		o["rehabType"] = pb.RehabType_None
	}

	var req pb.QotRequestHistoryKLRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// RequestHistoryKLQuotaWithContext 3104 - requests the history K-line quota with context.
func (client *Client) RequestHistoryKLQuotaWithContext(ctx context.Context, opts ...adapt.Option) (*pb.QotRequestHistoryKLQuotaResponse, error) {
	o := adapt.NewOptions(opts...)

	var req pb.QotRequestHistoryKLQuotaRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// RequestRehabWithContext 3105 - requests the rehab data with context.
//
// code: security code
func (client *Client) RequestRehabWithContext(ctx context.Context, code string) (*pb.QotRequestRehabResponse, error) {
	req := &pb.QotRequestRehabRequest{
		Security: adapt.NewSecurity(code),
	}

	return req.Dispatch(ctx, client)
}

// GetStaticInfoWithContext 3202 - gets the static information with context.
func (client *Client) GetStaticInfoWithContext(ctx context.Context, opts ...adapt.Option) ([]*pb.SecurityStaticInfo, error) {
	o := adapt.NewOptions(opts...)

	var req pb.QotGetStaticInfoRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetStaticInfoList(), nil
	}
}

// GetSecuritySnapshotWithContext 3203 - gets the security snapshot with context.
//
// codes: security codes
func (client *Client) GetSecuritySnapshotWithContext(ctx context.Context, codes []string) ([]*pb.Snapshot, error) {
	req := &pb.QotGetSecuritySnapshotRequest{
		SecurityList: adapt.NewSecurities(codes),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetSnapshotList(), nil
}

// GetPlateSetWithContext 3204 - gets the plate set with context.
//
// market: market
//
// plateSetType: plate set type
func (client *Client) GetPlateSetWithContext(ctx context.Context, market pb.QotMarket, plateSetType pb.PlateSetType) ([]*pb.PlateInfo, error) {
	req := &pb.QotGetPlateSetRequest{
		Market:       market.Enum(),
		PlateSetType: plateSetType.Enum(),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetPlateInfoList(), nil
	}
}

// GetPlateSecurityWithContext 3205 - gets the plate securities with context.
//
// plateCode: plate code
func (client *Client) GetPlateSecurityWithContext(ctx context.Context, plateCode string, opts ...adapt.Option) ([]*pb.SecurityStaticInfo, error) {
	o := adapt.NewOptions(opts...)
	o["plate"] = adapt.NewSecurity(plateCode)

	var req pb.QotGetPlateSecurityRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetStaticInfoList(), nil
}

// GetReferenceWithContext 3206 - gets the reference with context.
//
// code: security code
//
// refType: reference type
func (client *Client) GetReferenceWithContext(ctx context.Context, code string, refType pb.ReferenceType) ([]*pb.SecurityStaticInfo, error) {
	req := &pb.QotGetReferenceRequest{
		Security:      adapt.NewSecurity(code),
		ReferenceType: refType.Enum(),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetStaticInfoList(), nil
}

// GetOwnerPlateWithContext 3207 - gets the owner plate with context.
//
// codes: security codes
func (client *Client) GetOwnerPlateWithContext(ctx context.Context, codes []string) ([]*pb.SecurityOwnerPlate, error) {
	req := &pb.QotGetOwnerPlateRequest{
		SecurityList: adapt.NewSecurities(codes),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetOwnerPlateList(), nil
}

// GetOptionChainWithContext 3209 - gets the option chain with context.
//
// code: security code
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) GetOptionChainWithContext(ctx context.Context, code string, beginTime string, endTime string, opts ...adapt.Option) ([]*pb.OptionChain, error) {
	o := adapt.NewOptions(opts...)
	o["owner"] = adapt.NewSecurity(code)
	o["beginTime"] = beginTime
	o["endTime"] = endTime

	var req pb.QotGetOptionChainRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetOptionChain(), nil
}

// GetWarrantWithContext 3210 - gets the warrant with context, only available in Hong Kong market.
// Sort by score in descending order by default.
//
// begin: begin index
//
// num: number of warrants
func (client *Client) GetWarrantWithContext(ctx context.Context, begin int32, num int32, opts ...adapt.Option) (*pb.QotGetWarrantResponse, error) {
	o := adapt.NewOptions(opts...)
	o["begin"] = begin
	o["num"] = num

	if _, ok := o["sortField"]; !ok {
		o["sortField"] = pb.SortField_Score
	}

	if _, ok := o["ascend"]; !ok {
		o["ascend"] = false
	}

	var req pb.QotGetWarrantRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetCapitalFlowWithContext 3211 - gets the capital flow with context.
//
// code: security code
func (client *Client) GetCapitalFlowWithContext(ctx context.Context, code string, opts ...adapt.Option) (*pb.QotGetCapitalFlowResponse, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)

	if _, ok := o["periodType"]; !ok {
		o["periodType"] = pb.PeriodType_INTRADAY
	}

	var req pb.QotGetCapitalFlowRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetCapitalDistributionWithContext 3212 - gets the capital distribution with context.
//
// code: security code
func (client *Client) GetCapitalDistributionWithContext(ctx context.Context, code string) (*pb.QotGetCapitalDistributionResponse, error) {
	req := &pb.QotGetCapitalDistributionRequest{
		Security: adapt.NewSecurity(code),
	}

	return req.Dispatch(ctx, client)
}

// GetUserSecurityWithContext 3213 - gets the user security with context.
//
// groupName: group name
func (client *Client) GetUserSecurityWithContext(ctx context.Context, groupName string) ([]*pb.SecurityStaticInfo, error) {
	req := &pb.QotGetUserSecurityRequest{
		GroupName: proto.String(groupName),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetStaticInfoList(), nil
}

// ModifyUserSecurityWithContext 3214 - modifies the user security with context.
//
// groupName: group name
//
// codes: security codes
//
// op: operation
func (client *Client) ModifyUserSecurityWithContext(ctx context.Context, groupName string, codes []string, op pb.ModifyUserSecurityOp) error {
	req := &pb.QotModifyUserSecurityRequest{
		GroupName:    proto.String(groupName),
		SecurityList: adapt.NewSecurities(codes),
		Op:           op.Enum(),
	}

	_, err := req.Dispatch(ctx, client)
	return err
}

// StockFilterWithContext 3215 - filters the stock with context.
//
// market: market
func (client *Client) StockFilterWithContext(ctx context.Context, market int32, opts ...adapt.Option) (*pb.QotStockFilterResponse, error) {
	o := adapt.NewOptions(opts...)
	o["market"] = market

	if _, ok := o["begin"]; !ok {
		o["begin"] = 0
	}

	if _, ok := o["num"]; !ok {
		o["num"] = 200
	}

	var req pb.QotStockFilterRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	return req.Dispatch(ctx, client)
}

// GetIpoListWithContext 3217 - gets the IPO list with context.
//
// market: market
func (client *Client) GetIpoListWithContext(ctx context.Context, market pb.QotMarket) ([]*pb.IpoData, error) {
	req := &pb.QotGetIpoListRequest{
		Market: market.Enum(),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetIpoList(), nil
}

// GetFutureInfoWithContext 3218 - gets the future information with context.
//
// codes: security codes
func (client *Client) GetFutureInfoWithContext(ctx context.Context, codes []string) ([]*pb.FutureInfo, error) {
	req := &pb.QotGetFutureInfoRequest{
		SecurityList: adapt.NewSecurities(codes),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetFutureInfoList(), nil
}

// RequestTradeDateWithContext 3219 - requests the trade date with context.
//
// market: market
//
// code: security code
//
// beginTime: begin time, format: "yyyy-MM-dd"
//
// endTime: end time, format: "yyyy-MM-dd"
func (client *Client) RequestTradeDateWithContext(ctx context.Context, market pb.TradeDateMarket, code string, beginTime, endTime string) ([]*pb.TradeDate, error) {
	req := &pb.QotRequestTradeDateRequest{
		Market:    market.Enum(),
		BeginTime: proto.String(beginTime),
		EndTime:   proto.String(endTime),
	}
	if code != "" {
		req.Security = adapt.NewSecurity(code)
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetTradeDateList(), nil
}

// SetPriceReminderWithContext 3220 - sets the price reminder with context.
//
// code: security code
//
// op: operation
func (client *Client) SetPriceReminderWithContext(ctx context.Context, code string, op pb.SetPriceReminderOp, opts ...adapt.Option) (int64, error) {
	o := adapt.NewOptions(opts...)
	o["security"] = adapt.NewSecurity(code)
	o["op"] = op

	var req pb.QotSetPriceReminderRequest
	if err := o.ToProto(&req); err != nil {
		return 0, err
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return 0, err
	}

	return resp.GetKey(), nil
}

// GetPriceReminderWithContext 3221 - gets the price reminder with context.
//
// code: security code
//
// market: market, if security is set, this param is ignored
func (client *Client) GetPriceReminderWithContext(ctx context.Context, code string, market pb.QotMarket) ([]*pb.PriceReminder, error) {
	req := &pb.QotGetPriceReminderRequest{
		Security: adapt.NewSecurity(code),
		Market:   market.Enum(),
	}

	resp, err := req.Dispatch(ctx, client)
	if err != nil {
		return nil, err
	}

	return resp.GetPriceReminderList(), nil
}

// GetUserSecurityGroupWithContext 3222 - gets the user security group with context.
//
// groupType: group type
func (client *Client) GetUserSecurityGroupWithContext(ctx context.Context, groupType pb.GroupType) ([]*pb.GroupData, error) {
	req := &pb.QotGetUserSecurityGroupRequest{
		GroupType: groupType.Enum(),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetGroupList(), nil
	}
}

// GetMarketStateWithContext 3223 - gets the market state with context.
//
// codes: security codes
func (client *Client) GetMarketStateWithContext(ctx context.Context, codes []string) ([]*pb.MarketInfo, error) {
	req := &pb.QotGetMarketStateRequest{
		SecurityList: adapt.NewSecurities(codes),
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetMarketInfoList(), nil
	}
}

// GetOptionExpirationDateWithContext 3224 - gets the option expiration date with context.
//
// code: security code
func (client *Client) GetOptionExpirationDateWithContext(ctx context.Context, code string, opts ...adapt.Option) ([]*pb.OptionExpirationDate, error) {
	o := adapt.NewOptions(opts...)
	o["owner"] = adapt.NewSecurity(code)

	var req pb.QotGetOptionExpirationDateRequest
	if err := o.ToProto(&req); err != nil {
		return nil, err
	}

	if resp, err := req.Dispatch(ctx, client); err != nil {
		return nil, err
	} else {
		return resp.GetDateList(), nil
	}
}

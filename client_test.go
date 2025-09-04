package futu_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/santsai/futu-go"
	"github.com/santsai/futu-go/pb"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"
)

var privateKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDMY2rkpYPQG+UQYs1/pVLa+gA6qY3j9vPO7rkfFwtQl+HwnWZO
koKjB4plJXDhCBW8KYr+8HLObS7B2v7bPxMASbZcifBEzBDDarXlXe7U9rId1YkK
k+zgt/+VmLsb/Pn6H2FunpBEaGMXQ4JjMKP7l3KKCYLR1qnWfH9vqW2/jQIDAQAB
AoGAWiX3fcXrL50+GU4a6ZQXyXG/LCCg7s17l+f/StE1MmTrSN+DwmAospOe3SH4
eajBPW7tWajgWmAZ4XMgOb/GNu12/n0TTzomGULSaGwi73e7VFExFVvRmrY7zCx4
hmx5yg1i4IquACPfY7zaDUOYNOrVdJ3hHSCalZvWqfaSmUECQQD9SjaidkJszhue
FoCqmCuO2Hh71ttpo8uovBooaTf+EHDIZyjp/JcNA4+GNLrzPLFbq3XIus87NGkp
7g0JpY4lAkEAzpNCA3JRvzuukzgVIXsnNlYhKB+yP38i6YyLAXfk1kufxWzR+as6
R4WCbG32w0wvaE3FPjuqGZey8mvsxXgrSQJBAMChml+ANRBux845Ku2TAT2IIEl+
pCv5aEARnosxSmYstrmSyyj48x/wn0zf+XZXqEMhaViZylUqjPhYlQ3LHQkCQGS4
ZA1uJfGJ1fqt84+Zjmrt38jCe5R+FrWs8vHKVWcvBD2sa0zCce4BaLAZhaF/efXv
RWasjKlhz7xnZtB5YRECQCqIQLTNMzHBTiCCHgbJvE3C9+uzpb+El4oleM6n77IW
VcwwDz6DhmTnlMjfKeJN6MgJmYnKSDt+rmheQD1bw7U=
-----END RSA PRIVATE KEY-----`)

type ClientTestSuite struct {
	suite.Suite
	sdk             *futu.Client
	accList         []*pb.TrdAcc
	usAccountHeader *pb.TrdHeader
}

// TestClientTestSuite runs the http client test suite
func TestClientTestSuite(t *testing.T) {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	suite.Run(t, new(ClientTestSuite))
}

// SetupSuite run once at the very start of the testing suite, before any tests are run.
func (ts *ClientTestSuite) SetupSuite() {
	var err error

	ts.sdk, err = futu.NewClient(
		futu.WithPrivateKey(privateKey),
		futu.WithTimeout(15*time.Second),
	)
	if err != nil {
		log.Error().Err(err).Msg("new client error")
		ts.T().SkipNow()
	}
	ts.sdk.RegisterHandler(pb.ProtoId_Notify, func(s2c proto.Message) error {
		msg := s2c.(*pb.NotifyResponse)
		log.Info().Interface("s2c", msg).Msg("custom handler")
		return nil
	})

	//
	accListReq := &pb.TrdGetAccListRequest{
		NeedGeneralSecAccount: proto.Bool(true),
	}
	if resp, err := accListReq.Dispatch(context.TODO(), ts.sdk); err != nil {
		log.Error().Err(err).Msg("GetAccList error")
		ts.T().SkipNow()
	} else {
		ts.accList = resp.GetAccList()
		for _, acc := range ts.accList {
			log.Info().Str("acc", fmt.Sprintf("%+v", acc)).Msg("GetAccList")
		}
	}

	//
	acc := futu.FindAccount(ts.accList, pb.TrdMarket_US, pb.TrdAccType_Margin, pb.SimAccType_Stock)
	if acc == nil {
		ts.T().Skip("no suitable account found")
	}

	log.Info().Str("acc", fmt.Sprintf("%+v", acc)).Msg("FindSimAccount")
	ts.usAccountHeader = &pb.TrdHeader{
		TrdEnv:    pb.TrdEnv_Simulate.Enum(),
		AccID:     proto.Uint64(acc.GetAccID()),
		TrdMarket: pb.TrdMarket_US.Enum(),
	}

	//
	subReq := &pb.QotSubRequest{
		SecurityList: futu.NewSecurities(
			"HK.09988", "HK.00700",
		),

		SubTypeList: []pb.SubType{
			pb.SubType_Basic,
			pb.SubType_RT,
			pb.SubType_KL_Day,
			pb.SubType_KL_3Min,
			pb.SubType_Ticker,
			pb.SubType_OrderBook,
			pb.SubType_Broker,
		},

		IsSubOrUnSub: proto.Bool(true),
	}

	if _, err := subReq.Dispatch(context.TODO(), ts.sdk); err != nil {
		log.Error().Err(err).Msg("QotSubRequest")
	}
}

// TearDownSuite run once at the very end of the testing suite, after all tests have been run.
func (ts *ClientTestSuite) TearDownSuite() {
	if ts.sdk != nil {
		ts.sdk.Close()
	}
}

func (ts *ClientTestSuite) TestGetGlobalState() {
	should := require.New(ts.T())

	req := &pb.GetGlobalStateRequest{}
	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)

	fmt.Println(resp)
}

func (ts *ClientTestSuite) TestLockTrade() {
	should := require.New(ts.T())

	req := &pb.TrdUnlockTradeRequest{
		Unlock:       proto.Bool(false),
		SecurityFirm: pb.SecurityFirm_FutuSecurities.Enum(),
	}
	_, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
}

func (ts *ClientTestSuite) TestSubscribeAccPush() {
	should := require.New(ts.T())

	req := &pb.TrdSubAccPushRequest{
		AccIDList: []uint64{ts.usAccountHeader.GetAccID()},
	}

	_, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
}

func (ts *ClientTestSuite) TestGetFunds() {
	should := require.New(ts.T())

	req := &pb.TrdGetFundsRequest{
		Header: ts.usAccountHeader,
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	log.Info().Interface("data", resp.GetFunds()).Msg("GetFunds")
}

func (ts *ClientTestSuite) TestGetPositionList() {
	should := require.New(ts.T())

	req := &pb.TrdGetPositionListRequest{
		Header: ts.usAccountHeader,
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	for _, pos := range resp.GetPositionList() {
		log.Info().Interface("position", pos).Msg("GetPositionList")
	}
}

func (ts *ClientTestSuite) TestGetMaxTrdQtys() {
	should := require.New(ts.T())

	req := &pb.TrdGetMaxTrdQtysRequest{
		Header:    ts.usAccountHeader,
		OrderType: pb.OrderType_Normal.Enum(),
		Code:      proto.String("AAPL"),
		Price:     proto.Float64(200),
		SecMarket: pb.TrdSecMarket_US.Enum(),
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	log.Info().Interface("data", resp.GetMaxTrdQtys()).Msg("GetMaxTrdQtys")
}

func (ts *ClientTestSuite) TestGetOpenOrderList() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderListRequest{
		Header:           ts.usAccountHeader,
		FilterStatusList: []pb.OrderStatus{pb.OrderStatus_Submitted},
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	for _, order := range resp.GetOrderList() {
		log.Info().Interface("open order", order).Msg("GetOpenOrderList")
	}
}

func (ts *ClientTestSuite) TestPlaceOrderAndModifyOrder() {
	should := require.New(ts.T())

	orderReq := &pb.TrdPlaceOrderRequest{
		Header:    ts.usAccountHeader,
		TrdSide:   pb.TrdSide_Buy.Enum(),
		OrderType: pb.OrderType_Market.Enum(),
		Code:      proto.String("AAPL"),
		SecMarket: pb.TrdSecMarket_US.Enum(),
		Qty:       proto.Float64(1),
		Remark:    proto.String("go sdk"),
	}
	resp, err := orderReq.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	log.Info().Interface("result", resp).Msg("PlaceOrder")

	// cancel the order
	cancelReq := &pb.TrdModifyOrderRequest{
		Header:        ts.usAccountHeader,
		OrderID:       proto.Uint64(resp.GetOrderID()),
		ModifyOrderOp: pb.ModifyOrderOp_Cancel.Enum(),
	}
	cancelResp, err := cancelReq.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	log.Info().Interface("result", cancelResp).Msg("ModifyOrder")
}

func (ts *ClientTestSuite) TestGetOrderFillList() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderFillListRequest{
		Header: ts.usAccountHeader,
	}

	_, err := req.Dispatch(context.TODO(), ts.sdk)
	should.Error(err) // 模拟交易不支持成交数据
}

func (ts *ClientTestSuite) TestGetHistoryOrderList() {
	should := require.New(ts.T())

	now := time.Now()
	req := &pb.TrdGetHistoryOrderListRequest{
		Header: ts.usAccountHeader,
		FilterConditions: &pb.TrdFilterConditions{
			BeginTime: futu.DateTimePtr(now.AddDate(0, 0, -7)),
			EndTime:   futu.DateTimePtr(now),
		},
		FilterStatusList: []pb.OrderStatus{pb.OrderStatus_Filled_All},
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	for _, order := range resp.GetOrderList() {
		log.Info().Interface("history order", order).Msg("GetHistoryOrderList")
	}
}

func (ts *ClientTestSuite) TestGetHistoryOrderFillList() {
	should := require.New(ts.T())

	now := time.Now()
	req := &pb.TrdGetHistoryOrderFillListRequest{
		Header: ts.usAccountHeader,
		FilterConditions: &pb.TrdFilterConditions{
			BeginTime: futu.DateTimePtr(now.AddDate(0, 0, -7)),
			EndTime:   futu.DateTimePtr(now),
		},
	}

	_, err := req.Dispatch(context.TODO(), ts.sdk)
	should.Error(err) // 模拟交易不支持成交数据
}

func (ts *ClientTestSuite) TestGetMarginRatio() {
	should := require.New(ts.T())

	// Get Margin Ratio requires a real account

	acc := futu.FindAccount(ts.accList, pb.TrdMarket_US, pb.TrdAccType_Margin, pb.SimAccType_Unknown)
	if acc == nil {
		ts.T().Skip("no suitable acc found")
	}

	req := &pb.TrdGetMarginRatioRequest{
		Header: &pb.TrdHeader{
			TrdEnv:    pb.TrdEnv_Real.Enum(),
			AccID:     proto.Uint64(acc.GetAccID()),
			TrdMarket: pb.TrdMarket_US.Enum(),
		},
		SecurityList: futu.NewSecurities("US.AAPL"),
	}

	resp, err := req.Dispatch(context.TODO(), ts.sdk)
	should.NoError(err)
	log.Info().Interface("margin ratio", resp).Msg("GetMarginRatio")
}

func (ts *ClientTestSuite) TestGetOrderFee() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderFeeRequest{
		Header:        ts.usAccountHeader,
		OrderIdExList: []string{"1234"},
	}
	_, err := req.Dispatch(context.TODO(), ts.sdk)
	should.Error(err) // 模拟交易不支持查询交易费用
}

/*
func (ts *ClientTestSuite) TestTrdFlowSummary() {
	should := require.New(ts.T())

	_, err := ts.sdk.TrdFlowSummary(ts.usAccountHeader, time.Now().Format("2006-01-02"))
	should.EqualError(err, "模拟账户不支持查询现金流水")
}

func (ts *ClientTestSuite) TestGetSubInfo() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetSubInfo()
	should.NoError(err)
	log.Info().Interface("result", res).Msg("GetSubInfo")
}

func (ts *ClientTestSuite) TestGetBasicQot() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetBasicQot([]string{"HK.00700", "HK.09988"})
	should.NoError(err)
	for _, qot := range res {
		log.Info().Interface("qot", qot).Msg("GetBasicQot")
	}
}

func (ts *ClientTestSuite) TestGetKL() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetKL(
		"HK.09988",
		adapt.KLType_Day,
		adapt.With("rehabType", adapt.RehabType_Forward),
		adapt.With("reqNum", 3),
	)
	should.NoError(err)
	for _, kl := range res.GetKlList() {
		log.Info().Interface("kline", kl).Msg("GetKL")
	}
}

func (ts *ClientTestSuite) TestGetRT() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetRT("HK.09988")
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())
	log.Info().Str("stock", res.GetName()).Int("num", len(res.GetRtList())).Msg("GetRT")
}

func (ts *ClientTestSuite) TestGetTicker() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetTicker("HK.09988")
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())
	log.Info().Str("stock", res.GetName()).Int("num", len(res.GetTickerList())).Msg("GetTicker")
}

func (ts *ClientTestSuite) TestGetOrderBook() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetOrderBook("HK.09988")
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())
	log.Info().Str("stock", res.GetName()).
		Int("实时卖盘", len(res.GetOrderBookAskList())).
		Int("实时买盘", len(res.GetOrderBookBidList())).
		Msg("GetOrderBook")
	for _, ask := range res.GetOrderBookAskList() {
		log.Info().Interface("data", ask).Msg("实时卖盘")
	}
	for _, bid := range res.GetOrderBookBidList() {
		log.Info().Interface("data", bid).Msg("实时买盘")
	}
}

func (ts *ClientTestSuite) TestGetBroker() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetBroker("HK.09988")
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())

	log.Info().Str("stock", res.GetName()).
		Int("实时经纪卖盘", len(res.GetBrokerAskList())).
		Int("实时经纪买盘", len(res.GetBrokerBidList())).
		Msg("GetBroker")
	for _, ask := range res.GetBrokerAskList() {
		log.Info().Interface("data", ask).Msg("实时经纪卖盘")
	}
	for _, bid := range res.GetBrokerBidList() {
		log.Info().Interface("data", bid).Msg("实时经纪买盘")
	}
}

func (ts *ClientTestSuite) TestRequestHistoryKL() {
	should := require.New(ts.T())

	res, err := ts.sdk.RequestHistoryKL(
		"HK.09988",
		adapt.KLType_Day,
		"2024-10-01",
		"2024-10-15",
		adapt.With("rehabType", adapt.RehabType_Forward),
		adapt.With("maxAckKLNum", 3), // 每次只取3条，模拟分页
	)
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())

	for _, kl := range res.GetKlList() {
		log.Info().Str("date", kl.GetTime()).Str("stock", res.GetName()).Float64("close", kl.GetClosePrice()).Msg("RequestHistoryKL")
	}

	next := res.GetNextReqKey()
	for len(next) > 0 {
		res, err = ts.sdk.RequestHistoryKL(
			"HK.09988",
			adapt.KLType_Day,
			"2024-10-01",
			"2024-10-15",
			adapt.With("rehabType", adapt.RehabType_Forward),
			adapt.With("maxAckKLNum", 3),
			adapt.With("nextReqKey", next),
		)
		should.NoError(err)

		for _, kl := range res.GetKlList() {
			log.Info().Str("date", kl.GetTime()).Str("stock", res.GetName()).Float64("close", kl.GetClosePrice()).Msg("RequestHistoryKL")
		}

		next = res.GetNextReqKey()
	}
}

func (ts *ClientTestSuite) TestRequestHistoryKLQuota() {
	should := require.New(ts.T())

	res, err := ts.sdk.RequestHistoryKLQuota(
		adapt.With("bGetDetail", true), // 可选：返回详细拉取过的历史纪录
	)
	should.NoError(err)
	log.Info().Interface("result", res).Msg("RequestHistoryKLQuota")
}

func (ts *ClientTestSuite) TestRequestRehab() {
	should := require.New(ts.T())

	res, err := ts.sdk.RequestRehab("HK.09988")
	should.NoError(err)

	for _, rehab := range res.GetRehabList() {
		log.Info().Interface("rehab", rehab).Msg("RequestRehab")
	}
}

func (ts *ClientTestSuite) TestGetStaticInfo() {
	should := require.New(ts.T())

	// use securities to filter
	res, err := ts.sdk.GetStaticInfo(adapt.WithSecurities([]string{"HK.09988", "HK.00700"}))
	should.NoError(err)

	for _, info := range res {
		log.Info().Interface("info", info).Msg("GetStaticInfo by securities")
	}

	// use market and secType to filter
	res2, err := ts.sdk.GetStaticInfo(
		adapt.With("market", adapt.QotMarket_HK),
		adapt.With("secType", adapt.SecurityType_Eqty),
	)
	should.NoError(err)
	log.Info().Int("num", len(res2)).Msg("GetStaticInfo by market")
}

func (ts *ClientTestSuite) TestGetSecuritySnapshot() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetSecuritySnapshot([]string{"HK.09988", "HK.00700"})
	should.NoError(err)

	for _, snap := range res {
		log.Info().Interface("snapshot", snap).Msg("GetSecuritySnapshot")
	}
}

func (ts *ClientTestSuite) TestGetPlateSet() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetPlateSet(adapt.QotMarket_HK, adapt.PlateSetType_Industry)
	should.NoError(err)

	for _, plate := range res {
		log.Info().Str("name", plate.GetName()).
			Int32("type", plate.GetPlateType()).
			Interface("plate", plate.GetPlate()).
			Msg("GetPlateSet")
	}
}

func (ts *ClientTestSuite) TestGetPlateSecurity() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetPlateSecurity(
		"HK.LIST1059",
		adapt.With("sortField", adapt.SortField_Turnover),
		adapt.With("ascend", false),
	)
	should.NoError(err)

	for _, info := range res {
		log.Info().Interface("info", info).Msg("GetPlateSecurity")
	}
}

func (ts *ClientTestSuite) TestGetReference() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetReference("HK.09988", adapt.ReferenceType_Warrant)
	should.NoError(err)
	log.Info().Int("num", len(res)).Msg("GetReference")
}

func (ts *ClientTestSuite) TestGetOwnerPlate() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetOwnerPlate([]string{"HK.09988"})
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetOwnerPlate")
}

func (ts *ClientTestSuite) TestGetOptionChain() {
	should := require.New(ts.T())

	beginTime := time.Now().AddDate(0, 0, -1).Format(futu.DateFormat)
	endTime := time.Now().Format(futu.DateFormat)

	res, err := ts.sdk.GetOptionChain("HK.09988", beginTime, endTime)
	should.NoError(err)
	log.Info().Int("num", len(res)).Msg("GetOptionChain")
}

func (ts *ClientTestSuite) TestGetWarrant() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetWarrant(0, 3,
		adapt.With("owner", adapt.NewSecurity("HK.00981")),
		adapt.With("status", adapt.WarrantStatus_Normal),
	)
	should.NoError(err)
	log.Info().Int32("count", res.GetAllCount()).Msg("GetWarrant")
	for _, warrant := range res.GetWarrantDataList() {
		log.Info().Interface("warrant", warrant).Msg("GetWarrant")
	}
}

func (ts *ClientTestSuite) TestGetCapitalFlow() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetCapitalFlow(
		"HK.09988",
		adapt.With("beginTime", time.Now().AddDate(0, 0, -1).Format(futu.TimeFormat)),
		adapt.With("endTime", time.Now().Format(futu.TimeFormat)),
		adapt.With("periodType", adapt.PeriodType_DAY),
	)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetCapitalFlow")
}

func (ts *ClientTestSuite) TestGetCapitalDistribution() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetCapitalDistribution("HK.09988")
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetCapitalDistribution")
}

func (ts *ClientTestSuite) TestGetUserSecurity() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetUserSecurity("特别关注")
	should.NoError(err)
	log.Info().Int("count", len(res)).Msg("GetUserSecurity")
}

func (ts *ClientTestSuite) TestModifyUserSecurity() {
	should := require.New(ts.T())

	err := ts.sdk.ModifyUserSecurity(
		"特别关注",
		[]string{"HK.09988"},
		adapt.ModifyUserSecurityOp_Add,
	)
	should.Error(err) // 仅支持修改自定义分组，不支持修改系统分组
}

func (ts *ClientTestSuite) TestStockFilter() {
	should := require.New(ts.T())

	// f := &qotstockfilter.BaseFilter{
	// 	FieldName:  proto.Int32(int32(qotstockfilter.StockField_StockField_MarketVal)),
	// 	FilterMin:  proto.Float64(10000000000),
	// 	SortDir:    proto.Int32(int32(qotstockfilter.SortDir_SortDir_Ascend)),
	// 	IsNoFilter: proto.Bool(false),
	// }
	f := adapt.NewBaseFilter(
		qotstockfilter.StockField_StockField_MarketVal,
		10000000000,
		0,
		qotstockfilter.SortDir_SortDir_Ascend,
	)

	res, err := ts.sdk.StockFilter(
		adapt.QotMarket_HK,
		adapt.With("begin", 0),
		adapt.With("num", 10),
		adapt.WithBaseFilters(f),
	)
	should.NoError(err)
	log.Info().Int("count", int(res.GetAllCount())).Msg("StockFilter")
	for _, stock := range res.GetDataList() {
		log.Info().Interface("stock", stock).Msg("StockFilter")
	}
}

func (ts *ClientTestSuite) TestGetIpoList() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetIpoList(adapt.QotMarket_HK)
	should.NoError(err)

	for _, ipo := range res {
		log.Info().Interface("ipo", ipo).Msg("GetIpoList")
	}
}

func (ts *ClientTestSuite) TestGetFutureInfo() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetFutureInfo([]string{"HK.TCHmain"})
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetFutureInfo")
}

func (ts *ClientTestSuite) TestRequestTradeDate() {
	should := require.New(ts.T())

	res, err := ts.sdk.RequestTradeDate(adapt.QotMarket_HK, "", "2024-12-01", "2024-12-10")
	should.NoError(err)
	log.Info().Interface("data", res).Msg("RequestTradeDate")
}

func (ts *ClientTestSuite) TestSetPriceReminder() {
	should := require.New(ts.T())

	res, err := ts.sdk.SetPriceReminder(
		"HK.09988",
		adapt.SetPriceReminderOp_Add,
		adapt.With("type", adapt.PriceReminderType_PriceDown),
		adapt.With("freq", adapt.PriceReminderFreq_OnlyOnce),
		adapt.With("value", 80),
		adapt.With("note", "go sdk"),
	)
	should.NoError(err)
	log.Info().Int64("result", res).Msg("SetPriceReminder")
}

func (ts *ClientTestSuite) TestGetPriceReminder() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetPriceReminder("", adapt.QotMarket_HK)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetPriceReminder")

	// remove all the reminders
	for _, reminder := range res {
		_, err := ts.sdk.SetPriceReminder(
			adapt.SecurityToCode(reminder.GetSecurity()),
			adapt.SetPriceReminderOp_DelAll,
		)
		should.NoError(err)
	}
}

func (ts *ClientTestSuite) TestGetUserSecurityGroup() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetUserSecurityGroup(adapt.GroupType_System)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetUserSecurityGroup")
}

func (ts *ClientTestSuite) TestGetMarketState() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetMarketState([]string{"HK.09988"})
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetMarketState")
}

func (ts *ClientTestSuite) TestGetOptionExpirationDate() {
	should := require.New(ts.T())

	res, err := ts.sdk.GetOptionExpirationDate("HK.09988")
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetOptionExpirationDate")
}
*/

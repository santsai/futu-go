package futu_test

import (
	"context"
	"errors"
	"flag"
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

var pemFile = flag.String("privateKey", "./OpenD/data/opend-dev-key.pem", "private key file")

type ClientTestSuite struct {
	suite.Suite
	client          *futu.Client
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

	pem, err := os.ReadFile(*pemFile)
	if err != nil {
		log.Error().Err(err).Msg("error reading private key")
		ts.T().SkipNow()
	}

	ts.client, err = futu.NewClient(
		futu.WithPrivateKey(pem),
		futu.WithTimeout(15*time.Second),
	)
	if err != nil {
		log.Error().Err(err).Msg("new client error")
		ts.T().SkipNow()
	}
	ts.client.RegisterHandler(pb.ProtoId_Notify, func(s2c proto.Message) error {
		msg := s2c.(*pb.NotifyResponse)
		log.Info().Interface("s2c", msg).Msg("notify handler")
		return nil
	})

	//
	accListReq := &pb.TrdGetAccListRequest{
		NeedGeneralSecAccount: proto.Bool(true),
	}
	if resp, err := accListReq.Dispatch(context.TODO(), ts.client); err != nil {
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
		SecurityList: futu.NewSecurityList(
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

	if _, err := subReq.Dispatch(context.TODO(), ts.client); err != nil {
		log.Error().Err(err).Msg("QotSubRequest")
	}
}

// TearDownSuite run once at the very end of the testing suite, after all tests have been run.
func (ts *ClientTestSuite) TearDownSuite() {
	if ts.client != nil {
		ts.client.Close()
	}
}

func (ts *ClientTestSuite) TestGetGlobalState() {
	should := require.New(ts.T())

	req := &pb.GetGlobalStateRequest{}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	fmt.Println(resp)
}

func (ts *ClientTestSuite) TestLockTrade() {
	should := require.New(ts.T())

	req := &pb.TrdUnlockTradeRequest{
		Unlock:       proto.Bool(false),
		SecurityFirm: pb.SecurityFirm_FutuSecurities.Enum(),
	}
	_, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
}

func (ts *ClientTestSuite) TestSubscribeAccPush() {
	should := require.New(ts.T())

	req := &pb.TrdSubAccPushRequest{
		AccIDList: []uint64{ts.usAccountHeader.GetAccID()},
	}

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
}

func (ts *ClientTestSuite) TestGetFunds() {
	should := require.New(ts.T())

	req := &pb.TrdGetFundsRequest{
		Header: ts.usAccountHeader,
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetFunds()).Msg("GetFunds")
}

func (ts *ClientTestSuite) TestGetPositionList() {
	should := require.New(ts.T())

	req := &pb.TrdGetPositionListRequest{
		Header: ts.usAccountHeader,
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
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

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetMaxTrdQtys()).Msg("GetMaxTrdQtys")
}

func (ts *ClientTestSuite) TestGetOpenOrderList() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderListRequest{
		Header:           ts.usAccountHeader,
		FilterStatusList: []pb.OrderStatus{pb.OrderStatus_Submitted},
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
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
		Remark:    proto.String("go client"),
	}
	resp, err := orderReq.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("result", resp).Msg("PlaceOrder")

	// cancel the order
	cancelReq := &pb.TrdModifyOrderRequest{
		Header:        ts.usAccountHeader,
		OrderID:       proto.Uint64(resp.GetOrderID()),
		ModifyOrderOp: pb.ModifyOrderOp_Cancel.Enum(),
	}
	cancelResp, err := cancelReq.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("result", cancelResp).Msg("ModifyOrder")
}

func (ts *ClientTestSuite) TestGetOrderFillList() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderFillListRequest{
		Header: ts.usAccountHeader,
	}

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err, futu.ErrNotSupportedInSimEnv))
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

	resp, err := req.Dispatch(context.TODO(), ts.client)
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

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err, futu.ErrNotSupportedInSimEnv))
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
		SecurityList: futu.NewSecurityList("US.AAPL"),
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("margin ratio", resp).Msg("GetMarginRatio")
}

func (ts *ClientTestSuite) TestGetOrderFee() {
	should := require.New(ts.T())

	req := &pb.TrdGetOrderFeeRequest{
		Header:        ts.usAccountHeader,
		OrderIdExList: []string{"1234"},
	}

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err, futu.ErrNotSupportedInSimEnv))
}

func (ts *ClientTestSuite) TestTrdFlowSummary() {
	should := require.New(ts.T())

	req := &pb.TrdFlowSummaryRequest{
		Header:       ts.usAccountHeader,
		ClearingDate: futu.DatePtr(time.Now()),
	}

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err, futu.ErrNotSupportedInSimEnv))
}

func (ts *ClientTestSuite) TestGetSubInfo() {
	should := require.New(ts.T())

	req := pb.QotGetSubInfoRequest{}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("result", res).Msg("GetSubInfo")
}

func (ts *ClientTestSuite) TestGetBasicQot() {
	should := require.New(ts.T())

	req := &pb.QotGetBasicQotRequest{
		SecurityList: futu.NewSecurityList("HK.00700", "HK.09988"),
	}

	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	for _, qot := range res.GetBasicQotList() {
		log.Info().Interface("qot", qot).Msg("GetBasicQot")
	}
}

func (ts *ClientTestSuite) TestGetKL() {
	should := require.New(ts.T())

	req := &pb.QotGetKLRequest{
		Security:  futu.NewSecurity("HK.09988"),
		KlType:    pb.KLType_Day.Enum(),
		RehabType: pb.RehabType_Forward.Enum(),
		ReqNum:    proto.Int32(3),
	}

	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	for _, kl := range res.GetKlList() {
		log.Info().Interface("kline", kl).Msg("GetKL")
	}
}

func (ts *ClientTestSuite) TestGetRT() {
	should := require.New(ts.T())

	req := &pb.QotGetRTRequest{
		Security: futu.NewSecurity("HK.09988"),
	}

	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())
	log.Info().Str("stock", res.GetName()).Int("num", len(res.GetRtList())).Msg("GetRT")
}

func (ts *ClientTestSuite) TestGetTicker() {
	should := require.New(ts.T())

	req := &pb.QotGetTickerRequest{
		Security:  futu.NewSecurity("HK.09988"),
		MaxRetNum: proto.Int32(1000),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	should.Equal("09988", res.GetSecurity().GetCode())
	log.Info().Str("stock", res.GetName()).Int("num", len(res.GetTickerList())).Msg("GetTicker")
}

func (ts *ClientTestSuite) TestGetOrderBook() {
	should := require.New(ts.T())

	req := &pb.QotGetOrderBookRequest{
		Security: futu.NewSecurity("HK.09988"),
		Num:      proto.Int32(100),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
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

	req := &pb.QotGetBrokerRequest{
		Security: futu.NewSecurity("HK.09988"),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
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

	var next []byte = nil
	first := true

	for len(next) > 0 || first {
		req := &pb.QotRequestHistoryKLRequest{
			Security:    futu.NewSecurity("HK.09988"),
			KlType:      pb.KLType_Day.Enum(),
			BeginTime:   proto.String("2024-10-01"),
			EndTime:     proto.String("2024-10-15"),
			RehabType:   pb.RehabType_Forward.Enum(),
			MaxAckKLNum: proto.Int32(3),
			NextReqKey:  next,
		}

		res, err := req.Dispatch(context.TODO(), ts.client)

		should.NoError(err)
		should.Equal("09988", res.GetSecurity().GetCode())

		for _, kl := range res.GetKlList() {
			log.Info().Str("date", kl.GetTime()).Str("stock", res.GetName()).Float64("close", kl.GetClosePrice()).Msg("RequestHistoryKL")
		}

		next = res.GetNextReqKey()
		first = false
	}
}

func (ts *ClientTestSuite) TestRequestHistoryKLQuota() {
	should := require.New(ts.T())

	req := &pb.QotRequestHistoryKLQuotaRequest{
		BGetDetail: proto.Bool(true),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("result", res).Msg("RequestHistoryKLQuota")
}

func (ts *ClientTestSuite) TestRequestRehab() {
	should := require.New(ts.T())

	req := &pb.QotRequestRehabRequest{
		Security: futu.NewSecurity("HK.09988"),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	for _, rehab := range res.GetRehabList() {
		log.Info().Interface("rehab", rehab).Msg("RequestRehab")
	}
}

func (ts *ClientTestSuite) TestGetStaticInfo() {
	should := require.New(ts.T())

	// use securities to filter
	req := pb.QotGetStaticInfoRequest{
		SecurityList: futu.NewSecurityList("HK.09988", "HK.00700"),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	for _, info := range resp.GetStaticInfoList() {
		log.Info().Interface("info", info).Msg("GetStaticInfo by securities")
	}

	// use market and secType to filter
	req2 := pb.QotGetStaticInfoRequest{
		Market:  pb.QotMarket_HK_Security.Enum(),
		SecType: pb.SecurityType_Eqty.Enum(),
	}
	resp2, err := req2.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Int("num", len(resp2.GetStaticInfoList())).Msg("GetStaticInfo by market")
}

func (ts *ClientTestSuite) TestGetSecuritySnapshot() {
	should := require.New(ts.T())

	req := &pb.QotGetSecuritySnapshotRequest{
		SecurityList: futu.NewSecurityList("HK.09988", "HK.00700"),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	for _, snap := range resp.GetSnapshotList() {
		log.Info().Interface("snapshot", snap).Msg("GetSecuritySnapshot")
	}
}

func (ts *ClientTestSuite) TestGetPlateSet() {
	should := require.New(ts.T())

	req := &pb.QotGetPlateSetRequest{
		Market:       pb.QotMarket_HK_Security.Enum(),
		PlateSetType: pb.PlateSetType_Industry.Enum(),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	for _, plate := range resp.GetPlateInfoList() {
		log.Info().Str("name", plate.GetName()).
			Interface("type", plate.GetPlateType()).
			Interface("plate", plate.GetPlate()).
			Msg("GetPlateSet")
	}
}

func (ts *ClientTestSuite) TestGetPlateSecurity() {
	should := require.New(ts.T())

	req := &pb.QotGetPlateSecurityRequest{
		Plate:     futu.NewSecurity("HK.LIST1059"),
		SortField: pb.SortField_Turnover.Enum(),
		Ascend:    proto.Bool(false),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	for _, info := range resp.GetStaticInfoList() {
		log.Info().Interface("info", info).Msg("GetPlateSecurity")
	}
}

func (ts *ClientTestSuite) TestGetReference() {
	should := require.New(ts.T())

	req := &pb.QotGetReferenceRequest{
		Security:      futu.NewSecurity("HK.09988"),
		ReferenceType: pb.ReferenceType_Warrant.Enum(),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Int("num", len(resp.GetStaticInfoList())).Msg("GetReference")
}

func (ts *ClientTestSuite) TestGetOwnerPlate() {
	should := require.New(ts.T())

	req := &pb.QotGetOwnerPlateRequest{
		SecurityList: futu.NewSecurityList("HK.09988"),
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetOwnerPlateList()).Msg("GetOwnerPlate")
}

func (ts *ClientTestSuite) TestGetOptionChain() {
	should := require.New(ts.T())

	now := time.Now()
	req := &pb.QotGetOptionChainRequest{
		Owner:     futu.NewSecurity("HK.09988"),
		BeginTime: futu.DatePtr(now.AddDate(0, 0, -1)),
		EndTime:   futu.DatePtr(now),
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Int("num", len(resp.GetOptionChain())).Msg("GetOptionChain")
}

func (ts *ClientTestSuite) TestGetWarrant() {
	should := require.New(ts.T())

	req := &pb.QotGetWarrantRequest{
		Begin:     proto.Int32(0),
		Num:       proto.Int32(3),
		Owner:     futu.NewSecurity("HK.00981"),
		Status:    pb.WarrantStatus_Normal.Enum(),
		SortField: pb.SortField_Score.Enum(),
		Ascend:    proto.Bool(false),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Int32("count", res.GetAllCount()).Msg("GetWarrant")
	for _, warrant := range res.GetWarrantDataList() {
		log.Info().Interface("warrant", warrant).Msg("GetWarrant")
	}
}

func (ts *ClientTestSuite) TestGetCapitalFlow() {
	should := require.New(ts.T())

	now := time.Now()
	req := &pb.QotGetCapitalFlowRequest{
		Security:   futu.NewSecurity("HK.09988"),
		PeriodType: pb.PeriodType_DAY.Enum(),
		BeginTime:  futu.DateTimePtr(now.AddDate(0, 0, -1)),
		EndTime:    futu.DateTimePtr(now),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetCapitalFlow")
}

func (ts *ClientTestSuite) TestGetCapitalDistribution() {
	should := require.New(ts.T())

	req := pb.QotGetCapitalDistributionRequest{
		Security: futu.NewSecurity("HK.09988"),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("GetCapitalDistribution")
}

func (ts *ClientTestSuite) TestUserSecurity() {

	// get all groups
	should := require.New(ts.T())
	req := &pb.QotGetUserSecurityGroupRequest{
		GroupType: pb.GroupType_All.Enum(),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	var sysGroupName string
	groupList := resp.GetGroupList()
	for _, sg := range groupList {
		log.Info().Interface("sec group", sg).Msg("GetUserSecurityGroup")

		// retain a system group name for later testing
		if sysGroupName == "" &&
			sg.GetGroupType() == pb.GroupType_System {
			sysGroupName = sg.GetGroupName()
		}
	}

	// test not exist
	req2 := &pb.QotGetUserSecurityRequest{
		GroupName: proto.String("does_not_exists"),
	}
	_, err2 := req2.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err2, futu.ErrUnknownWatchlist))

	// test modify system group
	req3 := &pb.QotModifyUserSecurityRequest{
		GroupName:    proto.String(sysGroupName),
		Op:           pb.ModifyUserSecurityOp_Add.Enum(),
		SecurityList: futu.NewSecurityList("HK.09988"),
	}
	_, err3 := req3.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err3, futu.ErrModifyingSysSecGroup))
}

func (ts *ClientTestSuite) TestStockFilter() {
	should := require.New(ts.T())

	// min max missing testing
	baseFilter := &pb.BaseFilter{
		FieldName:  pb.StockField_MarketVal.Enum(),
		SortDir:    pb.SortDir_Ascend.Enum(),
		IsNoFilter: proto.Bool(false),
	}

	req := &pb.QotStockFilterRequest{
		Market:         pb.QotMarket_HK_Security.Enum(),
		Begin:          proto.Int32(0),
		Num:            proto.Int32(10),
		BaseFilterList: []*pb.BaseFilter{baseFilter},
	}

	_, err := req.Dispatch(context.TODO(), ts.client)
	should.True(errors.Is(err, futu.ErrFilterMinMaxRequired))

	// fill in missing value
	baseFilter.FilterMin = proto.Float64(100_000_000)
	baseFilter.FilterMax = proto.Float64(100_000_000_000)
	res, err := req.Dispatch(context.TODO(), ts.client)

	should.NoError(err)
	log.Info().Int("count", int(res.GetAllCount())).Msg("StockFilter")
	for _, stock := range res.GetDataList() {
		log.Info().Interface("stock", stock).Msg("StockFilter")
	}
}

func (ts *ClientTestSuite) TestGetIpoList() {
	should := require.New(ts.T())

	req := &pb.QotGetIpoListRequest{
		Market: pb.QotMarket_HK_Security.Enum(),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)

	for _, ipo := range resp.GetIpoList() {
		log.Info().Interface("ipo", ipo).Msg("GetIpoList")
	}
}

func (ts *ClientTestSuite) TestGetFutureInfo() {
	should := require.New(ts.T())

	req := pb.QotGetFutureInfoRequest{
		SecurityList: futu.NewSecurityList("HK.TCHmain"),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetFutureInfoList()).Msg("GetFutureInfo")
}

func (ts *ClientTestSuite) TestRequestTradeDate() {
	should := require.New(ts.T())

	req := pb.QotRequestTradeDateRequest{
		Market:    pb.TradeDateMarket_HK.Enum(),
		BeginTime: proto.String("2024-12-01"),
		EndTime:   proto.String("2024-12-10"),
	}
	res, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", res).Msg("RequestTradeDate")
}

func (ts *ClientTestSuite) TestSetPriceReminder() {
	should := require.New(ts.T())

	tag := "go-testing-30624700"
	sec := futu.NewSecurity("HK.09988")

	// set remainder
	req := pb.QotSetPriceReminderRequest{
		Security: sec,
		Op:       pb.SetPriceReminderOp_Add.Enum(),
		Type:     pb.PriceReminderType_PriceDown.Enum(),
		Freq:     pb.PriceReminderFreq_OnlyOnce.Enum(),
		Value:    proto.Float64(0.1),
		Note:     proto.String(tag),
	}

	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Int64("result", resp.GetKey()).Msg("SetPriceReminder")

	// get remainder
	req2 := pb.QotGetPriceReminderRequest{
		Security: sec,
	}

	res2, err2 := req2.Dispatch(context.TODO(), ts.client)
	should.NoError(err2)
	log.Info().Interface("data", res2.GetPriceReminderList()).Msg("GetPriceReminder")

	// remove all added reminders
	for _, reminder := range res2.GetPriceReminderList() {
		for _, item := range reminder.GetItemList() {
			if item.GetNote() == tag {
				req := &pb.QotSetPriceReminderRequest{
					Security: reminder.GetSecurity(),
					Op:       pb.SetPriceReminderOp_Del.Enum(),
					Key:      proto.Int64(item.GetKey()),
				}
				_, err := req.Dispatch(context.TODO(), ts.client)
				should.NoError(err)
			}
		}
	}
}

func (ts *ClientTestSuite) TestGetMarketState() {
	should := require.New(ts.T())

	req := &pb.QotGetMarketStateRequest{
		SecurityList: futu.NewSecurityList("HK.09988"),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetMarketInfoList()).Msg("GetMarketState")
}

func (ts *ClientTestSuite) TestGetOptionExpirationDate() {
	should := require.New(ts.T())

	req := pb.QotGetOptionExpirationDateRequest{
		Owner: futu.NewSecurity("HK.09988"),
	}
	resp, err := req.Dispatch(context.TODO(), ts.client)
	should.NoError(err)
	log.Info().Interface("data", resp.GetDateList()).Msg("GetOptionExpirationDate")
}

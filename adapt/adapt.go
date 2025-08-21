package adapt

import (
	"strings"

	"github.com/santsai/futu-go/pb"
	"google.golang.org/protobuf/proto"
)

// NewSecurity creates a new Security by a code string, e.g. "HK.00700".
func NewSecurity(code string) *pb.Security {
	arr := strings.Split(code, ".")
	if len(arr) != 2 {
		return nil
	}

	return &pb.Security{
		Market: GetMarketID(arr[0]).Enum(),
		Code:   &arr[1],
	}
}

// NewSecurities creates a slice of Securities by a slice of code strings.
func NewSecurities(codes []string) []*pb.Security {
	sa := make([]*pb.Security, 0)
	for _, v := range codes {
		s := NewSecurity(v)
		if s != nil {
			sa = append(sa, s)
		}
	}

	return sa
}

// SecurityToCode converts a Security to a code string, e.g. "HK.00700".
func SecurityToCode(s *pb.Security) string {
	return GetMarketName(s.GetMarket()) + "." + s.GetCode()
}

// NewTradeHeader creates a new TrdHeader for a real trade account.
func NewTradeHeader(id uint64, market pb.TrdMarket) *pb.TrdHeader {
	return &pb.TrdHeader{
		TrdEnv:    pb.TrdEnv_Real.Enum(),
		AccID:     proto.Uint64(id),
		TrdMarket: market.Enum(),
	}
}

// NewSimulationTradeHeader creates a new TrdHeader for a simulation trade account.
func NewSimulationTradeHeader(id uint64, market pb.TrdMarket) *pb.TrdHeader {
	return &pb.TrdHeader{
		TrdEnv:    pb.TrdEnv_Simulate.Enum(),
		AccID:     proto.Uint64(id),
		TrdMarket: market.Enum(),
	}
}

// NewBaseFilter creates a new BaseFilter for StockFilter method.
func NewBaseFilter(fieldName pb.StockField, min, max float64, sortDir pb.SortDir) *pb.BaseFilter {
	return &pb.BaseFilter{
		FieldName:  fieldName.Enum(),
		FilterMin:  proto.Float64(min),
		FilterMax:  proto.Float64(max),
		SortDir:    sortDir.Enum(),
		IsNoFilter: proto.Bool(false),
	}
}

// NewAccumulateFilter creates a new AccumulateFilter for StockFilter method.
func NewAccumulateFilter(fieldName pb.AccumulateField, min, max float64, days int32, sortDir pb.SortDir) *pb.AccumulateFilter {
	return &pb.AccumulateFilter{
		FieldName:  fieldName.Enum(),
		FilterMin:  proto.Float64(min),
		FilterMax:  proto.Float64(max),
		Days:       proto.Int32(days),
		SortDir:    sortDir.Enum(),
		IsNoFilter: proto.Bool(false),
	}
}

// NewFinancialFilter creates a new FinancialFilter for StockFilter method.
func NewFinancialFilter(fieldName pb.FinancialField, min, max float64, quarter pb.FinancialQuarter, sortDir pb.SortDir) *pb.FinancialFilter {
	return &pb.FinancialFilter{
		FieldName:  fieldName.Enum(),
		FilterMin:  proto.Float64(min),
		FilterMax:  proto.Float64(max),
		Quarter:    quarter.Enum(),
		SortDir:    sortDir.Enum(),
		IsNoFilter: proto.Bool(false),
	}
}

// NewPatternFilter creates a new PatternFilter for StockFilter method.
func NewPatternFilter(fieldName pb.PatternField, klType pb.KLType, consecutivePeriod int32) *pb.PatternFilter {
	return &pb.PatternFilter{
		FieldName:         fieldName.Enum(),
		KlType:            klType.Enum(),
		IsNoFilter:        proto.Bool(false),
		ConsecutivePeriod: proto.Int32(consecutivePeriod),
	}
}

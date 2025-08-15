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
		Market: GetMarketID(arr[0]),
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
func NewTradeHeader(id uint64, market int32) *pb.TrdHeader {
	return &pb.TrdHeader{
		TrdEnv:    proto.Int32(int32(pb.TrdEnv_Real)),
		AccID:     proto.Uint64(id),
		TrdMarket: proto.Int32(market),
	}
}

// NewSimulationTradeHeader creates a new TrdHeader for a simulation trade account.
func NewSimulationTradeHeader(id uint64, market int32) *pb.TrdHeader {
	return &pb.TrdHeader{
		TrdEnv:    proto.Int32(int32(pb.TrdEnv_TrdEnv_Simulate)),
		AccID:     proto.Uint64(id),
		TrdMarket: proto.Int32(market),
	}
}

// NewBaseFilter creates a new BaseFilter for StockFilter method.
func NewBaseFilter(fieldName pb.StockField, min float64, max float64, sortDir pb.SortDir) *pb.BaseFilter {
	f := &pb.BaseFilter{
		FieldName:  proto.Int32(int32(fieldName)),
		SortDir:    proto.Int32(int32(sortDir)),
		IsNoFilter: proto.Bool(false),
	}

	if min > 0 {
		f.FilterMin = proto.Float64(min)
	}

	if max > 0 {
		f.FilterMax = proto.Float64(max)
	}

	return f
}

// NewAccumulateFilter creates a new AccumulateFilter for StockFilter method.
func NewAccumulateFilter(fieldName pb.AccumulateField, min float64, max float64, days int32, sortDir pb.SortDir) *pb.AccumulateFilter {
	f := &pb.AccumulateFilter{
		FieldName:  proto.Int32(int32(fieldName)),
		Days:       proto.Int32(days),
		SortDir:    proto.Int32(int32(sortDir)),
		IsNoFilter: proto.Bool(false),
	}

	if min > 0 {
		f.FilterMin = proto.Float64(min)
	}

	if max > 0 {
		f.FilterMax = proto.Float64(max)
	}

	return f
}

// NewFinancialFilter creates a new FinancialFilter for StockFilter method.
func NewFinancialFilter(fieldName pb.FinancialField, min float64, max float64, quarter int32, sortDir pb.SortDir) *pb.FinancialFilter {
	return &pb.FinancialFilter{
		FieldName:  proto.Int32(int32(fieldName)),
		FilterMin:  proto.Float64(min),
		FilterMax:  proto.Float64(max),
		Quarter:    proto.Int32(quarter),
		SortDir:    proto.Int32(int32(sortDir)),
		IsNoFilter: proto.Bool(false),
	}
}

// NewPatternFilter creates a new PatternFilter for StockFilter method.
func NewPatternFilter(fieldName pb.PatternField, klType pb.KLType, consecutivePeriod int32) *pb.PatternFilter {
	return &pb.PatternFilter{
		FieldName:         proto.Int32(int32(fieldName)),
		KlType:            proto.Int32(int32(klType)),
		IsNoFilter:        proto.Bool(false),
		ConsecutivePeriod: proto.Int32(consecutivePeriod),
	}
}

// NewCustomIndicatorFilter creates a new CustomIndicatorFilter for StockFilter method.
func NewCustomIndicatorFilter(opts ...Option) *pb.CustomIndicatorFilter {
	o := NewOptions(opts...)
	o["isNoFilter"] = false

	var f pb.CustomIndicatorFilter
	_ = o.ToProto(&f)

	return &f
}

// NewFilterConditions creates a new TrdFilterConditions.
func NewFilterConditions(opts ...Option) *pb.TrdFilterConditions {
	o := NewOptions(opts...)
	var f pb.TrdFilterConditions
	_ = o.ToProto(&f)

	return &f
}

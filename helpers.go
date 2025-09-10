package futu

import (
	"github.com/santsai/futu-go/pb"
	"strings"
	"time"
)

const (
	DateFormat = "2006-01-02"
	TimeFormat = "2006-01-02 15:04:05"
)

func DateTimePtr(t time.Time) *string {
	return ProtoPtr(t.Format(TimeFormat))
}

func DatePtr(t time.Time) *string {
	return ProtoPtr(t.Format(DateFormat))
}

type ProtoPtrType interface {
	bool | uint64 | int32 | float32 | float64 | string
}

func ProtoPtr[T ProtoPtrType](v T) *T {
	pv := new(T)
	*pv = v
	return pv
}

// find TrdAcc in accList
// simAccType: Unknown means finding a Real account
func FindAccount(accList []*pb.TrdAcc, mkt pb.TrdMarket, accType pb.TrdAccType, simAccType pb.SimAccType) *pb.TrdAcc {

	for _, acc := range accList {

		if acc.GetSimAccType() != simAccType {
			continue
		}

		if acc.GetAccType() != accType {
			continue
		}

		for _, accMkt := range acc.TrdMarketAuthList {
			if accMkt == mkt {
				return acc
			}
		}
	}
	return nil
}

var marketIDs = map[string]pb.QotMarket{
	"HK": pb.QotMarket_HK_Security,
	"US": pb.QotMarket_US_Security,
	"SH": pb.QotMarket_CNSH_Security,
	"SZ": pb.QotMarket_CNSZ_Security,
	"SG": pb.QotMarket_SG_Security,
	"JP": pb.QotMarket_JP_Security,
}

// GetMarketID 根据市场名称返回市场ID
func GetMarketID(name string) pb.QotMarket {
	id, ok := marketIDs[strings.ToUpper(name)]
	if ok {
		return id
	}

	return pb.QotMarket_Unknown
}

// GetMarketName 根据市场ID返回市场名称
func GetMarketName(id pb.QotMarket) string {
	for k, v := range marketIDs {
		if v == id {
			return k
		}
	}

	return "Unknown"
}

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

// NewSecurityList creates a slice of Securities by a slice of code strings.
func NewSecurityList(codes ...string) []*pb.Security {
	sa := make([]*pb.Security, 0)
	for _, v := range codes {
		s := NewSecurity(v)
		if s != nil {
			sa = append(sa, s)
		}
	}

	return sa
}

// NewSecurityCode converts a Security to a code string, e.g. "HK.00700".
func NewSecurityCode(s *pb.Security) string {
	return GetMarketName(s.GetMarket()) + "." + s.GetCode()
}

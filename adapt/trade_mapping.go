package adapt

import (
	"strings"

	"github.com/santsai/futu-go/pb"
)

const (
	// TrdSecMarket_Unknown 未知证券市场
	TrdSecMarket_Unknown = int32(pb.TrdSecMarket_TrdSecMarket_Unknown)

	// TrdSecMarket_HK 香港证券市场
	TrdSecMarket_HK = int32(pb.TrdSecMarket_TrdSecMarket_HK)

	// TrdSecMarket_US 美国证券市场
	TrdSecMarket_US = int32(pb.TrdSecMarket_TrdSecMarket_US)

	// TrdSecMarket_SH 沪股市场
	TrdSecMarket_SH = int32(pb.TrdSecMarket_TrdSecMarket_CN_SH)

	// TrdSecMarket_SZ 深股市场
	TrdSecMarket_SZ = int32(pb.TrdSecMarket_TrdSecMarket_CN_SZ)

	// TrdSecMarket_SG 新加坡期货市场
	TrdSecMarket_SG = int32(pb.TrdSecMarket_TrdSecMarket_SG)

	// TrdSecMarket_JP 日本期货市场
	TrdSecMarket_JP = int32(pb.TrdSecMarket_TrdSecMarket_JP)
)

var trdSecMarketIDs = map[string]pb.TrdSecMarket{
	"HK": pb.TrdSecMarket_HK,
	"US": pb.TrdSecMarket_US,
	"SH": pb.TrdSecMarket_CN_SH,
	"SZ": pb.TrdSecMarket_CN_SZ,
	"SG": pb.TrdSecMarket_SG,
	"JP": pb.TrdSecMarket_JP,
}

// GetTrdMarketID 根据市场名称返回交易市场ID
func GetTrdMarketID(name string) pb.TrdSecMarket {
	id, ok := trdSecMarketIDs[strings.ToUpper(name)]
	if ok {
		return id
	}

	return 0
}

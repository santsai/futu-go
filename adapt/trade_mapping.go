package adapt

import (
	"strings"

	"github.com/santsai/futu-go/pb"
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

package adapt

import (
	"strings"

	"github.com/santsai/futu-go/pb"
)

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

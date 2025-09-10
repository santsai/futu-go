package futu_test

import (
	"testing"

	"github.com/santsai/futu-go"
	"github.com/santsai/futu-go/pb"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewSecurity(t *testing.T) {
	should := require.New(t)

	s := futu.NewSecurity("HK.00700")
	should.NotNil(s)
	should.Equal(pb.QotMarket_HK_Security, s.GetMarket())
	should.Equal("00700", s.GetCode())

	should.Nil(futu.NewSecurity("00700"))
}

func TestNewSecurityList(t *testing.T) {
	should := require.New(t)

	sa := futu.NewSecurityList("HK.00700", "US.AAPL")
	should.Len(sa, 2)
	should.Equal(pb.QotMarket_HK_Security, sa[0].GetMarket())
	should.Equal("00700", sa[0].GetCode())
	should.Equal(pb.QotMarket_US_Security, sa[1].GetMarket())
	should.Equal("AAPL", sa[1].GetCode())

	sa = futu.NewSecurityList("HK.00700", "00700")
	should.Len(sa, 1)
	should.Equal(pb.QotMarket_HK_Security, sa[0].GetMarket())
	should.Equal("00700", sa[0].GetCode())
}

func TestNewSecurityCode(t *testing.T) {
	should := require.New(t)

	s := &pb.Security{
		Market: pb.QotMarket_HK_Security.Enum(),
		Code:   proto.String("00700"),
	}
	should.Equal("HK.00700", futu.NewSecurityCode(s))
}

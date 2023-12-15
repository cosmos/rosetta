package rosetta_test

import (
	"testing"

	"github.com/cosmos/rosetta"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

type PluginTestSuite struct {
	suite.Suite

	ir  codectypes.InterfaceRegistry
	cdc *codec.ProtoCodec

	requiredInterfaces []string
}

func (s *PluginTestSuite) SetupTest() {
	s.ir = codectypes.NewInterfaceRegistry()
	s.cdc = codec.NewProtoCodec(s.ir)
	s.requiredInterfaces = []string{
		"cosmos.base.v1beta1.Msg",
		"cosmos.tx.v1beta1.Tx",
		"cosmos.crypto.PubKey",
		"cosmos.crypto.PrivKey",
		"ibc.core.client.v1.ClientState",
		"ibc.core.client.v1.Height",
		"cosmos.tx.v1beta1.MsgResponse",
		"ibc.core.client.v1.Header",
	}
}

func (s *PluginTestSuite) TestLoadPlugin() {
	s.Run("Load cosmos-hub plugin", func() {
		err := rosetta.LoadPlugin(s.ir, "cosmos-hub")
		s.Require().NoError(err)
		interfaceList := s.ir.ListAllInterfaces()

		interfaceListMap := make(map[string]bool)
		for _, interfaceTypeURL := range interfaceList {
			interfaceListMap[interfaceTypeURL] = true
		}

		for _, requiredInterfaceTypeURL := range s.requiredInterfaces {
			s.Require().True(interfaceListMap[requiredInterfaceTypeURL])
		}
	})
}

func TestPluginTestSuite(t *testing.T) {
	suite.Run(t, new(PluginTestSuite))
}

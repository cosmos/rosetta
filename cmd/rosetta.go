package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/rosetta"
)

// RosettaCommand builds the rosetta root command given
// a protocol buffers serializer/deserializer
func RosettaCommand(ir codectypes.InterfaceRegistry, cdc codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rosetta",
		Short: "spin up a rosetta server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := rosetta.FromFlags(cmd.Flags())
			if err != nil {
				return err
			}

			protoCodec, ok := cdc.(*codec.ProtoCodec)
			if !ok {
				return fmt.Errorf("exoected *codec.ProtoMarshaler, got: %T", cdc)
			}
			conf.WithCodec(ir, protoCodec)

			err = rosetta.LoadPlugin(ir, cmd.Flag("plugin").Value.String())
			if err != nil {
				fmt.Printf("[Rosetta]- Error while loading cosmos-hub plugin: %s", err.Error())
				return err
			}

			if cmd.Flag("grpc-types-server").Value.String() != "" {
				err = rosetta.ReflectInterfaces(ir, cmd.Flag("grpc-types-server").Value.String())
				if err != nil {
					fmt.Printf("[Rosetta]- Error while reflecting from grpc server: %s", err.Error())
					return err
				}
			}

			rosettaSrv, err := rosetta.ServerFromConfig(conf)
			if err != nil {
				fmt.Printf("[Rosetta]- Error while creating server: %s", err.Error())
				return err
			}
			return rosettaSrv.Start()
		},
	}
	rosetta.SetFlags(cmd.Flags())

	return cmd
}

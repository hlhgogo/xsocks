package cmd

import (
	"github.com/hlhgogo/xsocks-go/adapater/shadowsocks"
	"github.com/spf13/cobra"
)

var (
	password, method *string
	port             *int
)
var createSessCmd = &cobra.Command{
	Use:   "session",
	Short: "Create a new sessoin",
	Long:  `Create a new sessoin`,
	Run:   createSession,
}

func createSession(cmd *cobra.Command, args []string) {
	shadowsocks.StartSs(*method, *password, *port)
}

func init() {
	method = createSessCmd.Flags().String("method", "CHACHA20-IETF-POLY1305", "ss method")
	password = createSessCmd.Flags().String("password", "transocks", "ss password")
	port = createSessCmd.Flags().Int("port", 23114, "ss port")

	createCmd.AddCommand(createSessCmd)
}

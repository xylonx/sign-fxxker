package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/xylonx/sign-fxxker/internal/config"
	"github.com/xylonx/sign-fxxker/internal/service"
)

var rootCmd = &cobra.Command{
	Use:   "sign-fxxker",
	Short: "sign in the fxxking chaoxing, teachermate etc.",
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		err = config.Setup(cfgFile)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		err = service.StartAutoSignIn()
		if err != nil {
			return err
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM)

		<-sig
		return nil
	},
}

var cfgFile string

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "config.default.yaml", "specify config file path")
}

func Execute() error {
	return rootCmd.Execute()
}

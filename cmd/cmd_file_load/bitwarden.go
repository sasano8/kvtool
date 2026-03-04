package cmd_file_load

import (
	"github.com/sasano8/kvtool/pkg/common"
	"github.com/sasano8/kvtool/pkg/encoders"
	"github.com/sasano8/kvtool/pkg/filesystems"
	"github.com/spf13/cobra"
)

func init() {
	cmd := CmdLoadFromBitwarden
	cmd.Flags().StringP("access-token", "t", "", "Bitwarden access token")
	cmd.Flags().StringP("organization-id", "g", "", "Bitwarden organization ID")
	cmd.Flags().StringP("project-id", "p", "", "Filter by project ID")
	cmd.Flags().StringP("output", "o", "@", "output path (@ = stdout)")
	cmd.MarkFlagRequired("access-token")
	cmd.MarkFlagRequired("organization-id")
}

var CmdLoadFromBitwarden = &cobra.Command{
	Use:   "bitwarden [key]",
	Short: "Load secrets from Bitwarden Secrets Manager",
	Long:  "Load secrets from Bitwarden Secrets Manager via official Go SDK.\nWithout [key], returns all secrets as JSON map.\nWith [key], returns the value of the specified secret.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accessToken := must(cmd.Flags().GetString("access-token"))
		organizationID := must(cmd.Flags().GetString("organization-id"))
		projectID := must(cmd.Flags().GetString("project-id"))
		pathOut := must(cmd.Flags().GetString("output"))

		key := ""
		if len(args) > 0 {
			key = args[0]
		}

		fs, err := filesystems.NewBitwardenFs(&filesystems.BitwardenFsConfig{
			AccessToken:    accessToken,
			OrganizationID: organizationID,
			ProjectID:      projectID,
		})
		if err != nil {
			return err
		}
		defer fs.Close()

		file, err := fs.GetFile(key)
		if err != nil {
			return err
		}

		result, err := file.LoadAsJson()
		if err != nil {
			return err
		}

		out := filesystems.FsLocalFileCli{Path: pathOut, NoMkdir: false}
		w, err := out.OpenWriter()
		if err != nil {
			return err
		}
		defer w.Close()

		enc := &encoders.ObjToJsonEncoder{Newline: true}
		mars := common.NewMarshaller(enc, result)
		return mars.Dump(w)
	},
}

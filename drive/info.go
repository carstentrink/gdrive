package drive

import (
	"fmt"
	"io"

	"google.golang.org/api/drive/v3"
)

type FileInfoArgs struct {
	Out         io.Writer
	Id          string
	SizeInBytes bool
}

func (self *Drive) Info(args FileInfoArgs) error {
	f, err := self.service.Files.Get(args.Id).SupportsTeamDrives(true).Fields("id", "name", "size", "createdTime", "modifiedTime", "md5Checksum", "mimeType", "parents", "shared", "description", "webContentLink", "webViewLink").Do()
	if err != nil {
		return fmt.Errorf("Failed to get file: %s", err)
	}

	pathfinder := self.newPathfinder()
	absPath, err := pathfinder.absPath(f)
	if err != nil {
		return err
	}

	PrintFileInfo(PrintFileInfoArgs{
		Out:         args.Out,
		File:        f,
		Path:        absPath,
		SizeInBytes: args.SizeInBytes,
	})

	return nil
}

type PrintFileInfoArgs struct {
	Out         io.Writer
	File        *drive.File
	Path        string
	SizeInBytes bool
}

func PrintFileInfo(args PrintFileInfoArgs) {
	f := args.File

	items := []kv{
		{"Id", f.Id},
		{"Name", f.Name},
		{"Path", args.Path},
		{"Description", f.Description},
		{"Mime", f.MimeType},
		{"Size", formatSize(f.Size, args.SizeInBytes)},
		{"Created", formatDatetime(f.CreatedTime)},
		{"Modified", formatDatetime(f.ModifiedTime)},
		{"Md5sum", f.Md5Checksum},
		{"Shared", formatBool(f.Shared)},
		{"Parents", formatList(f.Parents)},
		{"ViewUrl", f.WebViewLink},
		{"DownloadUrl", f.WebContentLink},
//		{"copyRequiresWriterPermission", formatBool(f.CopyRequiresWriterPermission)},
	}

	for _, item := range items {
		if item.value != "" {
			fmt.Fprintf(args.Out, "%s: %s\n", item.key, item.value)
		}
	}
}

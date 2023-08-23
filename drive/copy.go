package drive

import (
	"fmt"
	"io"

	"google.golang.org/api/drive/v3"
)

type FileCopyArgs struct {
	Out         io.Writer
	Id          string
	Name        string
	Parents     []string
}

func (self *Drive) Copy(args FileCopyArgs) error {
   // Instantiate empty drive file
   dstFile := &drive.File{Description: "Created by Drive API"}
   // Set parent folders
   dstFile.Parents = args.Parents
   dstFile.Name = args.Name

	f, err := self.service.Files.Copy(args.Id, dstFile).SupportsTeamDrives(true).Fields("name", "mimeType", "size", "id").Do()
	if err != nil {
		return fmt.Errorf("Failed to get file: %s", err)
	}

//	pathfinder := self.newPathfinder()
//	absPath, err := pathfinder.absPath(f)
//	if err != nil {
//		return err
//	}

	PrintFileInfo(PrintFileInfoArgs{
		Out:         args.Out,
		File:        f,
	})

	return nil
}

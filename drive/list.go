package drive

import (
	"fmt"
	"io"
	"text/tabwriter"

	"golang.org/x/net/context"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

type ListFilesArgs struct {
	Out         io.Writer
	MaxFiles    int64
	Selection   int64
	NameWidth   int64
	Query       string
	SortOrder   string
	SkipHeader  bool
	SizeInBytes bool
	AbsPath     bool
}

func (self *Drive) List(args ListFilesArgs) (err error) {
	listArgs := listAllFilesArgs{
		query:     args.Query,
		fields:    []googleapi.Field{"nextPageToken", "files(id,name,md5Checksum,mimeType,size,modifiedTime,parents)"},
		sortOrder: args.SortOrder,
		maxFiles:  args.MaxFiles,
		selection:  args.Selection,
	}
	files, err := self.listAllFiles(listArgs)
	if err != nil {
		return fmt.Errorf("Failed to list files: %s", err)
	}

	pathfinder := self.newPathfinder()

	if args.AbsPath {
		// Replace name with absolute path
		for _, f := range files {
			f.Name, err = pathfinder.absPath(f)
			if err != nil {
				return err
			}
		}
	}

	PrintFileList(PrintFileListArgs{
		Out:         args.Out,
		Files:       files,
		NameWidth:   int(args.NameWidth),
		SkipHeader:  args.SkipHeader,
		SizeInBytes: args.SizeInBytes,
	})

	return
}

type listAllFilesArgs struct {
	query     string
	fields    []googleapi.Field
	sortOrder string
	maxFiles  int64
	selection int64
}

func (self *Drive) listAllFiles(args listAllFilesArgs) ([]*drive.File, error) {
	var files []*drive.File

	var pageSize int64
	if args.maxFiles > 0 && args.maxFiles < 1000 {
		pageSize = args.maxFiles
	} else {
		pageSize = 1000
	}

	var changeQuery string

	if args.selection == 1 {
		changeQuery = "( ( visibility = 'anyoneCanFind' or visibility = 'anyoneWithLink' or visibility = 'domainCanFind' or visibility = 'domainWithLink' or visibility = 'limited' ) ) and trashed = false and ( mimeType != 'application/vnd.google-apps.folder' )"
	} else if args.selection == 2 {
		changeQuery = "( ( visibility = 'anyoneCanFind' or visibility = 'anyoneWithLink' or visibility = 'domainCanFind' or visibility = 'domainWithLink' or visibility = 'limited' ) and trashed = false ) "
	} else if args.selection == 3 {
		changeQuery = "( ( visibility = 'anyoneCanFind' or visibility = 'anyoneWithLink' or visibility = 'domainCanFind' or visibility = 'domainWithLink' or visibility = 'limited' ) ) and starred = true and trashed = false and ( mimeType != 'application/vnd.google-apps.folder' )"
	} else if args.selection == 4 {
		changeQuery = "( ( visibility = 'anyoneCanFind' or visibility = 'anyoneWithLink' or visibility = 'domainCanFind' or visibility = 'domainWithLink' or visibility = 'limited' ) and trashed = false ) and (mimeType contains 'video/')"
	} else {
		changeQuery = args.query
	}

   controlledStop := fmt.Errorf("Controlled stop")

   err := self.service.Files.List().SupportsTeamDrives(true).IncludeItemsFromAllDrives(true).IncludeTeamDriveItems(true).Q(changeQuery).Fields(args.fields...).OrderBy(args.sortOrder).PageSize(pageSize).Pages(context.TODO(), func(fl *drive.FileList) error {

		files = append(files, fl.Files...)

		// Stop when we have all the files we need
		if args.maxFiles > 0 && len(files) >= int(args.maxFiles) {
			return controlledStop
		}

		return nil
	})

	if err != nil && err != controlledStop {
		return nil, err
	}

	if args.maxFiles > 0 {
		n := min(len(files), int(args.maxFiles))
		return files[:n], nil
	}

	return files, nil
}

type PrintFileListArgs struct {
	Out         io.Writer
	Files       []*drive.File
	NameWidth   int
	SkipHeader  bool
	SizeInBytes bool
}

func PrintFileList(args PrintFileListArgs) {
	w := new(tabwriter.Writer)
	w.Init(args.Out, 0, 0, 3, ' ', 0)

	if !args.SkipHeader {
		fmt.Fprintln(w, "Id Name Type Size CreatedTime")
	}

	for _, f := range args.Files {
		fmt.Fprintf(w, "%s %s %s %s %s\n",
			f.Id,
			truncateString(f.Name, args.NameWidth),
			filetype(f),
			formatSize(f.Size, args.SizeInBytes),
			formatDatetime(f.CreatedTime),
		)
	}

	w.Flush()
}

func filetype(f *drive.File) string {
	if isDir(f) {
		return "dir"
	} else if isBinary(f) {
		return "bin"
	}
	return "doc"
}

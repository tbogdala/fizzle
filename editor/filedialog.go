// Copyright 2017, Timothy Bogdala <tdb@animal-machine.com>
// See the LICENSE file for more details.

package editor

import (
	"io/ioutil"

	"github.com/golang-ui/nuklear/nk"
)

// FileDialog is a Nuklear dialog box to browse the file system.
type FileDialog struct {
	Title     string
	Directory string
	Filename  string
	Bounds    nk.Rect

	cachedItems []fileDialogListEntry
}

type fileDialogListEntry struct {
	Name        string
	IsDirectory bool
}

// NewFileDialog creates a new FileDialog dialog box with the initial settings provided.
func NewFileDialog(title string, directory string, filename string, customBounds *nk.Rect) *FileDialog {
	dlg := new(FileDialog)
	dlg.Title = title
	dlg.Directory = directory
	dlg.Filename = filename
	if customBounds != nil {
		dlg.Bounds = *customBounds
	} else {
		dlg.Bounds = nk.NkRect(200, 200, 300, 450)
	}
	dlg.cachedItems = make([]fileDialogListEntry, 0, 32)

	return dlg
}

// UpdateFileList caches the result of querying the filesystem directory for items.
func (dlg *FileDialog) UpdateFileList() error {
	dlg.cachedItems = dlg.cachedItems[:0]

	// query the filesystem
	fileInfos, err := ioutil.ReadDir(dlg.Directory)
	if err != nil {
		return err
	}

	// build a new slice of cached file entries
	for _, fi := range fileInfos {
		var entry fileDialogListEntry
		entry.IsDirectory = fi.IsDir()
		entry.Name = fi.Name()
		dlg.cachedItems = append(dlg.cachedItems, entry)
	}

	return nil
}

// an enumeration for the return value of Render() indicating button press results.
const (
	FileDialogNoPress = 0
	FileDialogOkay    = 1
	FileDialogCancel  = 2
)

// Render draws the file dialog box.
func (dlg *FileDialog) Render(ctx *nk.Context) int {
	result := FileDialogNoPress

	update := nk.NkBegin(ctx, dlg.Title, dlg.Bounds, nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowTitle)
	if update > 0 {
		// directory name
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplatePushStatic(ctx, 20)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "Dir:", nk.TextLeft)
			dlg.Directory, _ = editString(ctx, nk.EditField, dlg.Directory, nk.NkFilterDefault)
			if nk.NkButtonLabel(ctx, "R") > 0 {
				dlg.UpdateFileList()
			}
		}

		// filename
		nk.NkLayoutRowTemplateBegin(ctx, 30)
		nk.NkLayoutRowTemplatePushStatic(ctx, 40)
		nk.NkLayoutRowTemplatePushVariable(ctx, 80)
		nk.NkLayoutRowTemplateEnd(ctx)
		{
			nk.NkLabel(ctx, "File:", nk.TextLeft)
			dlg.Filename, _ = editString(ctx, nk.EditField, dlg.Filename, nk.NkFilterDefault)
		}

		// file list for the directory
		totalSpace := nk.NkWindowGetContentRegion(ctx)
		listboxHeight := totalSpace.H() - 110.0
		nk.NkLayoutRowDynamic(ctx, listboxHeight, 1)
		nk.NkGroupBegin(ctx, "File List", nk.WindowBorder)
		if len(dlg.cachedItems) > 0 {
			for _, cachedItem := range dlg.cachedItems {
				nk.NkLayoutRowDynamic(ctx, 20, 1)
				nk.NkLabel(ctx, cachedItem.Name, nk.TextLeft)
			}
		}
		nk.NkGroupEnd(ctx)

		// cancel button
		nk.NkLayoutRowDynamic(ctx, 30, 2)
		if nk.NkButtonLabel(ctx, "Cancel") > 0 {
			result = FileDialogCancel
		}
		// okay button
		if nk.NkButtonLabel(ctx, "Okay") > 0 {
			result = FileDialogOkay
		}
	}
	nk.NkEnd(ctx)

	return result
}

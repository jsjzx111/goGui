package main

import (
	"io/ioutil"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/flopp/go-findfont"
)

type App struct {
	output *widget.Label
}

type config struct {
	EditWidget   *widget.Entry
	PreiewWidget *widget.RichText
	CurrentFile  fyne.URI
	SaveMenuItem *fyne.MenuItem
}

var cfg config

func setFont() {
	fontPaths := findfont.List()
	for _, path := range fontPaths {
		// 微软雅黑-常规
		if strings.Contains(path, "msyh.ttc") {
			os.Setenv("FYNE_FONT", path)
			break
		}
	}
}

func main() {
	setFont()
	defer os.Unsetenv("FYNE_FONT")

	a := app.New()
	w := a.NewWindow("MarkDown")
	edit, preview := cfg.makeUI()
	w.SetContent(container.NewHSplit(edit, preview))
	cfg.makeMenu(w)
	w.Resize(fyne.Size{Width: 800, Height: 500})
	w.CenterOnScreen()
	w.ShowAndRun()
}

func (app *config) makeUI() (*widget.Entry, *widget.RichText) {
	edit := widget.NewMultiLineEntry()
	preview := widget.NewRichTextFromMarkdown("")
	app.EditWidget = edit
	app.PreiewWidget = preview
	edit.OnChanged = preview.ParseMarkdown
	return edit, preview
}

func (app *config) makeMenu(win fyne.Window) {
	openMenuItem := fyne.NewMenuItem("打开", app.openFunc(win))
	saveMenuItem := fyne.NewMenuItem("保存", app.saveFunc(win))
	app.SaveMenuItem = saveMenuItem
	app.SaveMenuItem.Disabled = true
	saveAsMenuItem := fyne.NewMenuItem("另存为", app.saveAsFunc(win))

	fileMenu := fyne.NewMenu("文件", openMenuItem, saveMenuItem, saveAsMenuItem)
	menu := fyne.NewMainMenu(fileMenu)
	win.SetMainMenu(menu)
}

var filter = storage.NewExtensionFileFilter([]string{".md", ".MD"})

func (app *config) openFunc(win fyne.Window) func() {
	return func() {
		openDialog := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			if read == nil {
				return
			}
			defer read.Close()

			data, err2 := ioutil.ReadAll(read)
			if err2 != nil {
				dialog.ShowError(err2, win)
			}
			app.EditWidget.SetText(string(data))

			app.CurrentFile = read.URI()
			win.SetTitle(win.Title() + " - " + read.URI().Name())
			app.SaveMenuItem.Disabled = false

		}, win)
		openDialog.SetFilter(filter)
		openDialog.Show()
	}
}

func (app *config) saveFunc(win fyne.Window) func() {
	return func() {
		if app.CurrentFile != nil {
			write, err := storage.Writer(app.CurrentFile)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			write.Write([]byte(app.EditWidget.Text))
			defer write.Close()
		}
	}
}

func (app *config) saveAsFunc(win fyne.Window) func() {
	return func() {
		saveDialog := dialog.NewFileSave(func(write fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, win)
				return
			}

			if write == nil {
				return
			}
			write.Write([]byte(app.EditWidget.Text))
			app.CurrentFile = write.URI()
			defer write.Close()
			win.SetTitle(win.Title() + " - " + write.URI().Name())
			app.SaveMenuItem.Disabled = false
		}, win)
		saveDialog.Show()
	}
}

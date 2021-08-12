package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cfm "github.com/0l1v3rr/cli-file-manager/pkg"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/skratchdot/open-golang/open"
)

var (
	path string
	l    = widgets.NewList()
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	defaultPath, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	flag.StringVar(&path, "path", defaultPath, "The path of the folder.")
	flag.Parse()

	initWidgets()
}

func initWidgets() {
	l.Title = "CLI File Manager"
	l.Rows = cfm.ReadFiles(path)
	l.TextStyle = ui.NewStyle(ui.ColorWhite)
	l.WrapText = false
	l.SetRect(0, 0, 35, 20)
	l.BorderStyle.Fg = ui.ColorBlue
	l.TitleStyle.Modifier = ui.ModifierBold
	l.SelectedRowStyle.Fg = ui.ColorBlue
	l.SelectedRowStyle.Modifier = ui.ModifierBold

	p := widgets.NewParagraph()
	p.Title = "Help Menu"
	p.Text = "[↑](fg:green) - Scroll Up\n[↓](fg:green) - Scroll Down\n[q](fg:green) - Quit\n[Enter](fg:green) - Open\n[m](fg:green) - Memory Usage\n[f](fg:green) - Disk Information\n[^D (2 times)](fg:green) - Remove file\n[^F](fg:green) - Create file\n[^N](fg:green) - Create folder"
	p.SetRect(35, 0, 70, 15)
	p.BorderStyle.Fg = ui.ColorBlue
	p.TitleStyle.Modifier = ui.ModifierBold

	disk := cfm.DiskUsage("/")

	p3 := widgets.NewParagraph()
	if cfm.ReadJson() == "memory" {
		p3.Title = "Memory Usage"
		p3.Text = cfm.ReadMemStats()
	} else {
		p3.Title = "Disk Information"
		p3.Text = fmt.Sprintf("[All: ](fg:green) - %.2f GB\n[Used:](fg:green) - %.2f GB\n[Free:](fg:green) - %.2f GB", float64(disk.All)/float64(1024*1024*1024), float64(disk.Used)/float64(1024*1024*1024), float64(disk.Free)/float64(1024*1024*1024))
	}
	p3.SetRect(35, 20, 70, 15)
	p3.BorderStyle.Fg = ui.ColorBlue
	p3.TitleStyle.Modifier = ui.ModifierBold

	p2 := widgets.NewParagraph()
	p2.Title = "File Information"
	p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
	p2.SetRect(0, 30, 70, 20)
	p2.BorderStyle.Fg = ui.ColorBlue
	p2.WrapText = false
	p2.TitleStyle.Modifier = ui.ModifierBold

	ui.Render(l, p, p2, p3)

	previousKey := ""
	inputField := ""
	fileCreatingInProgress := false
	dirCreatingInProgress := false
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Down>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				l.ScrollDown()
				p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
			}
		case "<Up>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				l.ScrollUp()
				p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
			}
		case "<Home>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				l.ScrollTop()
				p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
			}
		case "<End>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				l.ScrollBottom()
				p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
			}
		case "<C-d>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				if previousKey == "<C-d>" {
					selected := getFileName(l.SelectedRow)
					if selected != ".." && selected != "../" {
						filePath := ""
						if path[len(path)-1] == '/' || selected[0] == '/' {
							filePath = fmt.Sprintf("%v%v", path, selected)
						} else {
							filePath = fmt.Sprintf("%v/%v", path, selected)
						}
						err := os.Remove(filePath)
						if err == nil {
							l.Rows = cfm.ReadFiles(path)
							l.SelectedRow = 0
							p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
						}
					}
				}
			}
		case "<C-f>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				fileCreatingInProgress = true
				l.Rows = append(l.Rows, fmt.Sprintf("[?] %v", inputField))
				l.SelectedRow = len(l.Rows) - 1
			}
		case "<C-n>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				dirCreatingInProgress = true
				l.Rows = append(l.Rows, fmt.Sprintf("[$] %v", inputField))
				l.SelectedRow = len(l.Rows) - 1
			}
		case "<Escape>":
			if fileCreatingInProgress {
				fileCreatingInProgress = false
				inputField = ""
				l.SelectedRow = 0
				l.Rows = l.Rows[:len(l.Rows)-1]
			} else if dirCreatingInProgress {
				dirCreatingInProgress = false
				inputField = ""
				l.SelectedRow = 0
				l.Rows = l.Rows[:len(l.Rows)-1]
			}
		case "m":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				p3.Title = "Memory Usage"
				p3.Text = cfm.ReadMemStats()
			}
		case "f":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				p3.Title = "Disk Information"
				p3.Text = fmt.Sprintf("[All: ](fg:green) - %.2f GB\n[Used:](fg:green) - %.2f GB\n[Free:](fg:green) - %.2f GB", float64(disk.All)/float64(1024*1024*1024), float64(disk.Used)/float64(1024*1024*1024), float64(disk.Free)/float64(1024*1024*1024))
			}
		case "<Enter>":
			if !fileCreatingInProgress && !dirCreatingInProgress {
				selected := getFileName(l.SelectedRow)
				if selected[len(selected)-1] == '/' {
					if selected == "../" {
						splitted := strings.Split(path, "/")
						if len(splitted) > 0 {
							splitted = splitted[:len(splitted)-1]
						}
						path = strings.Join(splitted, "/")
					} else {
						if path[len(path)-1] == '/' || selected[0] == '/' {
							path = fmt.Sprintf("%v%v", path, selected)
						} else {
							path = fmt.Sprintf("%v/%v", path, selected)
						}
					}
					l.Rows = cfm.ReadFiles(path)

					l.SelectedRow = 0
					l.SelectedRowStyle.Fg = ui.ColorBlue
					l.SelectedRowStyle.Modifier = ui.ModifierBold
					p2.Text = cfm.GetFileInformations(fmt.Sprintf("%v/%v", path, getFileName(l.SelectedRow)))
				} else {
					var filePath string
					if path[len(path)-1] == '/' || selected[0] == '/' {
						filePath = fmt.Sprintf("%v%v", path, selected)
					} else {
						filePath = fmt.Sprintf("%v/%v", path, selected)
					}
					open.Start(filePath)
				}
			} else if fileCreatingInProgress {
				if len(inputField) >= 3 {
					err := ioutil.WriteFile(fmt.Sprintf("%v/%v", path, inputField), []byte(""), 0755)
					if err == nil {
						l.Rows = cfm.ReadFiles(path)
						l.SelectedRow = 0
						inputField = ""
						fileCreatingInProgress = false
					}
				}
			} else if dirCreatingInProgress {
				if len(inputField) >= 3 {
					err := os.Mkdir(fmt.Sprintf("%v/%v", path, inputField), 0755)
					if err == nil {
						l.Rows = cfm.ReadFiles(path)
						l.SelectedRow = 0
						inputField = ""
						dirCreatingInProgress = false
					}
				}
			}
		}

		if fileCreatingInProgress {
			if e.ID[0] != '<' {
				inputField = inputField + e.ID
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[?] %v", inputField)
			} else if e.ID == "<Backspace>" {
				le := len(inputField)
				if le > 0 {
					inputField = inputField[:le-1]
				}
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[?] %v", inputField)
			}
		} else if dirCreatingInProgress {
			if e.ID[0] != '<' {
				inputField = inputField + e.ID
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[$] %v", inputField)
			} else if e.ID == "<Backspace>" {
				le := len(inputField)
				if le > 0 {
					inputField = inputField[:le-1]
				}
				l.Rows[len(l.Rows)-1] = fmt.Sprintf("[$] %v", inputField)
			}
		}

		if !fileCreatingInProgress && !dirCreatingInProgress {
			if previousKey == "<C-d>" {
				previousKey = ""
			} else {
				previousKey = e.ID
			}
		}

		ui.Render(l, p, p2, p3)
	}
}

func getFileName(n int) string {
	row := l.Rows[n]
	sliced := strings.Split(strings.Replace(row, "](fg:green)", "", 1), " ")
	sliced = sliced[1:]
	result := strings.Join(sliced, " ")

	return result
}

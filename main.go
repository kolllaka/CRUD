package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	Id          uint
	Title       string
	Description string
}

var tasks []Task
var createContent *fyne.Container
var tasksContent *fyne.Container
var detailTaskContent *fyne.Container
var tasksList *widget.List

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Кошдачи")
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(400, 400))
	icon, _ := fyne.LoadResourceFromPath("./icons/bongo.ico")
	w.SetIcon(icon)

	// DB connection
	DB, err := gorm.Open(sqlite.Open("./db/todo.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	DB.AutoMigrate(&Task{})
	DB.Find(&tasks)

	noTasksLabel := canvas.NewText("нету задач", color.RGBA{255, 0, 0, 255})
	if len(tasks) != 0 {
		noTasksLabel.Hide()
	}

	// deckstop app
	tasksBar := container.NewHBox(
		canvas.NewText("Добавить задачу", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", newIcon("./icons/light/plus.png"), func() {
			w.SetContent(createContent)
		}),
	)

	tasksList = widget.NewList(
		func() int {
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return canvas.NewText("Default", color.Black)
		},
		func(i widget.ListItemID, co fyne.CanvasObject) {
			co.(*canvas.Text).Text = tasks[i].Title
		},
	)
	tasksScroll := container.NewScroll(tasksList)
	tasksScroll.SetMinSize(fyne.NewSize(400, 300))

	tasksContent = container.NewVBox(
		tasksBar,
		canvas.NewLine(color.Black),
		noTasksLabel,
		tasksScroll,
	)

	tasksList.OnSelected = func(id widget.ListItemID) {
		detailTaskBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf(
					"Детали о \"%s\"",
					tasks[id].Title,
				),
				color.Black,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", newIcon("./icons/light/back.png"), func() {
				tasksList.Unselect(id)
				w.SetContent(tasksContent)
			}),
		)

		taskTitle := canvas.NewText(tasks[id].Title, color.Black)
		taskTitle.TextStyle = fyne.TextStyle{TabWidth: 4, Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}
		taskDescription.Wrapping = fyne.TextWrapBreak

		buttonsBox := container.NewVBox(
			// EDIT
			widget.NewButtonWithIcon("", newIcon("./icons/light/edit.png"), func() {
				editTaskBar := container.NewHBox(
					canvas.NewText(
						fmt.Sprintf(
							"Изменить \"%s\"",
							tasks[id].Title,
						),
						color.Black,
					),
					layout.NewSpacer(),
					widget.NewButtonWithIcon("", newIcon("./icons/light/back.png"), func() {
						tasksList.Unselect(id)
						w.SetContent(tasksContent)
					}),
				)

				editTitle := widget.NewEntry()
				editTitle.SetText(tasks[id].Title)

				editDescription := widget.NewMultiLineEntry()
				editDescription.SetText(tasks[id].Description)

				editButton := widget.NewButtonWithIcon(
					"Сохранить",
					newIcon("./icons/light/save.png"),

					// EDIT on DB function
					func() {
						DB.Find(&Task{},
							"Id = ?",
							tasks[id].Id,
						).Updates(
							Task{
								Title:       editTitle.Text,
								Description: editDescription.Text,
							},
						)

						DB.Find(&tasks)

						taskTitle.Text = editTitle.Text
						taskTitle.Refresh()
						taskDescription.SetText(editDescription.Text)
						taskDescription.Refresh()
						w.SetContent(detailTaskContent)
					},
				)

				editContent := container.NewVBox(
					editTaskBar,
					canvas.NewLine(color.Black),
					editTitle,
					editDescription,
					editButton,
				)

				w.SetContent(editContent)
			}),

			// DELETE
			widget.NewButtonWithIcon("", newIcon("./icons/light/delete.png"), func() {
				dialog.ShowConfirm(
					"Удалить задачу?",
					fmt.Sprintf(
						"Вы уверены, что хотите удалить задачу \"%s\"?",
						tasks[id].Title,
					),
					func(b bool) {
						if b {
							DB.Delete(&Task{}, "id = ?", tasks[id].Id)
							DB.Find(&tasks)

							if len(tasks) == 0 {
								noTasksLabel.Show()
							} else {
								noTasksLabel.Hide()
							}

							w.SetContent(tasksContent)

							return
						}
					},

					w,
				)
			}),
		)

		detailTaskContent = container.NewVBox(
			detailTaskBar,
			canvas.NewLine(color.Black),
			taskTitle,
			taskDescription,
			buttonsBox,
		)

		w.SetContent(detailTaskContent)
	}

	// create tasks
	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Ваша задач...")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Описание задачи...")

	saveTaskButton := widget.NewButtonWithIcon("", newIcon("./icons/light/save.png"), func() {
		if titleEntry.Text == "" {
			return
		}

		task := Task{
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
		}

		DB.Create(&task)
		DB.Find(&tasks)

		// ? очищать содержимое
		titleEntry.Text = ""
		titleEntry.Refresh()
		descriptionEntry.Text = ""
		descriptionEntry.Refresh()

		if len(tasks) == 0 {
			noTasksLabel.Show()
		} else {
			noTasksLabel.Hide()
		}

		tasksList.UnselectAll()

		w.SetContent(tasksContent)
	})

	createBar := container.NewHBox(
		canvas.NewText("Задать задачу", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", newIcon("./icons/light/back.png"), func() {
			// ? очищать содержимое
			titleEntry.Text = ""
			titleEntry.Refresh()
			descriptionEntry.Text = ""
			descriptionEntry.Refresh()

			tasksList.UnselectAll()
			w.SetContent(tasksContent)
		}),
	)

	createContent = container.NewVBox(
		createBar,
		canvas.NewLine(color.Black),
		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	w.SetContent(tasksContent)
	w.Show()
	a.Run()
}

func newIcon(path string) fyne.Resource {
	icon, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		fmt.Println(err)
	}

	return icon
}

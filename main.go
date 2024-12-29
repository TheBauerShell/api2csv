package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type Person struct {
	ID       int    `json:"id"`
	Vorname  string `json:"vorname"`
	Nachname string `json:"nachname"`
	Email    string `json:"email"`
}

func main() {
	myApp := app.New()
	window := myApp.NewWindow("Personen API Downloader")

	var outputPath string

	formatSelect := widget.NewSelect([]string{"CSV", "Excel CSV", "JSON"}, nil)
	pathLabel := widget.NewLabel("Kein Speicherort ausgewählt")

	selectPathButton := widget.NewButton("Speicherort wählen", func() {
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			if uri == nil {
				return
			}
			outputPath = uri.URI().Path()
			pathLabel.SetText(outputPath)
		}, window)
	})

	downloadButton := widget.NewButton("Daten herunterladen", func() {
		if outputPath == "" {
			dialog.ShowError(fmt.Errorf("Bitte wählen Sie zuerst einen Speicherort"), window)
			return
		}

		resp, err := http.Get("http://localhost:8080/persons")
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		var persons []Person
		if err := json.Unmarshal(data, &persons); err != nil {
			dialog.ShowError(err, window)
			return
		}

		var saveErr error
		switch formatSelect.Selected {
		case "CSV":
			saveErr = saveAsCSV(persons, outputPath)
		case "Excel CSV":
			saveErr = saveAsExcelCSV(persons, outputPath)
		case "JSON":
			saveErr = saveAsJSON(persons, outputPath)
		default:
			dialog.ShowError(fmt.Errorf("Bitte Format auswählen"), window)
			return
		}

		if saveErr != nil {
			dialog.ShowError(saveErr, window)
			return
		}

		dialog.ShowInformation("Erfolg", "Daten wurden erfolgreich gespeichert", window)
	})

	content := container.NewVBox(
		widget.NewLabel("Wählen Sie das Ausgabeformat:"),
		formatSelect,
		selectPathButton,
		pathLabel,
		downloadButton,
	)

	window.SetContent(content)
	window.Resize(fyne.NewSize(400, 300))
	window.ShowAndRun()
}

func saveAsCSV(persons []Person, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"ID", "Vorname", "Nachname", "Email"}); err != nil {
		return err
	}

	for _, person := range persons {
		if err := writer.Write([]string{
			fmt.Sprintf("%d", person.ID),
			person.Vorname,
			person.Nachname,
			person.Email,
		}); err != nil {
			return err
		}
	}
	return nil
}

func saveAsExcelCSV(persons []Person, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	if err := writer.Write([]string{"ID", "Vorname", "Nachname", "Email"}); err != nil {
		return err
	}

	for _, person := range persons {
		if err := writer.Write([]string{
			fmt.Sprintf("%d", person.ID),
			person.Vorname,
			person.Nachname,
			person.Email,
		}); err != nil {
			return err
		}
	}
	return nil
}

func saveAsJSON(persons []Person, path string) error {
	jsonData, err := json.MarshalIndent(persons, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, jsonData, 0644)
}

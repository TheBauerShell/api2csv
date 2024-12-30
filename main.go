package main

import (
	"encoding/base64"
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

// Korrigiertes Base64-Icon (blau, 128x128 PNG)
const iconBase64 = `iVBORw0KGgoAAAANSUhEUgAAAQAAAAEACAMAAABrrFhUAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAADNQTFRFAAAA/////v7+AA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAA9dAAAAziFNewAAAA90Uk5TABAgMEBQYHCAn6+/z9/vIJ3ZrQAAAAFiS0dEBI9o2VEAAAAJcEhZcwAALiMAAC4jAXilP3YAAANwSURBVHja7d3bktowEERRhBACgfD/X7uVSu0LNmA8I/VZ+z0PVGqfGV8k2ZJ1dydJkiRJkiRJkiRJkiRJkiRJkiRJkiRJkiRJkiRJkiRJkiRJkiRJv6her9fr/QprAKANa7QNoNEWgEZbABptAWi0BaDRFoBGWwAabQFotAWg0RaARlsAGm0BaLQFoNEWgEZbABptAWi0BaDRFoBGWwAabQFotAWg0RaARlsAGm0BaLQFoNH2Hwf4WqG+O8D9CvUDAE9r1E8APKxRDwA8rlKPADyvUo8AvKxSjwC8rlKPALytUo8AvK9SjwB8rFKPAHyuUo8AfK1SjwB8r1KPAL5XqUcAP6vUI4DfVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRQFylHgHEVeoRwPsq9QjgbZV6BPC6Sj0C+FilHgF8rlKPAL5WqUcA36vUI4CfVeoRwM8q9Qggr1KPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOI69QggrlOPAOJK9QggrlSPAOJa9f8D4HGtegDwsFY9AHhaq/4PgKe16k91d7dW3UmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSJEmSdLP6A7f0qOh14QmHAAAAAElFTkSuQmCC`

type Person struct {
	ID       string `json:"id"`
	Vorname  string `json:"first_name"`
	Nachname string `json:"last_name"`
	Email    string `json:"email"`
	Telefon  string `json:"phone"`
}

func main() {
	myApp := app.NewWithID("com.thebauershell.api2csv")

	// Icon als Fyne Resource erstellen
	iconData, err := base64.StdEncoding.DecodeString(iconBase64)
	if err != nil {
		fmt.Println("Fehler beim Dekodieren des Icons:", err)
		return
	}
	iconResource := fyne.NewStaticResource("icon.png", iconData)
	myApp.SetIcon(iconResource)

	window := myApp.NewWindow("API2CSV")

	var outputPath string

	urlEntry := widget.NewEntry()
	urlEntry.SetText("http://localhost:8080/persons")
	urlEntry.SetPlaceHolder("API URL eingeben")

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

		if urlEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Bitte geben Sie eine API-URL ein"), window)
			return
		}

		resp, err := http.Get(urlEntry.Text)
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
		widget.NewLabel("API URL:"),
		urlEntry,
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

	if err := writer.Write([]string{"ID", "Vorname", "Nachname", "Email", "Telefon"}); err != nil {
		return err
	}

	for _, person := range persons {
		if err := writer.Write([]string{
			person.ID,
			person.Vorname,
			person.Nachname,
			person.Email,
			person.Telefon,
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

	if err := writer.Write([]string{"ID", "Vorname", "Nachname", "Email", "Telefon"}); err != nil {
		return err
	}

	for _, person := range persons {
		record := []string{
			person.ID,
			person.Vorname,
			person.Nachname,
			person.Email,
			person.Telefon,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}
func saveAsJSON(persons []Person, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(persons, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}

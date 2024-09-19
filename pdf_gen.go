package main

import (
	"bytes"
	"fmt"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
)

func handleUpload(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Получаем файл из формы
	file, _, err := r.FormFile("excelFile")
	if err != nil {
		http.Error(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Создаем временный файл для сохранения загруженного Excel
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Копируем загруженный файл во временный файл
	_, err = io.Copy(tempFile, file)
	if err != nil {
		http.Error(w, "Failed to save uploaded file", http.StatusInternalServerError)
		return
	}
	// Открываем Excel файл
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Получаем все строки из первого листа
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Fatal(err)
	}

	// Загружаем шрифт
	fontBytes, err := os.ReadFile("Lato-Bold.ttf") // Используйте Lato Bold
	if err != nil {
		log.Fatal(err)
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем PDF-документ
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: float64(defaultLayout.Width), H: float64(defaultLayout.Height)}})

	for i, row := range rows {
		if i == 0 { // Пропускаем заголовок
			continue
		}
		elements := strings.Split(row[1], "-")
		rowLabel := []string{"cl" + row[0], elements[0][1:], elements[1] + "-" + elements[2], elements[3]}
		fmt.Println(rowLabel)
		// Создаем этикетку
		img, err := createLabel(rowLabel, defaultLayout, font)
		if err != nil {
			log.Printf("Error creating label for row %d: %v", i, err)
			continue
		}

		// Создаем временный буфер для PNG
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			log.Printf("Error encoding PNG for row %d: %v", i, err)
			continue
		}

		// Добавляем изображение в PDF
		pdf.AddPage()
		imgHolder, err := gopdf.ImageHolderByBytes(buf.Bytes())
		if err != nil {
			log.Printf("Error creating image holder for row %d: %v", i, err)
			continue
		}
		pdf.ImageByHolder(imgHolder, 0, 0, &gopdf.Rect{W: float64(defaultLayout.Width), H: float64(defaultLayout.Height)})

		log.Printf("Label added to PDF for row %d", i)
	}

	// Создаем временный файл для PDF
	pdfFile, err := os.CreateTemp("", "labels-*.pdf")
	if err != nil {
		http.Error(w, "Failed to create temporary PDF file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(pdfFile.Name()) // Удаляем временный файл после завершения
	defer pdfFile.Close()

	// Записываем PDF во временный файл
	err = pdf.WritePdf(pdfFile.Name())
	if err != nil {
		http.Error(w, "Failed to save PDF", http.StatusInternalServerError)
		return
	}

	// Перемещаем указатель файла в начало
	_, err = pdfFile.Seek(0, 0)
	if err != nil {
		http.Error(w, "Failed to reset file pointer", http.StatusInternalServerError)
		return
	}

	// Устанавливаем заголовки для скачивания файла
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=labels.pdf")

	// Копируем содержимое PDF файла в ответ
	_, err = io.Copy(w, pdfFile)
	if err != nil {
		http.Error(w, "Failed to send PDF", http.StatusInternalServerError)
		return
	}

	log.Println("PDF sent to client successfully")
}

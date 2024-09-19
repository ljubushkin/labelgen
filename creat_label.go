package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/skip2/go-qrcode"
)

// LabelLayout содержит параметры макета этикетки
type LabelLayout struct {
	Width            int
	Height           int
	QRSize           int
	QRX              int
	QRY              int
	VerticalTextX    int
	VerticalTextY    int
	VerticalTextSize float64
	TopTextX         int
	TopTextY         int
	TopTextSize      float64
	BottomTextX      int
	BottomTextY      int
	BottomTextSize   float64
}

// Создаем макет по умолчанию (75мм x 25мм при 300 DPI)
var defaultLayout = LabelLayout{
	Width:            886, // 75мм * 300DPI / 25.4мм/дюйм
	Height:           295, // 25мм * 300DPI / 25.4мм/дюйм
	QRSize:           300, // 20мм * 300DPI / 25.4мм/дюйм
	QRX:              595,
	QRY:              0,
	VerticalTextX:    30,
	VerticalTextY:    265,
	VerticalTextSize: 10, // 3мм * 300DPI / 25.4мм/дюйм
	TopTextX:         305,
	TopTextY:         100,
	TopTextSize:      16,
	BottomTextX:      275,
	BottomTextY:      250,
	BottomTextSize:   42,
}

func rotateImage(img image.Image) image.Image {
	bounds := img.Bounds()
	rotated := image.NewRGBA(image.Rect(0, 0, bounds.Dy(), bounds.Dx()))
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			rotated.Set(y-bounds.Min.Y, bounds.Max.X-x-1, img.At(x, y))
		}
	}
	return rotated
}

func createLabel(data []string, layout LabelLayout, font *truetype.Font) (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, layout.Width, layout.Height))
	draw.Draw(img, img.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)

	// Генерируем QR-код
	qr, err := qrcode.New(data[0], qrcode.Medium)
	if err != nil {
		return nil, err
	}
	qrImage := qr.Image(layout.QRSize)
	draw.Draw(img, image.Rect(layout.QRX, layout.QRY, layout.QRX+layout.QRSize, layout.QRY+layout.QRSize), qrImage, image.Point{}, draw.Over)

	// Создаем изображение для вертикального текста
	verticalTextImg := image.NewRGBA(image.Rect(0, 0, layout.Height, layout.Width/4))
	c := freetype.NewContext()
	c.SetDPI(300)
	c.SetFont(font)
	c.SetFontSize(layout.VerticalTextSize)
	c.SetClip(verticalTextImg.Bounds())
	c.SetDst(verticalTextImg)
	c.SetSrc(image.Black)

	// Рисуем текст горизонтально и центрируем его
	verticalText := data[1]

	// Центрируем по горизонтали
	startX := 60
	// Центрируем по вертикали
	startY := 89

	pt := freetype.Pt(startX, startY)
	_, err = c.DrawString(verticalText, pt)
	if err != nil {
		return nil, err
	}

	// Поворачиваем изображение с текстом
	rotatedText := rotateImage(verticalTextImg)

	// Накладываем повернутый текст на основное изображение
	draw.Draw(img, image.Rect(0, 0, layout.Width/4, layout.Height), rotatedText, image.Point{}, draw.Over)

	// Смещение центра на 3 мм влево (примерно 35 пикселей при 300 DPI)
	centerOffset := 35

	// Добавляем верхний текст (центрированный со смещением)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetFontSize(layout.TopTextSize)
	topText := data[2]
	textWidth, _ := c.DrawString(topText, freetype.Pt(0, 0))
	widthInPixels := float64(textWidth.X) / 64
	availableWidth := float64(layout.QRX - layout.TopTextX)
	topTextX := layout.TopTextX + int((availableWidth-widthInPixels)/2) - centerOffset
	_, err = c.DrawString(topText, freetype.Pt(topTextX, layout.TopTextY))
	if err != nil {
		return nil, err
	}

	// Добавляем нижний текст (центрированный со смещением)
	c.SetFontSize(layout.BottomTextSize)
	bottomText := data[3]
	textWidth, _ = c.DrawString(bottomText, freetype.Pt(0, 0))
	widthInPixels = float64(textWidth.X) / 64
	availableWidth = float64(layout.QRX - layout.BottomTextX)
	bottomTextX := layout.BottomTextX + int((availableWidth-widthInPixels)/2) - centerOffset
	_, err = c.DrawString(bottomText, freetype.Pt(bottomTextX, layout.BottomTextY))
	if err != nil {
		return nil, err
	}

	return img, nil
}

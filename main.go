package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

func obtenerImagenLocal(ruta string) (image.Image, error) {
	archivoImagen, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivoImagen.Close()
	imagen, _, err := image.Decode(archivoImagen)
	return imagen, err
}

func colorMasCercanoSinConversion(valor uint32) uint32 {
	//log.Printf("%d,", valor)
	if valor < 31728 {
		// Negro
		return 0
	} else {
		return 65535
	}
}
func colorMasCercano(c color.Color) uint8 {
	valor := convertirAEscalaDeGrises(c)
	if valor < 128 {
		// Negro
		return 0
	} else {
		return 255
	}
}
func convertirAEscalaDeGrises(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return uint32(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

func establecerTodos(imagenResultante *image.RGBA, x int, y int, quant uint32) {
	/*

		pixels[x + 1][y    ] := pixels[x + 1][y    ] + quant_error × 7 / 16
		pixels[x - 1][y + 1] := pixels[x - 1][y + 1] + quant_error × 3 / 16
		pixels[x    ][y + 1] := pixels[x    ][y + 1] + quant_error × 5 / 16
		pixels[x + 1][y + 1] := pixels[x + 1][y + 1] + quant_error × 1 / 16
	*/
	establecer(imagenResultante, x+1, y, calcular716(quant, imagenResultante.At(x+1, y)))
	establecer(imagenResultante, x-1, y+1, calcular316(quant, imagenResultante.At(x-1, y+1)))
	establecer(imagenResultante, x, y+1, calcular516(quant, imagenResultante.At(x, y+1)))
	establecer(imagenResultante, x+1, y+1, calcular116(quant, imagenResultante.At(x+1, y+1)))
}

func establecer(imagenResultante *image.RGBA, x int, y int, nuevoValor uint32) {
	if x >= imagenResultante.Bounds().Dx() || x < 0 {
		return

	}
	if y >= imagenResultante.Bounds().Dy() || y < 0 {
		return
	}
	imagenResultante.Set(x, y, color.RGBA{
		R: uint8(nuevoValor / 257),
		G: uint8(nuevoValor / 257),
		B: uint8(nuevoValor / 257),
		A: 255,
	})
}

func calcular716(quant uint32, c color.Color) uint32 {
	// Es + no * XDDDD
	r, _, _, _ := c.RGBA()
	return r + quant*7/16
}

func calcular116(quant uint32, c color.Color) uint32 {
	r, _, _, _ := c.RGBA()
	return r + quant*1/16
}

func calcular316(quant uint32, c color.Color) uint32 {
	r, _, _, _ := c.RGBA()
	return r + quant*3/16
}

func calcular516(quant uint32, c color.Color) uint32 {
	r, _, _, _ := c.RGBA()
	return r + quant*5/16
}

func aplicarDitheringAImagen(ubicacion string) error {
	imagenOriginal, err := obtenerImagenLocal(ubicacion)
	if err != nil {
		return err
	}

	ancho, alto := imagenOriginal.Bounds().Max.X, imagenOriginal.Bounds().Max.Y
	imagenConDithering := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	for y := 0; y < alto; y++ {
		for x := 0; x < ancho; x++ {
			//_, _, _, alfaOriginal := imagenOriginal.At(x, y).RGBA()
			nivelDeGris := convertirAEscalaDeGrises(imagenOriginal.At(x, y))
			//log.Printf("Nivel original %d, a uint16 %d y con desplazamiento %d", nivelDeGris, uint16(nivelDeGris), uint16(nivelDeGris>>8))
			if x == 0 && y == 0 {
				log.Printf("En 0,0 va %d", uint8(nivelDeGris/257))
			}
			imagenConDithering.Set(x, y, color.RGBA{
				R: uint8(nivelDeGris / 257),
				G: uint8(nivelDeGris / 257),
				B: uint8(nivelDeGris / 257),
				A: 255,
			})
		}

	}
	// Y acá ya tenemos la imagen en escala de grises. Ya no necesitamos la original
	outFile, err := os.Create("grayscale.png")
	if err != nil {
		return err
	}
	defer outFile.Close()
	png.Encode(outFile, imagenConDithering)

	for y := 0; y < alto; y++ {
		for x := 0; x < ancho; x++ {
			rOriginal, _, _, _ := imagenConDithering.At(x, y).RGBA()
			if x == 0 && y == 0 {
				log.Printf("En 0,0 está %d", rOriginal)
			}

			// oldpixel := pixels[x][y]
			oldPixel := (rOriginal)
			//newpixel := find_closest_palette_color(oldpixel)
			newPixel := colorMasCercanoSinConversion((rOriginal))
			//pixels[x][y] := newpixel
			imagenConDithering.Set(x, y, color.RGBA{
				R: uint8(newPixel / 257),
				G: uint8(newPixel / 257),
				B: uint8(newPixel / 257),
				A: 255,
			})
			// quant_error := oldpixel - newpixel
			quant_error := oldPixel - newPixel
			establecerTodos(imagenConDithering, x, y, quant_error)
		}

	}

	outFile, err = os.Create("dithering.png")
	if err != nil {
		return err
	}
	defer outFile.Close()
	png.Encode(outFile, imagenConDithering)
	return nil
}

func convertirImagenABlancoYNegro(ubicacion string) error {
	imagenOriginal, err := obtenerImagenLocal(ubicacion)
	if err != nil {
		return err
	}

	ancho, alto := imagenOriginal.Bounds().Max.X, imagenOriginal.Bounds().Max.Y
	imagenConDithering := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	for y := 0; y < alto; y++ {
		for x := 0; x < ancho; x++ {
			_, _, _, alfaOriginal := imagenOriginal.At(x, y).RGBA()
			nivelDeGris := convertirAEscalaDeGrises(imagenOriginal.At(x, y))
			var tono uint16 = 0
			if nivelDeGris > 32768 {
				tono = 65535
			}
			imagenConDithering.SetRGBA64(x, y, color.RGBA64{
				R: tono,
				G: tono,
				B: tono,
				A: uint16(alfaOriginal),
			})
		}

	}
	outFile, err := os.Create("bn.png")
	if err != nil {
		return err
	}
	defer outFile.Close()
	png.Encode(outFile, imagenConDithering)
	return nil
}

func main() {
	archivo := "lagartija.jpg"
	log.Printf("%v", aplicarDitheringAImagen(archivo))
	log.Printf("%v", convertirImagenABlancoYNegro(archivo))
}

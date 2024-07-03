package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
)

const NivelAlfaTotalmenteOpacoPara8Bits = 255
const MayorNumeroRepresentadoCon16Bits = 65535
const ValorIntermedioRepresentadoCon16Bits = 31728

func obtenerImagenLocal(ruta string) (image.Image, error) {
	archivoImagen, err := os.Open(ruta)
	if err != nil {
		return nil, err
	}
	defer archivoImagen.Close()
	imagen, _, err := image.Decode(archivoImagen)
	return imagen, err
}

func colorMasCercanoSegunNivelDeGris(valor uint32) uint32 {
	if valor < ValorIntermedioRepresentadoCon16Bits {
		// Negro
		return 0
	} else {
		return MayorNumeroRepresentadoCon16Bits
	}
}
func convertirAEscalaDeGrises(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return uint32(0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b))
}

func propagarError(imagenResultante *image.RGBA, x int, y int, errorDeCuantificacion uint32) {
	/*
		pixels[x + 1][y    ] := pixels[x + 1][y    ] + quant_error × 7 / 16
		pixels[x - 1][y + 1] := pixels[x - 1][y + 1] + quant_error × 3 / 16
		pixels[x    ][y + 1] := pixels[x    ][y + 1] + quant_error × 5 / 16
		pixels[x + 1][y + 1] := pixels[x + 1][y + 1] + quant_error × 1 / 16
	*/
	establecerNivelDeColorEnImagen(imagenResultante, x+1, y, calcular716(errorDeCuantificacion, imagenResultante.At(x+1, y)))
	establecerNivelDeColorEnImagen(imagenResultante, x-1, y+1, calcular316(errorDeCuantificacion, imagenResultante.At(x-1, y+1)))
	establecerNivelDeColorEnImagen(imagenResultante, x, y+1, calcular516(errorDeCuantificacion, imagenResultante.At(x, y+1)))
	establecerNivelDeColorEnImagen(imagenResultante, x+1, y+1, calcular116(errorDeCuantificacion, imagenResultante.At(x+1, y+1)))
}

func establecerNivelDeColorEnImagen(imagenResultante *image.RGBA, x int, y int, nuevoValor uint32) {
	if x >= imagenResultante.Bounds().Dx() || x < 0 {
		return

	}
	if y >= imagenResultante.Bounds().Dy() || y < 0 {
		return
	}
	nuevoValorUint8 := uint32AUint8(nuevoValor)
	imagenResultante.Set(x, y, color.RGBA{
		R: nuevoValorUint8,
		G: nuevoValorUint8,
		B: nuevoValorUint8,
		A: NivelAlfaTotalmenteOpacoPara8Bits,
	})
}

func calcular716(quant uint32, c color.Color) uint32 {
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

// En Golang cada RGBA se representa con un uint32 aunque el
// máximo número ocupa 16 bits. Esta función convierte esos 16
// bits a 8 bits, de modo que, por ejemplo,
// convierte el máximo nivel de color 65535 a 255
// ya que 65535/257=255, y lo mismo para los otros valores
func uint32AUint8(valor uint32) uint8 {
	return uint8(valor / 257)
}

func floydSteinbergDithering(imagenOriginal image.Image) *image.RGBA {
	ancho, alto := imagenOriginal.Bounds().Max.X, imagenOriginal.Bounds().Max.Y
	imagenConDithering := image.NewRGBA(image.Rect(0, 0, ancho, alto))
	for y := 0; y < alto; y++ {
		for x := 0; x < ancho; x++ {
			nivelDeGris := convertirAEscalaDeGrises(imagenOriginal.At(x, y))
			nivelDeGrisUint8 := uint32AUint8(nivelDeGris)
			imagenConDithering.Set(x, y, color.RGBA{
				R: nivelDeGrisUint8,
				G: nivelDeGrisUint8,
				B: nivelDeGrisUint8,
				A: NivelAlfaTotalmenteOpacoPara8Bits,
			})
		}

	}
	for y := 0; y < alto; y++ {
		for x := 0; x < ancho; x++ {
			// En este punto la imagen ya es gris, podemos
			// acceder a cualquier nivel RGB para obtener
			// su nivel de gris, pues los 3 son iguales.
			// Yo accedo al nivel R
			nivelDeGrisOriginal, _, _, _ := imagenConDithering.At(x, y).RGBA()
			nivelBlancoONegro := colorMasCercanoSegunNivelDeGris(nivelDeGrisOriginal)
			nivelBlancoONegroUint8 := uint32AUint8(nivelBlancoONegro)
			imagenConDithering.Set(x, y, color.RGBA{
				R: nivelBlancoONegroUint8,
				G: nivelBlancoONegroUint8,
				B: nivelBlancoONegroUint8,
				A: NivelAlfaTotalmenteOpacoPara8Bits,
			})
			errorDeCuantificacion := nivelDeGrisOriginal - nivelBlancoONegro
			propagarError(imagenConDithering, x, y, errorDeCuantificacion)
		}

	}
	return imagenConDithering
}

func guardarImagen(nombre string, imagen *image.RGBA) error {
	archivoDeImagenResultante, err := os.Create(nombre)
	if err != nil {
		return err
	}
	defer archivoDeImagenResultante.Close()
	return png.Encode(archivoDeImagenResultante, imagen)
}

func demostrarDithering(origen string, destino string) error {
	imagen, err := obtenerImagenLocal(origen)
	if err != nil {
		return err
	}
	imagenConvertida := floydSteinbergDithering(imagen)
	return guardarImagen(destino, imagenConvertida)
}

func main() {
	nombreImagenEntrada := "lagartija.jpg"
	nombreImagenSalida := "convertida.png"
	log.Printf("Error al convertir: %v", demostrarDithering(nombreImagenEntrada, nombreImagenSalida))
}

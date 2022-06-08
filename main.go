package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/joway/libimagequant-go/pngquant"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/single_threaded"
)

// Be sure to close pools/instances when you're done with them.
var pool pdfium.Pool
var instance pdfium.Pdfium

func init() {
	// Init the PDFium library and return the instance to open documents.
	pool = single_threaded.Init(single_threaded.Config{})

	var err error
	instance, err = pool.GetInstance(time.Second * 30)
	if err != nil {
		log.Fatal(err)
	}
}

func renderPage(filePath string, output string) error {
	// Load the PDF file into a byte array.
	pdfBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Write the output to a file.
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	return renderPageFromBytes(pdfBytes, f)
}

func renderPageFromBytes(pdfBytes []byte, w io.Writer) error {
	start := time.Now()

	// Open the PDF using PDFium (and claim a worker)
	doc, err := instance.OpenDocument(&requests.OpenDocument{
		File: &pdfBytes,
	})
	if err != nil {
		return err
	}

	// Always close the document, this will release its resources.
	defer instance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})

	pc, err := instance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{Document: doc.Document})
	if err != nil {
		return err
	}

	var thumb *image.RGBA

	bgColor := color.RGBA{0xe5, 0xe7, 0xeb, 0xff}

	if pc.PageCount > 1 {
		dim := 1280
		span := 2
		np := pc.PageCount

		thumb = image.NewRGBA(image.Rect(0, 0, 1280, 1280))
		draw.DrawMask(thumb, thumb.Bounds(), image.NewUniform(bgColor), image.Point{}, nil, image.Point{}, draw.Src)

		if pc.PageCount > 4 {
			// use a 3x3 grid of first 9 pages
			dim = (1280 - (16 * 2)) / 3
			if np > 9 {
				np = 9
			}
			span = 3
		} else {
			// use a 2x2 grid of first 4 pages
			dim = (1280 - 16) / 2
		}

		for i := 0; i < np; i++ {
			x0 := i % span
			y0 := int(i / span)
			place := image.Rect(x0*(dim+16), y0*(dim+16), x0*(dim+16)+dim, y0*(dim+16)+dim)

			pageRender, err := instance.RenderPageInPixels(&requests.RenderPageInPixels{
				Width:  dim,
				Height: dim,
				Page: requests.Page{
					ByIndex: &requests.PageByIndex{
						Document: doc.Document,
						Index:    i,
					},
				}, // The page to render, 0-indexed.
			})
			if err != nil {
				return err
			}

			//log.Println(i, place, pageRender.Result.Image.Bounds())
			rxb := pageRender.Result.Image.Bounds()
			place = place.Add(image.Point{(dim - rxb.Dx()) / 2, (dim - rxb.Dy()) / 2})
			draw.DrawMask(thumb, place, pageRender.Result.Image, image.Point{}, nil, image.Point{}, draw.Src)
		}

	} else {
		// only one page, render it directly

		pageRender, err := instance.RenderPageInPixels(&requests.RenderPageInPixels{
			Width:  1280,
			Height: 1280,
			Page: requests.Page{
				ByIndex: &requests.PageByIndex{
					Document: doc.Document,
					Index:    0,
				},
			}, // The page to render, 0-indexed.
		})
		if err != nil {
			return err
		}
		thumb = pageRender.Result.Image
	}

	//////////////////
	elap := time.Since(start)
	log.Println("  ", pc.PageCount, "pages rendered in", elap.String())

	imgQuantized, err := pngquant.Compress(thumb, 50, 5)
	if err != nil {
		return err
	}

	pe := png.Encoder{CompressionLevel: png.BestCompression}
	err = pe.Encode(w, imgQuantized)
	if err != nil {
		return err
	}
	elap2 := time.Since(start)
	log.Println("   quantized in", (elap2 - elap).String())
	return nil
}

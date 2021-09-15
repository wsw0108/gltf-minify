package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/qmuntal/gltf"
)

func main() {
	var (
		infile string
		outdir string
		scale  float64
		expose bool
	)
	flag.StringVar(&infile, "infile", "", "input gltf file")
	flag.StringVar(&outdir, "outdir", "", "output dir")
	flag.Float64Var(&scale, "scale", 0, "scale factor for texture resources")
	flag.BoolVar(&expose, "expose", true, "expose embeded resources")
	flag.Parse()

	if infile == "" {
		log.Fatalln("infile is required")
	}
	indir := filepath.Dir(infile)
	inname := filepath.Base(infile)
	if outdir == "" {
		outdir = filepath.Join(indir, "minified")
	}

	model, err := gltf.Open(infile)
	if err != nil {
		log.Fatalln(err)
	}

	err = os.MkdirAll(outdir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	for i, b := range model.Buffers {
		if !b.IsEmbeddedResource() || expose {
			var (
				name     string
				path     string
				filename string
			)
			if b.IsEmbeddedResource() {
				name = fmt.Sprintf("buffer_%d", i)
				filename = name + ".bin"
				path = filepath.Join(outdir, filename)
				b.URI = filename
			} else {
				path = filepath.Join(outdir, b.URI)
			}
			os.MkdirAll(filepath.Dir(path), 0755)
			os.WriteFile(path, b.Data, 0644)
			b.Data = nil // FIXME: bug in RelativeFileHandler.WriteResource
		} else {
			b.Data = nil
		}
	}

	for i, img := range model.Images {
		if img.BufferView != nil {
			// TODO: extract and erase image data from buffer?
			continue
		}
		var (
			data []byte
		)
		if img.IsEmbeddedResource() {
			data, err = img.MarshalData()
			if err != nil {
				log.Fatalln(err)
			}
			// decode to raw pixel data?
		} else {
			var imgfile string
			imgfile = filepath.Join(indir, img.URI)
			data, err = os.ReadFile(imgfile)
			if err != nil {
				log.Fatalln(err)
			}
			// decode to raw pixel data?
		}
		if scale > 0 {
			// TODO: scale data depends on mimetype(from uri suffix or mimeType field)
		}
		if !img.IsEmbeddedResource() || expose {
			var (
				name     string
				path     string
				filename string
			)
			if img.IsEmbeddedResource() {
				name = fmt.Sprintf("image_%d", i)
				// FIXME: depends on mimetype
				filename = name + ".png"
				path = filepath.Join(outdir, filename)
				img.URI = filename
			} else {
				path = filepath.Join(outdir, img.URI)
			}
			os.MkdirAll(filepath.Dir(path), 0755)
			os.WriteFile(path, data, 0644)
		} else {
			pos := strings.Index(img.URI, ",")
			prefix := img.URI[:pos+1]
			encoded := base64.StdEncoding.EncodeToString(data)
			img.URI = prefix + encoded
		}
	}

	outfile := filepath.Join(outdir, inname)
	err = gltf.Save(model, outfile)
	if err != nil {
		log.Fatalln(err)
	}
}

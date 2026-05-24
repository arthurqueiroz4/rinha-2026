package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"os"
	"rinha2026/model"
	"rinha2026/quantization"
	"rinha2026/vptree"
)

func main() {
	f, err := os.Open("resources/references.json.gz")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		log.Fatal(err)
	}
	defer gz.Close()

	var refs []model.Reference
	dec := json.NewDecoder(gz)

	if _, err := dec.Token(); err != nil {
		log.Fatal(err)
	}

	for dec.More() {
		var ref model.Reference
		if err := dec.Decode(&ref); err != nil {
			log.Fatal(err)
		}
		refs = append(refs, ref)
	}

	if _, err := dec.Token(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Loaded %d references\n", len(refs))

	vp := vptree.Build(quantization.QuantizeReferences(refs))
	if err := vptree.Persist(vp); err != nil {
		log.Fatal(err)
	}
}

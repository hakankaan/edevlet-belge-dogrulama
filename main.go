package main

import (
	"belge-dogrulama/document"
	"flag"
	"log"
	"os"
)

const (
	exampleId      = "12345678910"
	exampleBarcode = "123A-FQ88-ADF5-QQQ1"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

var (
	id, barcode *string
)

func init() {
	// Init logger
	InfoLogger = log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(log.Writer(), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Init args
	id = flag.String("i", exampleId, "Citizen id of document owner")
	barcode = flag.String("b", exampleBarcode, "Barcode to verify")

	flag.Parse()

	if *id == exampleId || *barcode == exampleBarcode {
		flag.PrintDefaults()
		return
	}
}

func main() {
	InfoLogger.Println("Starting...")
	InfoLogger.Printf("Citizen id: %s\n", *id)
	InfoLogger.Printf("Barcode: %s\n", *barcode)

	d := document.New(*barcode, *id, InfoLogger)

	if err := d.GetToken(); err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	if err := d.InsertBarcode(); err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	if err := d.InsertID(); err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	if err := d.AcceptForm(); err != nil {
		ErrorLogger.Println(err)
		os.Exit(1)
	}

	if d.IsValid {
		InfoLogger.Println("Document is valid")
	} else {
		ErrorLogger.Println("Document is invalid")
	}
}

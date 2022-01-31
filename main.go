package main

import (
	"flag"
	"fmt"
	"log"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

var (
	outputDir  = flag.String("output_dir", "", `The path to the signature image if any. Default to "signature.png" in the same directory.`)
	sigImgPath = flag.String("signature_image_path", "", `The path to the signature image if any. Default to "signature.png" in the same directory.`)
)

func main() {
	flag.Parse()

	if err := validateFlags(*outputDir, *sigImgPath); err != nil {
		log.Fatalf("validateFlags(): %v", err)
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	log.Printf(`Adding title...`)
	addTitle(pdf)
	log.Printf(`Adding signature...`)
	addSignature(pdf, *sigImgPath)

	od, err := expandPath(*outputDir)
	if err != nil {
		log.Fatalf("expandPath(%q): %v", *outputDir, err)
	}
	filename := path.Join(od, fmt.Sprintf("rent_receipt.pdf"))

	log.Printf(`Creating file %q...`, filename)
	if err := pdf.OutputFileAndClose(filename); err != nil {
		log.Fatalf("pdf.OutputFileAndClose(%q): %v", filename, err)
	}
	log.Printf(`Done!`)
}

func validateFlags(outputDir string, sigImgPath string) error {
	return nil
}

func addTitle(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Courier", "", 12)

	dayStr := fmt.Sprintf("Date: %v", time.Now().Format("Jan 2, 2006"))

	tn := time.Now().Local()
	rentTimeStr := tn.AddDate(0, 1, -tn.Day()+1).Format("Jan, 2006")

	components := []string{
		"<center><b>RECEIPT</b></center>",
		dayStr + "<br>",
		"Address of Rental Unit: 723 Chesapeake Dr. Waterloo, ON, N2K 4G4",
		"Tenant(s):              Lindsay Demars, Grady Meston",
		"Payment received for:   [x] Rent [ ] Rent Deposit [ ] Other",
		"Payment Type:           [x] E-transfer [ ] Cheque [ ] Cash [ ] Other",
		fmt.Sprintf("Notes: Rent for %v", rentTimeStr),
		"Amount: $2650.00",
		"Landlord's Name: Yizheng Ding",
		"Landlord/Authorized Agent signature",
	}

	_, lineHt := pdf.GetFontSize()
	newLineStr := "<br><br>"
	htmlStr := strings.Join(components, newLineStr)
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
}

func addSignature(pdf *gofpdf.Fpdf, sigImgPath string) {
	pdf.Image(sigImgPath, 100, pdf.GetY()-6, 40, 20, false, "", 0, "")
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("user.Current(): %v", err)
		}
		return filepath.Join(u.HomeDir, path[1:]), nil
	}

	return filepath.Abs(path)
}

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
	month      = flag.String("month", "", `Overriding the month for output`)
	address    = flag.String("address", "723", `Overriding the month for output`)
)

type rentalInfo struct {
	address string
	tenants []string
	rent    string
}

var addresses = map[string]rentalInfo{
	"723": {
		address: "723 Chesapeake Dr. Waterloo, ON, N2K 4G4",
		tenants: []string{"Lindsay Demars", "Grady Meston"},
		rent:    "$2,740.00",
	},
	"2": {
		address: "2 Lesgay Crescent, North York, ON M2J 2H8",
		tenants: []string{"Rafael Antonio Valencia-Magana", "Silvia Natalia Amado Garcia"},
		rent:    "$3,790.00",
	},
}

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
	if _, ok := addresses[*address]; !ok {
		return fmt.Errorf("No %q in %v", *address, addresses)
	}
	return nil
}

func addTitle(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Courier", "", 12)

	dayStr := fmt.Sprintf("Date: %v", time.Now().Format("Jan 2, 2006"))

	tn := time.Now().Local()
	var rentTimeStr string
	if tn.Day() > 29 {
		// Pay on month end.
		rentTimeStr = tn.AddDate(0, 1, -tn.Day()+1).Format("Jan, 2006")
	} else {
		// Pay on month start.
		rentTimeStr = tn.Format("Jan, 2006")
	}

	_ = rentTimeStr

	rentalInfo := addresses[*address]

	components := []string{
		"<center><b>RECEIPT</b></center>",
		dayStr + "<br>",
		"Address of Rental Unit: " + rentalInfo.address,
		"Tenant(s):              " + strings.Join(rentalInfo.tenants, ", "),
		"Payment received for:   [x] Rent [ ] Rent Deposit [ ] Other",
		"Payment Type:           [x] E-transfer [ ] Cheque [ ] Cash [ ] Other",
		fmt.Sprintf("Notes: Rent for %v", rentTimeStr),
		// fmt.Sprintf("Notes: Rent for Apr 8 - Apr 30, 2022 ($2836) and key deposit ($300) paid via bank draft"),
		"Amount: " + rentalInfo.rent,
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

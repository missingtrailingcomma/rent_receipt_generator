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
	address = flag.String("address", "723", `Overriding the month for output`)

	paymentMethod     = flag.String("payment_method", "emt", `Payment method, options are: emt, cheque, bank_draft, cash. Default to emt.`)
	paymentPurpose    = flag.String("payment_purpose", "rent", `The purpose of the payment, options are: rent, rent_deposit, key_deposit, utility, other. Default to rent.`)
	rentDepositMonths = flag.String("rent_deposit_months", "", `Format: "<first month> and <last month>". Must provide if payment_purpose == rent_deposit.`)
	note              = flag.String("note", "", `Must provide if payment_purpose == utility or other.`)
	paymentAmount     = flag.String("payment_amount", "", `Must provide if payment_purpose != rent.`)

	monthOverride   = flag.String("month", "", `Overriding the month for output`)
	tenantsOverride = flag.String("tenants", "", `Overriding the tenants for output`)

	outputDir  = flag.String("output_dir", "~/Downloads", `The path where the pdf outputs. Default to ~/Downloads.`)
	sigImgPath = flag.String("signature_image_path", "signature.png", `The path to the signature image if any. Default to "signature.png" in the same directory.`)
)

type rentalInfo struct {
	address string
	tenants []string
	rent    string
}

var addresses = map[string]rentalInfo{
	"723": {
		address: "723 Chesapeake Dr. Waterloo, ON, N2K 4G4",
		tenants: []string{"Xiang Chen", "Yiwen Dong", "Guoshuai Shi", "Wenhan Zhu", "Zhifeng Yao"},
		rent:    "$3,200",
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
	addContent(pdf)
	log.Printf(`Adding signature...`)
	addSignature(pdf, *sigImgPath)

	od, err := expandPath(*outputDir)
	if err != nil {
		log.Fatalf("expandPath(%q): %v", *outputDir, err)
	}
	filename := path.Join(od, "receipt.pdf")

	log.Printf(`Creating file %q...`, filename)
	if err := pdf.OutputFileAndClose(filename); err != nil {
		log.Fatalf("pdf.OutputFileAndClose(%q): %v", filename, err)
	}
	log.Printf(`Done!`)
}

func validateFlags(outputDir string, sigImgPath string) error {
	if _, ok := addresses[*address]; !ok {
		return fmt.Errorf("no %q in %v", *address, addresses)
	}

	if *paymentPurpose == "rent_deposit" && *rentDepositMonths == "" {
		return fmt.Errorf("-rent_deposit_months must be provided when payment_purpose == rent_deposit")
	}

	if (*paymentPurpose == "utility" || *paymentPurpose == "other") && *note == "" {
		return fmt.Errorf("-note must be provided when payment_purpose == other or utility")
	}

	if *paymentPurpose != "rent" && *paymentAmount == "" {
		return fmt.Errorf("-payment_amount must be provided when payment_purpose != rent")
	}

	return nil
}

func addContent(pdf *gofpdf.Fpdf) {
	pdf.SetFont("Courier", "", 12)

	dayStr := fmt.Sprintf("Date: %v", time.Now().Format("Jan 2, 2006"))
	rentalInfo := addresses[*address]

	tenants := rentalInfo.tenants
	if *tenantsOverride != "" {
		tenants = strings.Split(*tenantsOverride, ",")
	}

	components := []string{
		"<center><b>RECEIPT</b></center>",
		dayStr + "<br>",
		"Address:      " + rentalInfo.address,
		"Tenant(s):    " + strings.Join(tenants, ", "),
		"Payment For:  " + paymentPurposeStr(*paymentPurpose),
		"Pyament Note: " + noteStr(*paymentPurpose),
		"Payment Type: " + paymentMethodStr(*paymentMethod),
		"Amount:       " + amountStr(*paymentPurpose, rentalInfo.rent),
		"Landlord:     Yizheng Ding",
		"Signature:",
	}

	_, lineHt := pdf.GetFontSize()
	newLineStr := "<br><br>"
	htmlStr := strings.Join(components, newLineStr)
	html := pdf.HTMLBasicNew()
	html.Write(lineHt, htmlStr)
}

func paymentPurposeStr(paymentPurpose string) string {
	forRent := " "
	forRentDeposit := " "
	forKeyDeposit := " "
	forUtility := ""
	forOther := " "
	switch paymentPurpose {
	case "rent":
		forRent = "x"
	case "rent_deposit":
		forRentDeposit = "x"
	case "key_deposit":
		forKeyDeposit = "x"
	case "utility":
		forUtility = "x"
	case "other":
		forOther = "x"
	}
	return fmt.Sprintf("[%s] Rent [%s] Rent Deposit [%s] Key Deposit [%s] Utility<br>              [%s] Other", forRent, forRentDeposit, forKeyDeposit, forUtility, forOther)
}

func noteStr(paymentPurpose string) string {
	tn := time.Now().Local()

	switch paymentPurpose {
	case "rent":
		var rentTimeStr string
		if *monthOverride != "" {
			rentTimeStr = *monthOverride + ", " + tn.Format("2006")
		} else {
			if tn.Day() >= 24 {
				// Pay on month end.
				rentTimeStr = tn.AddDate(0, 1, -tn.Day()+1).Format("Jan, 2006")
			} else {
				// Pay on month start.
				rentTimeStr = tn.Format("Jan, 2006")
			}
		}
		return "Rent for " + rentTimeStr
	case "rent_deposit":
		return "Rent for " + *rentDepositMonths
	case "key_deposit":
		return "Key deposit"
	case "utility", "other":
		return *note
	default:
		return "N/A"
	}
}

func paymentMethodStr(paymentMethod string) string {
	useEmt := " "
	useCheque := " "
	useBankDraft := " "
	useCash := " "
	switch paymentMethod {
	case "emt":
		useEmt = "x"
	case "cheque":
		useCheque = "x"
	case "bank_draft":
		useBankDraft = "x"
	case "cash":
		useCash = "x"
	}
	return fmt.Sprintf("[%s] E-transfer [%s] Cheque [%s] Bank Draft [%s] Cash [ ] Other", useEmt, useCheque, useBankDraft, useCash)
}

func amountStr(paymentPurpose, rent string) string {
	switch paymentPurpose {
	case "rent":
		return rent
	default:
		return *paymentAmount
	}
}

func addSignature(pdf *gofpdf.Fpdf, sigImgPath string) {
	pdf.Image(sigImgPath, 42, pdf.GetY()-6, 40, 20, false, "", 0, "")
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

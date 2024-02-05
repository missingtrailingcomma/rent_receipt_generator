package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/veqryn/go-email/email"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"

	wkhtml "github.com/SebastiaanKlippert/go-wkhtmltopdf"
)

func main() {
	ctx := context.Background()

	if err := generateEmailPDF(ctx); err != nil {
		log.Fatalf("generateEmailPDF(): %v", err)
	}
}

const (
	credentialFile = "credentials.json"
)

func generateEmailPDF(ctx context.Context) error {
	b, err := os.ReadFile(credentialFile)
	if err != nil {
		return fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.MailGoogleComScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	gClient := getClient(config)

	gmailSrv, err := gmail.NewService(ctx, option.WithHTTPClient(gClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	user := "me"
	// emtTenantKeyword := "LINDSAY DEMARS"
	r, err := gmailSrv.Users.Messages.List(user).Q(fmt.Sprintf("from:notify@payments.interac.ca LINDSAY DEMARS")).MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve labels: %v", err)
	}
	if len(r.Messages) == 0 {
		fmt.Printf("shen no message \n")
	}
	msgID := r.Messages[0].Id

	msg, err := gmailSrv.Users.Messages.Get(user, msgID).Format("raw").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve message %q: %v", msgID, err)
	}

	// for _, part := range msg.Payload.Parts {
	// 	decoded, _ := base64.StdEncoding.DecodeString(part.Body.Data)
	// 	decodedStr := string(decoded)
	// 	fmt.Printf("%s", decodedStr)
	// }

	bb, err := base64.URLEncoding.DecodeString(msg.Raw)
	if err != nil {
		log.Fatalf("DecodeString(): %v", err)
	}
	// fmt.Printf("raw %v", string(bb))

	htmlText := string(bb)

	reader := strings.NewReader(htmlText)
	msgg, err := email.ParseMessage(reader)
	if err != nil {
		log.Fatalf("ParseMessage(): %v", err)
	}

	pdfg, err := wkhtml.NewPDFGenerator()
	if err != nil {
		log.Fatalf("NewPDFGenerator(): %v", err)
	}
	pdfg.AddPage(wkhtml.NewPageReader(strings.NewReader(string(msgg.Parts[0].Parts[0].Parts[1].Body))))

	err = pdfg.Create()
	if err != nil {
		log.Fatal(err)
	}

	//Your Pdf Name
	log.Printf("ssss 1\n")
	err = pdfg.WriteFile("/Users/yizhengd/Downloads/out.pdf")
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

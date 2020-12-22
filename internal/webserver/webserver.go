package webserver

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/alpdeniz/themachine/internal/compute"
	"github.com/alpdeniz/themachine/internal/keystore"
	"github.com/alpdeniz/themachine/internal/network"
	"github.com/alpdeniz/themachine/internal/transaction"
	"github.com/go-chi/chi"
)

// Web template data types

type ObjectType struct {
	Name  string
	Value int
}
type Key struct {
	Name           string
	Address        string
	DerivationPath string
}

type Organization struct {
	Name            string
	Description     string
	MasterPublicKey string
	Transaction     string
	Active          bool
}

type CommonData struct {
	PageTitle     string
	Organizations []Organization
	Keys          []Key
	Transactions  []transaction.Transaction
	ObjectTypes   []ObjectType
	Result        string
}

type ShowTransactionData struct {
	PageTitle    string
	Organization transaction.Organization
	Hash         string
	ObjectType   string
	Data         string
	Targets      []string
	Date         string
	DownloadLink string // in case it is a downloadable
	Result       string // in case it is an executable
}

var tmpl *template.Template

// Start sets up a webserver and it routes to handlers
func Start(port int) {
	fmt.Println("Starting web server")

	// setup templates
	var err error
	tmpl, err = template.ParseGlob("./templates/*")
	if err != nil {
		fmt.Println("Cannot parse templates:", err)
		return
	}

	// set router
	r := chi.NewRouter()
	r.Get("/", homeHandler)
	r.Get("/{cmd}", cmdHandler)
	r.Get("/{cmd}/{txid}", txOperationHandler)

	r.Post("/create", relayHandler)

	// start http server
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	fmt.Println("Serving on port ", port)
}

// WEB HANDLER START
//--------------------------

// homeHandler serves a home page e.g. for intro or docs
func homeHandler(w http.ResponseWriter, r *http.Request) {

	var templatePath = "home.html"
	info := CommonData{
		PageTitle: "TheMachine - Home",
	}

	organizationTransactions := transaction.RetrieveByObjectType(transaction.Genesis)
	for _, v := range organizationTransactions {
		org := Organization{
			v.Organization.Name,
			v.Organization.Description,
			hex.EncodeToString(v.Organization.MasterPublicKey),
			hex.EncodeToString(v.Hash),
			true, // active
		}
		info.Organizations = append(info.Organizations, org)
	}

	tmpl.ExecuteTemplate(w, templatePath, info)
}

// Handles bare web user requests e.g. stats
func cmdHandler(w http.ResponseWriter, r *http.Request) {

	info := CommonData{}
	var templatePath string
	cmd := chi.URLParam(r, "cmd")

	switch cmd {
	case "keys":

		info.PageTitle = "The Machine - Keys"
		for _, v := range keystore.CurrentKeyMap {
			key := Key{
				v.Name,
				string(v.Address),
				v.DerivationPath,
			}
			info.Keys = append(info.Keys, key)
		}
		fmt.Println("Keys: ", info.Keys)
		templatePath = "keys.html"

	case "create":

		info.PageTitle = "The Machine - Keys"

		// load object types
		for i := 0; i < 10; i++ {
			obj := ObjectType{
				transaction.ObjectType(byte(i)).String(),
				i,
			}
			info.ObjectTypes = append(info.ObjectTypes, obj)
		}

		// load keys
		for _, v := range keystore.CurrentKeyMap {
			key := Key{
				v.Name,
				v.Address,
				v.DerivationPath,
			}
			info.Keys = append(info.Keys, key)
		}
		fmt.Println("Keys: ", info.Keys)

		// load organizations
		organizationTransactions := transaction.RetrieveByObjectType(transaction.Genesis)
		for _, v := range organizationTransactions {
			org := Organization{
				v.Organization.Name,
				v.Organization.Description,
				hex.EncodeToString(v.Organization.MasterPublicKey),
				hex.EncodeToString(v.Hash),
				true, // active
			}
			info.Organizations = append(info.Organizations, org)
		}

		templatePath = "create.html"
	}

	tmpl.ExecuteTemplate(w, templatePath, info)
}

// Handles web user requests
// - /web/show/{txhash}     : Shows transaction contents
// - /web/download/{txhash} : Downloads transaction contents
// - /web/run/{txhash}      : Executes code (depending on the permission setup) and displays the result
func txOperationHandler(w http.ResponseWriter, r *http.Request) {

	info := ShowTransactionData{}
	var templatePath string
	const RFC822 = "02 Jan 2006 15:04 MST" // Date format layout

	fmt.Println(r.URL.Path)

	cmd := chi.URLParam(r, "cmd")
	txhex := chi.URLParam(r, "txid")
	txhexBytes, err := hex.DecodeString(txhex)
	if err != nil {
		fmt.Println("Cannot decode transaction hex", err)
		return
	}

	// check if transaction exists
	tx := transaction.Retrieve(txhexBytes)
	if tx == nil {
		fmt.Fprintf(w, "There is no transaction with hash %s", hex.EncodeToString(txhexBytes))
		return
	}

	// TODO: Better routing
	switch cmd {
	case "show":

		info.PageTitle = "The Machine - Transaction Details"
		info.Hash = hex.EncodeToString(tx.Hash)
		info.Data = string(tx.Data)
		info.ObjectType = tx.ObjectType.String()
		info.Organization = tx.Organization
		info.Targets = tx.Targets
		info.Date = tx.Date.Format(RFC822)
		fmt.Println(info.Hash, info.ObjectType, info.Data, info.Targets, info.Date)
		templatePath = "show.html"

	case "download":

		// check if downloadable file
		if tx.ObjectType != transaction.File {
			fmt.Fprintf(w, "The transaction is not a FILE")
			return
		}
		http.ServeContent(w, r, fmt.Sprintf("transaction-%s.%s", txhex[:4], string(tx.SubType)), tx.Date, bytes.NewReader(tx.Data))
		return

	case "run":
		// set template
		templatePath = "run.html"
		// is executable?
		// then execute via compute.Execute()
		if tx.ObjectType == transaction.Executable {
			info.Result = string(compute.Execute(string(tx.Data)))
			// fmt.Fprintf(w, "Got compute request at %s. Result is: %s", txhex, result)

		} else {
			fmt.Fprintf(w, "Transaction is not an executable(%d). It is %d", transaction.Executable, transaction.ObjectType(tx.ObjectType))
		}
	}
	err = tmpl.ExecuteTemplate(w, templatePath, info)
	if err != nil {
		fmt.Println(err)
	}
}

// Relays a transaction via web interface
func relayHandler(w http.ResponseWriter, r *http.Request) {

	// read posted transaction
	objectTypeStr := r.PostFormValue("objectType")
	data := r.PostFormValue("data")
	targets := strings.Split(r.PostFormValue("targetPaths"), ",")
	organizationHex := r.PostFormValue("organization")

	// cast object type to int
	objectTypeInt, err := strconv.Atoi(objectTypeStr)
	if err != nil {
		fmt.Println("Error parsing object type", err, objectTypeStr, objectTypeInt)
		w.Write([]byte("Error objecttype"))
		return
	}

	// cast organization transaction hex to bytes
	organization, err := hex.DecodeString(organizationHex)
	if err != nil {
		fmt.Println("Cannot parse organization tx hex")
		return
	}

	// build transaction
	tx, err := transaction.Build(transaction.ObjectType(objectTypeInt), "json", organization, []byte(data), targets)
	if err != nil {
		fmt.Println("Error while building transaction via web", err)
		return
	}

	ok, err := tx.Validate()
	if !ok || err != nil {
		fmt.Println("Transaction is not valid", err)
		return
	}

	tx.Save()

	// fire away
	counter := network.RelayTransaction(nil, tx.ToBytes())
	if counter == 0 {
		w.Write([]byte("Could not relay transaction to any nodes"))
	}

	w.Write([]byte(tx.Hash))
}

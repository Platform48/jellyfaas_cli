package main

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/kpango/glg"
	"golang.org/x/term"
	"gopkg.in/src-d/go-git.v4"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Platform48/jellyfaas_cli/entities"
	"github.com/imroc/req/v3"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"

	"github.com/charmbracelet/glamour"
)

type Operations struct {
	Status string `json:"status"`
}

type Config struct {
	APIKey string `yaml:"apikey"`
	Env    string `yaml:"env"`
}

// Version
const version = "1.0.0"

const p48CoreService = "https://api.jellyfaas.com/core-service/v1"

const webUi = "https://app.jellyfaas.com/function/"

const p48AuthService = "https://api.jellyfaas.com/auth-service/v1"
const p48templatesRepo = "https://github.com/Platform48/jellyfaas_public_templates.git"
const hiddenDataFile = ".jellyfaas"
const jfApikeyHeader = "x-jf-apikey"
const jellyfaasEndpoint = "https://api.jellyfaas.com/"

const maxOpsLoops = 10

const specfile = "jellyspec.json"

const goTemplate = "go-template"
const dotnetTemplate = "dotnet-template"
const javaTemplate = "java-template"
const nodeTemplate = "node-template"
const phpTemplate = "php-template"
const pythonTemplate = "python-template"
const rubyTemplate = "ruby-template"

func main() {
	var opts entities.Options

	if strings.Contains(p48CoreService, "localhost") {
		color.Cyan("Running in development mode")
	}

	parser := flags.NewParser(&opts, flags.Default)

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if parser.Active.Name != "version" {
		color.Yellow("\nJellyFaaS CLI v" + version + " - http://app.jellyfaas.com \n\n")
	}

	switch parser.Active.Name {
	case "user":
		switch parser.Active.Active.Name {
		case "create":
			createUser(opts.User.Create.Email, opts.User.Create.Name)
		case "delete":
			deleteUser(opts.User.Delete.Email)
		case "list":
			listUsers()
		}
	case "secret":
		getApiKey()
	case "library":
		getLibrary(opts.Library.Details, opts.Library.ReadMe)
	case "deploy":
		deployFunction(opts.Deploy.ZipFile, opts.Deploy.Wait)
	case "token":
		getToken(opts.Token)
	case "spec":
		generateSpec(opts.Spec.Name, opts.Spec.Raw, opts.Spec.Flat)
	case "builds":
		switch parser.Active.Active.Name {
		case "list":
			getBadBuilds()
		case "clean":
			cleanBadBuilds(opts.BadBuilds.Clean.BuildId)
		}
	case "create":
		createProject(opts.Create.Name, opts.Create.Language, opts.Create.Destination, opts.Create.Always)
	case "zip":
		zipProjectAndDeploy(opts.Zip.Source, opts.Zip.Overwrite, opts.Zip.Deploy, opts.Zip.Wait)
	case "exists":
		checkIfFunctionExists(opts.Exists.Name)
	case "base64":
		base64EncodeDecode(opts.Base64.Encode, opts.Base64.Decode)
	case "version":
		showVersion()
	default:
		fmt.Println("Unknown command")
	}
}

func showVersion() {
	fmt.Printf("JellyFaaS CLI v%s\n", version)
}

func base64EncodeDecode(encode string, decode string) {
	if encode != "" {
		encoded := base64.StdEncoding.EncodeToString([]byte(encode))
		fmt.Println(encoded)
		return
	}

	if decode != "" {
		decoded, err := base64.StdEncoding.DecodeString(decode)
		if err != nil {
			fmt.Println("Error decoding string", err)
			return
		}
		fmt.Println(string(decoded))
		return
	}

	fmt.Println("Unknown command")
}

func generateSpec(json string, raw bool, flat bool) {
	outputSchema, err := entities.GenerateJsonSchemaFromJsonString(json, flat)
	if err != nil {
		fmt.Println("Error generating schema", err)
		return
	}

	if raw {
		fmt.Println(*outputSchema)
		return
	}
	fmt.Println("Json Schema (basic):\n------------------------------\n")
	fmt.Println(*outputSchema)
	fmt.Println("\n------------------------------\n")
}

func createUser(email string, username string) {

	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	var userRequest = entities.UserRequest{
		Type:  "user",
		Name:  username,
		Email: email,
	}
	var userResponse entities.UserResponse

	url := p48CoreService + "/entity"

	fmt.Println("Creating user: " + email)
	request, err := req.NewClient().NewRequest().SetBody(userRequest).SetHeader(jfApikeyHeader, configFile.APIKey).SetSuccessResult(&userResponse).Post(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 201 {
		fmt.Println("\tAn error happened when attempting to create a user!")
		return
	}

	fmt.Printf("\tUser created, password set too: %s\n\tYou cannot get this password again, please note it down.\n", userResponse.Password)
}

func getToken(opts entities.GetTokenCommand) {
	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	var tokenResponse entities.TokenResponse
	url := p48AuthService + "/validate"

	request, err := req.NewClient().NewRequest().SetSuccessResult(&tokenResponse).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to get token!")
		return
	}
	fmt.Printf("Token details:\n\n")
	fmt.Printf("Token:\n%s\n\n", tokenResponse.Token)
	fmt.Printf("Expiry: %s\n", tokenResponse.Expiry)
}

func deleteUser(email string) {
	fmt.Printf("Deleting user: %s \n", email)
}

func listUsers() {

	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	var listUsersResponse entities.Entity
	url := p48CoreService + "/entity"

	request, err := req.NewClient().NewRequest().SetSuccessResult(&listUsersResponse).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to list users!")
		return
	}

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"Name", "Email", "Created At", "Updated At"})
	for _, user := range listUsersResponse.Entities {
		t.AppendRow([]interface{}{user.Name, user.Email, user.CreatedAt, user.UpdatedAt})

	}
	t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
	t.Render()

}

func getApiKey() {

	fmt.Print("Enter (or paste from your UI Profile page) your secret key: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("\nError reading password:", err)
		return
	}

	if len(password) < 15 {
		fmt.Println("\nSecret key too short, are you sure you entered the correct key?")
		return
	}

	fmt.Println("\nSecret entered successfully, writing to .jellyfaas file.")
	_ = writeP48KeyFile(&Config{
		APIKey: string(password),
		Env:    "jellyfaas",
	})

	fmt.Printf("\tSecret Key written to file " + hiddenDataFile + "\n")
}

func getBadBuilds() {
	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	url := p48CoreService + "/badbuilds"
	var response entities.BadBuildResponse

	request, err := req.NewClient().NewRequest().SetSuccessResult(&response).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to get Secret Key!")
		return
	}

	headerBold := color.New(color.BgHiRed, color.Bold).SprintFunc()
	fmt.Printf("\n\n%s\n", headerBold("Bad Builds:"))

	if len(response.BadBuilds) > 0 {
		displayBadBuilds(response)
	}

	fmt.Println("\n\nDone!")
	return
}

func displayBadBuilds(response entities.BadBuildResponse) {
	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	yellowBold := color.New(color.FgYellow, color.Bold, color.BgBlack).SprintFunc()

	for _, b := range response.BadBuilds {
		fmt.Printf("  %s %s\n", greenBold("Build Id:"), b.BuildId)
		fmt.Printf("  %s %s\n", greenBold("Created At:"), b.CreatedAt.Format(time.RFC1123))
		fmt.Printf("  %s %s\n", greenBold("Name:"), b.Name)
		fmt.Printf("  %s %s\n", greenBold("Function ID:"), b.FunctionId)
		fmt.Printf("  %s %s\n", greenBold("Error Message:"), yellowBold(b.ErrorMessage))
		fmt.Printf("\n")
	}
}

func cleanBadBuilds(buildId string) {
	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	url := p48CoreService + "/badbuilds"
	var response entities.BadBuildCleanResponse

	request, err := req.NewClient().NewRequest().SetQueryParam("id", buildId).SetSuccessResult(&response).SetHeader(jfApikeyHeader, configFile.APIKey).Delete(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to get secret key!")
		return
	}

	fmt.Println("Bad build cleaned successfully:")
	fmt.Println("  Build ID: " + response.BuildId)
	for _, v := range response.Functions {
		fmt.Println("  Function ID: " + v)
	}

	fmt.Println("\nDone!")
	return
}

func checkIfFunctionExists(name string) {
	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	url := p48CoreService + "/exists?name=" + name
	var response entities.ExistsResponse

	request, err := req.NewClient().NewRequest().SetSuccessResult(&response).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to get Secret Key!")
		return
	}

	fmt.Printf("Function %s exists: %t\n", name, response.Exists)
}

func getLibrary(details string, readme bool) {

	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	//Get the library
	if details == "" {
		url := p48CoreService + "/library"
		var response entities.LibraryResponse

		request, err := req.NewClient().NewRequest().SetSuccessResult(&response).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
		if err != nil {
			fmt.Println("\tError calling out to service", err)
			return
		}

		if request.StatusCode != 200 {
			fmt.Println("\tAn error happened when attempting to get Secret Key!")
			return
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Name", "Id", "Owner", "Versions", "Created At", "Latest Change", "Description"})
		for _, library := range response.LibraryItem {
			t.AppendRow([]interface{}{library.Name, library.FunctionId, library.Owner, library.Versions, library.CreatedAt.Format("01-01-2006"), library.LastRelease.Format("01-01-2006"), library.Description})
			//t.AppendSeparator()
		}

		t.SetStyle(table.StyleColoredYellowWhiteOnBlack)
		t.Render()

		if len(response.BadBuilds) > 0 {

			headerBold := color.New(color.BgHiRed, color.Bold).SprintFunc()
			fmt.Printf("\n\n%s\n", headerBold("Bad Builds:"))
			greenBold := color.New(color.FgGreen, color.Underline).SprintFunc()
			yellowBold := color.New(color.FgYellow, color.Bold, color.BgBlack).SprintFunc()

			for _, b := range response.BadBuilds {
				fmt.Printf("  %s %s\n", greenBold("Build Id:"), b.BuildId)
				fmt.Printf("  %s %s\n", greenBold("Created At:"), b.CreatedAt.Format(time.RFC1123))
				fmt.Printf("  %s %s\n", greenBold("Name:"), b.Name)
				fmt.Printf("  %s %s\n", greenBold("Function ID:"), b.FunctionId)
				fmt.Printf("  %s %s\n", greenBold("Error Message:"), yellowBold(b.ErrorMessage))
			}
		}

		fmt.Println("Done!")
		return
	}

	functionId := details

	url := p48CoreService + "/library/" + functionId
	var fd entities.LibraryItemDetailsResponse

	request, err := req.NewClient().NewRequest().SetSuccessResult(&fd).SetHeader(jfApikeyHeader, configFile.APIKey).Get(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tCannot find library item requested, is the name correct?")
		return
	}

	var readmePlaceHolder []byte

	greenBold := color.New(color.FgGreen, color.Underline).SprintFunc()
	redBold := color.New(color.FgRed, color.Underline).SprintFunc()

	fmt.Printf("%s %s\n", greenBold("Function Name:"), fd.Name)
	fmt.Printf("%s %s\n", greenBold("Function ID:"), fd.FunctionId)
	fmt.Printf("%s %s\n", greenBold("Owner:"), fd.Owner)
	fmt.Printf("%s %s\n", greenBold("Owner Description:"), fd.OwnerDescription)
	fmt.Printf("%s %d\n", greenBold("Version Count:"), fd.VersionCount)
	fmt.Printf("%s %s\n", greenBold("Created At:"), fd.CreatedAt.Format(time.RFC1123))
	fmt.Printf("%s %s\n", greenBold("Updated At:"), fd.UpdatedAt.Format(time.RFC1123))
	fmt.Println(greenBold("Versions:"))
	for _, v := range fd.Versions {

		fmt.Printf("%s %s\n", greenBold("Description:"), v.Description)
		fmt.Printf("%s %s\n", greenBold("Entry Point:"), v.EntryPoint)
		fmt.Printf("  %s %d\n", greenBold("Version:"), v.Version)
		fmt.Printf("  %s %t\n", greenBold("Latest:"), v.Latest)
		for _, s := range v.Sizes {
			fmt.Printf("    %s %s\n", greenBold("FunctionId:"), s.FunctionId)
			fmt.Printf("    %s %s\n", greenBold("Function URL:"), webUi+fd.FunctionId)
			fmt.Printf("    %s %s\n", greenBold("URL:"), jellyfaasEndpoint+s.FunctionId+"/"+fd.FunctionId)
		}
		fmt.Printf("  %s %s\n", greenBold("Release Date:"), v.ReleaseDate.Format(time.RFC1123))
		fmt.Printf("  %s %s\n", greenBold("Runtime:"), v.Runtime)

		if v.Latest {
			if v.ReadMeFileEncoded != "" {
				fmt.Printf("  %s %s\n", greenBold("Readme File:"), "Found")
				readmePlaceHolder, err = base64.StdEncoding.DecodeString(v.ReadMeFileEncoded)
			} else {
				fmt.Printf("  %s %s\n", greenBold("Readme File:"), "Not found")
			}

			if v.ChangeLogFileEncoded != "" {
				fmt.Printf("  %s %s\n", greenBold("ChangeLog File:"), "found")
			} else {
				fmt.Printf("  %s %s\n", greenBold("ChangeLog File:"), "Not found")
			}
			fmt.Println()
			fmt.Println("  " + greenBold("Requirements:"))
			fmt.Printf("    %s %s\n", greenBold("Request Type:"), v.Requirements.RequestType)
			if v.Requirements.InputType != nil {
				fmt.Printf("    %s %s\n", greenBold("Input Type:"), *v.Requirements.InputType)
			}
			for _, p := range v.Requirements.QueryParams {
				fmt.Printf("     %s %s, Required: %v\n", greenBold("Query Param:"), p.Name, p.Required)
			}

			if v.Requirements.InputJsonSchemaEncoded != nil {
				encodedSchema, err := base64.StdEncoding.DecodeString(*v.Requirements.InputJsonSchemaEncoded)
				if err != nil {
					fmt.Println("    " + redBold("Input Schema: Error getting schema, please contact support"))
				}
				fmt.Println("    "+greenBold("Input Schema:"), string(encodedSchema))

				if v.Requirements.InputJsonExample != nil {
					decoded, err := base64.StdEncoding.DecodeString(*v.Requirements.InputJsonExample)
					if err != nil {
						fmt.Println("    " + redBold("Input JSON Example: Error getting example data, please contact support"))
					}
					fmt.Println("    "+greenBold("Input JSON Example:"), string(decoded))
				}
			}

			if v.Requirements.InputFileSchema != nil {
				var fileSchema entities.FileSchema
				err := json.Unmarshal(*v.Requirements.InputFileSchema, &fileSchema)
				if err != nil {
					fmt.Println("    " + greenBold("Input File Schema: Error getting schema, please contact support"))
				}

				fmt.Println("    "+greenBold("Input File Description:"), fileSchema.Description)
				fmt.Println("    "+greenBold("Input File Required:"), fileSchema.Required)
				array := strings.Join(fileSchema.Extensions, ", ")
				fmt.Println("    "+greenBold("Input File Extensions:"), array)
			}

			if v.Requirements.OutputJsonSchemaEncoded != nil {
				encodedSchema, err := base64.StdEncoding.DecodeString(*v.Requirements.OutputJsonSchemaEncoded)
				if err != nil {
					fmt.Println("    " + redBold("Error decoding output JSON schema"))
				}
				fmt.Println("    "+greenBold("Output Schema:"), string(encodedSchema))

				if *v.Requirements.OutputJsonExample != "" {
					fmt.Println("    "+greenBold("Output JSON Example:"), string(*v.Requirements.OutputJsonExample))
				}
				//encodedOutputExample, err := base64.StdEncoding.DecodeString(*v.Requirements.OutputJsonExample)
				//if err != nil {
				//	fmt.Println("    " + redBold("Error decoding output example JSON schema"))
				//}

			}

			if v.Requirements.OutputFileSchema != nil {
				var fileSchema entities.FileSchema
				err := json.Unmarshal(*v.Requirements.OutputFileSchema, &fileSchema)
				if err != nil {
					fmt.Println("    " + greenBold("Input Output Schema: Error getting schema, please contact support"))
				}

				fmt.Println("    "+greenBold("Output File Description:"), fileSchema.Description)
				array := strings.Join(fileSchema.Extensions, ", ")
				fmt.Println("    "+greenBold("Output File Extensions:"), array)
			}
		}
		fmt.Println()
	}

	if readmePlaceHolder != nil && readme {
		output, err := mdToANSI(string(readmePlaceHolder))
		if err == nil {
			fmt.Println()
			fmt.Println(output)
		}
	}

	return

}

func deployFunction(filename string, wait bool) {

	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	fmt.Println("\n\tDeploying function " + filename)

	var functionResponse entities.DeployedFunctionResponse
	var errorDetails entities.ErrorDetails
	url := p48CoreService + "/upload"

	request, err := req.NewClient().NewRequest().SetSuccessResult(&functionResponse).SetErrorResult(&errorDetails).SetFile("file", filename).SetHeader(jfApikeyHeader, configFile.APIKey).Post(url)
	if err != nil {
		fmt.Println("\tError calling out to backend deployment service")
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to get deploy, this normally happens when an upgrade is already in progress or another error")
		fmt.Printf("\tSupport ID: %s \nError: %s\n", errorDetails.ErrorId, errorDetails.ErrorMessage)
		return
	}

	for _, v := range functionResponse.DeployedDetails {
		functionName := strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
		fmt.Printf("\tFunction URL: %s%s\n", webUi, functionName)
		fmt.Printf("\tAPI Endpoint: %s\n", v.FunctionUrl)
	}

	if functionResponse.New {
		fmt.Println("\tFunction is a new function, and is currently deploying.")
	} else {
		fmt.Printf("\tFunction upgrading, current version is: %d, new version will be %d \n\n", functionResponse.CurrentVersion, functionResponse.DeployingVersion)
	}

	if wait {
		fmt.Println("Waiting for function to be ready..")

		var opIds []string
		for _, v := range functionResponse.DeployedDetails {
			opIds = append(opIds, v.Opid)
		}

		checkIfDeployedSuccessfully(functionResponse.FunctionId, opIds, configFile.APIKey)
	}

}

func setPublishedState(id string, state bool) {
	configFile, err := readP48KeyFile()
	if err != nil {
		return
	}

	var url = p48CoreService

	if state {
		url = url + "/library/publish/" + id
	} else {
		url = url + "/library/withdraw/" + id
	}
	var response = entities.LibraryItemDetailsResponse{}

	request, err := req.NewClient().NewRequest().SetSuccessResult(&response).SetHeader(jfApikeyHeader, configFile.APIKey).Put(url)
	if err != nil {
		fmt.Println("\tError calling out to service", err)
		return
	}

	if request.StatusCode != 200 {
		fmt.Println("\tAn error happened when attempting to publish function")
		return
	}

	fmt.Printf("\tFunction published successfully\n")
}

func checkIfDeployedSuccessfully(functionId string, opsLink []string, key string) {

	type opsLinkStatus struct {
		Complete   bool
		opsUrl     string
		opsLink    string
		functionId string
	}

	var opsLinkStatuses []opsLinkStatus

	for _, v := range opsLink {
		opsLinkStatuses = append(opsLinkStatuses, opsLinkStatus{Complete: false, opsLink: p48CoreService + "/upload/" + v + "/" + functionId})
	}

	client := req.NewClient()

	for i := 1; i < maxOpsLoops+1; i++ {
		for index, v := range opsLinkStatuses {

			if v.Complete {
				continue
			}

			ops := Operations{}

			request, err := client.NewRequest().SetSuccessResult(&ops).SetHeader(jfApikeyHeader, key).Get(v.opsLink)
			if err != nil || request.StatusCode != 200 {
				fmt.Println("\tError calling backend service to validate status", err)
				return
			}

			if ops.Status == "DEPLOYED" {
				opsLinkStatuses[index].Complete = true
			} else {
				fmt.Printf("Count %d/%d : Operation is not complete, waiting for function to be ready, status : %s\n", i, maxOpsLoops, ops.Status)
			}

		}

		//Check if all operations are complete
		allComplete := true
		for _, v := range opsLinkStatuses {
			if !v.Complete {
				allComplete = false
			}
		}

		if allComplete {
			fmt.Println("\n\tOperation is complete, function(s) is ready to be used!")
			return
		}

		time.Sleep(30 * time.Second)
	}
}

func readP48KeyFile() (*Config, error) {
	filePath, err := getHiddenFileLocation()
	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func getHiddenFileLocation() (*string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return nil, err
	}

	filePath := filepath.Join(homeDir, hiddenDataFile)
	return &filePath, err
}

func writeP48KeyFile(config *Config) error {
	filename, err := getHiddenFileLocation()
	buf, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(*filename), os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(*filename, buf, 0644)
}

func getMetricsBySize(size string, response entities.LibraryItemResponse) (int, time.Duration) {

	for _, v := range *response.Sizes {
		if v.Size == size {
			return v.InvocationCount, v.AvgResponseTimeMs
		}
	}
	return 0, time.Duration(0)
}

func createProject(functionName, language, destinationDir string, always bool) {
	// Check if the directory exists
	_, err := readP48KeyFile()
	if err != nil {
		return
	}

	greenBold := color.New(color.FgGreen, color.Bold).SprintFunc()
	redBold := color.New(color.FgRed, color.Bold).SprintFunc()

	finalPath := filepath.Join(destinationDir, functionName)

	if !always {
		if _, err := os.Stat(finalPath); !os.IsNotExist(err) {
			fmt.Printf("  %s %s\n", redBold("Error, folder already exists:"), finalPath)
			return
		}

		// Create the directory
		if err := os.MkdirAll(finalPath, 0755); err != nil {
			fmt.Printf("  %s %s\n", redBold("Error, failed to create directory:"), finalPath)
			return
		}
	}

	// Clone the GitHub repository
	repoUrl := p48templatesRepo
	tempDir := destinationDir + "/.temp-repo"
	if err := gitClone(repoUrl, tempDir); err != nil {
		fmt.Printf("  %s \n", redBold("Error, failed to clone repository"))
		return
	}
	defer os.RemoveAll(tempDir)

	var srcDir string

	switch language {
	case "go", "golang":
		srcDir = filepath.Join(tempDir, goTemplate)
	case "python":
		srcDir = filepath.Join(tempDir, pythonTemplate)
	case "php":
		srcDir = filepath.Join(tempDir, phpTemplate)
	case "nodejs", "javascript", "js", "node.js", "node":
		srcDir = filepath.Join(tempDir, nodeTemplate)
	case "ruby":
		srcDir = filepath.Join(tempDir, rubyTemplate)
	case "java":
		srcDir = filepath.Join(tempDir, javaTemplate)
	case "dotnet", "csharp", "c#", "dn":
		srcDir = filepath.Join(tempDir, dotnetTemplate)
	default:
		fmt.Printf("  %s %s\n", redBold("Error, Language is not supported:"), language)
		return
	}

	// Copy the required language folder
	if err := copyDir(srcDir, finalPath); err != nil {
		fmt.Printf("  %s %v\n", redBold("Error, Failed to copy folder:"), err)
		return
	}

	// Update the jellyfaas.json file
	if err := updateSpecFile(finalPath, functionName); err != nil {
		fmt.Printf("  %s %v\n", redBold("Error, Failed to update spec file:"), err)
		return
	}

	fmt.Println(greenBold("\nProject created successfully!\nPlease read the README.md for getting started."))
}

func gitClone(repoUrl, dest string) error {
	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:      repoUrl,
		Progress: os.Stdout,
	})
	return err
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func updateSpecFile(destinationDir, functionName string) error {
	specFilePath := filepath.Join(destinationDir, specfile)
	specData, err := os.ReadFile(specFilePath)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %v", err)
	}

	var spec entities.JellySpec
	if err := json.Unmarshal(specData, &spec); err != nil {
		return fmt.Errorf("failed to unmarshal spec file: %v", err)
	}

	spec.ShortName = functionName
	//remove spaces from shortname or underscores.
	spec.ShortName = strings.ReplaceAll(spec.ShortName, " ", "")
	spec.ShortName = strings.ReplaceAll(spec.ShortName, "_", "")

	updatedSpecData, err := json.MarshalIndent(&spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated spec file: %v", err)
	}

	if err := os.WriteFile(specFilePath, updatedSpecData, 0644); err != nil {
		return fmt.Errorf("failed to write updated spec file: %v", err)
	}

	return nil
}

func getProjectName(destinationDir string, mode int) (string, error) {

	specFilePath := filepath.Join(destinationDir, specfile)
	specData, err := os.ReadFile(specFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open spec file: %v", err)
	}
	var spec entities.JellySpec
	if err := json.Unmarshal(specData, &spec); err != nil {
		return "", fmt.Errorf("failed to unmarshal spec file: %v", err)
	}
	return spec.ShortName, nil

}

func zipProjectAndDeploy(destinationDir string, overwrite bool, deploy bool, wait bool) {

	_, err := readP48KeyFile()
	if err != nil {
		return
	}

	dirToZip := filepath.Join(destinationDir)

	//Check the mode, are we using p48spec.yaml or jellyspec.json
	var mode int
	jellySpecPath := filepath.Join(destinationDir, "jellyspec.json")

	if _, err := os.Stat(jellySpecPath); err == nil {
		mode = 1
	} else {
		// Neither file exists
		fmt.Println(glg.Red("Error: jellyspec.json not found in the directory (did you supply the right folder name?)"))
		return
	}

	projName, err := getProjectName(destinationDir, mode)

	if err != nil {
		fmt.Println(glg.Red("Error reading the jellyspec.json: " + err.Error()))
		return
	}

	zipFileName := projName + ".zip"
	//remove _ and spaces from the zip file name
	zipFileName = strings.ReplaceAll(zipFileName, " ", "")
	zipFileName = strings.ReplaceAll(zipFileName, "_", "")

	excludePatterns := []string{".git", ".idea", "vendor", "node_modules", ".temp-repo"}

	if err := validatePaths(dirToZip, zipFileName, overwrite); err != nil {
		fmt.Println(glg.Red("Validation error: " + err.Error()))
		return
	}

	err = zipDirectory(dirToZip, zipFileName, excludePatterns, mode)
	if err != nil {
		fmt.Println(glg.Red("Error zipping directory: " + err.Error()))
		return
	} else {
		fmt.Println(glg.Green("Directory zipped successfully!"))
		fmt.Println(glg.Green("Zip file: " + zipFileName))
		fmt.Println(glg.Green("Ready for deploy:"))
	}

	if deploy {
		deployFunction(zipFileName, wait)
	}

	fmt.Println(glg.Green("This usually takes a few mins, you can always see if you have\nand issue with 'jellyfaas builds list' to see if you had a deploy problem.\n"))
}

func validatePaths(source, target string, overwrite bool) error {
	// Check if the source directory exists
	info, err := os.Stat(source)
	if os.IsNotExist(err) {
		return fmt.Errorf("source directory %s does not exist", source)
	}
	if !info.IsDir() {
		return fmt.Errorf("source %s is not a directory", source)
	}

	// Check if the target zip file already exists
	if _, err := os.Stat(target); err == nil {

		if !overwrite {
			return fmt.Errorf("target zip file %s already exists", target)
		}
		os.Remove(target)
	}
	return nil
}

func zipDirectory(source, target string, excludePatterns []string, mode int) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	baseDir, err := getProjectName(source, mode)
	if err != nil {
		return err
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the target zip file
		if path == target {
			return nil
		}

		// Skip excluded patterns
		for _, pattern := range excludePatterns {
			if strings.Contains(path, pattern) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
			header.Name = filepath.ToSlash(header.Name)
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}

func mdToANSI(markdown string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
	)
	if err != nil {
		return "", err
	}

	return renderer.Render(markdown)
}

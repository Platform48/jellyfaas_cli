package entities

import (
	"encoding/json"
	"github.com/invopop/jsonschema"
	"time"
)

type BadBuildsItemResponse struct {
	CreatedAt    time.Time `json:"createdAt" bson:"createdAt"`
	BuildId      string    `json:"buildId" bson:"buildId"`
	Name         string    `json:"name" bson:"name"`
	FunctionId   string    `json:"functionId" bson:"functionId"`
	ErrorMessage string    `json:"errorDetails" bson:"errorDetails"`
}

type BadBuildResponse struct {
	Count     int                     `json:"count" bson:"count"`
	BadBuilds []BadBuildsItemResponse `json:"badBuilds" bson:"badBuilds"`
}

type BadBuildCleanResponse struct {
	BuildId   string   `json:"buildId"`
	Functions []string `json:"functions"`
}

type LibraryResponse struct {
	Count       int                     `json:"count" bson:"count"`
	LibraryItem []LibraryItemResponse   `json:"libraryItem" bson:"libraryItem"`
	BadBuilds   []BadBuildsItemResponse `json:"badBuilds" bson:"badBuilds"`
	UniqueTags  []string                `json:"uniqueTags,omitempty" bson:"uniqueTags"`
}

type LibraryItemResponse struct {
	CreatedAt        time.Time                          `json:"createdAt" bson:"createdAt"`
	Name             string                             `json:"name" bson:"name"`
	Description      string                             `json:"description" bson:"description"`
	Owner            string                             `json:"owner" bson:"owner"`
	OwnerDescription string                             `json:"ownerDescription" bson:"ownerDescription"`
	FunctionId       string                             `json:"functionId" bson:"functionId"`
	Versions         int                                `json:"versions" bson:"versions"`
	LastRelease      time.Time                          `json:"lastRelease" bson:"lastRelease"`
	Active           *bool                              `json:"active,omitempty" bson:"active"`
	Published        *bool                              `json:"published,omitempty" bson:"published"`
	Sizes            *[]VersionDetailsResponseSizeStats `json:"sizes,omitempty" bson:"sizes"`
	AvgRating        *int                               `json:"avgRating,omitempty" bson:"avgRating"`
	RatingCount      *int                               `json:"ratingCount,omitempty" bson:"ratingCount"`
	ReadMeFile       *[]byte                            `json:"readmeFile,omitempty" bson:"readmeFile"`
	Tags             []string                           `json:"tags,omitempty" bson:"tags"`
	Ai               bool                               `json:"ai" bson:"ai"`
}

type VersionDetailsResponse struct {
	Version              int                           `json:"version"`
	Latest               bool                          `json:"latest"`
	ReleaseDate          time.Time                     `json:"releaseDate"`
	Runtime              string                        `json:"runtime"`
	Requirements         FunctionRequirement           `json:"requirements,omitempty"`
	Sizes                []VersionDetailsResponseSizes `json:"sizes"`
	LanguageSamples      []LanguageSamples             `json:"languageSamples,omitempty"`
	ChangeLogFileEncoded string                        `json:"changeLog,omitempty"`
	ChangeLogDetails     ChangeLogDetails              `json:"changeLogDetails,omitempty"`
	ReadMeFileEncoded    string                        `json:"readmeFile"`
	EntryPoint           string                        `json:"entryPoint"`
	Description          string                        `json:"description"`
}

type VersionDetailsResponseSizes struct {
	Size       string `json:"size"`
	FunctionId string `json:"functionId"`
}

type ChangeLogDetails struct {
	Date    string `json:"date"`
	Added   []byte `json:"added"`
	Updated []byte `json:"updated"`
	Removed []byte `json:"removed"`
}

type LanguageSamples struct {
	Language string `yaml:"language,omitempty" json:"language,omitempty"`
	Params   string `yaml:"params,omitempty" json:"params,omitempty"`
	Input    string `yaml:"input,omitempty" json:"input,omitempty"`
	Output   string `yaml:"output,omitempty" json:"output,omitempty"`
	SdkCode  string `yaml:"sdkCode,omitempty" json:"sdkCode,omitempty"`
}

type LibraryItemDetailsResponse struct {
	CreatedAt        time.Time                `json:"createdAt"`
	UpdatedAt        time.Time                `json:"updatedAt"`
	Name             string                   `json:"name"`
	FunctionId       string                   `json:"functionId"`
	Owner            string                   `json:"owner"`
	OwnerDescription string                   `json:"ownerDescription"`
	VersionCount     int                      `json:"versionCount"`
	Versions         []VersionDetailsResponse `json:"versions"`
	Active           bool                     `json:"active"`
	Published        bool                     `json:"published"`
	AvgRating        int                      `json:"avgRating"`
	RatingCount      int                      `json:"ratingCount"`
	Tags             []string                 `json:"tags"`
	Ai               bool                     `json:"ai"`
}

type FileSchema struct {
	Extensions  []string `yaml:"extensions,omitempty" json:"extensions,omitempty"`
	Required    bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
}

type VersionDetailsResponseSizeStats struct {
	Size              string        `json:"size"`
	InvocationCount   int           `json:"invocationCount"`
	AvgResponseTimeMs time.Duration `json:"avgResponseTime"`
}

type ExistsResponse struct {
	FunctionName string `json:"functionName"`
	Exists       bool   `json:"exists"`
}

type FunctionRequirement struct {
	RequestType             string             `yaml:"requestType" json:"requestType" bson:"requestType"` // Must be one of GET, POST, PUT, DELETE
	InputType               *string            `yaml:"inputType" json:"inputType" bson:"inputType"`       // Must be one of JSON, FILE, NONE etc
	OutputType              *string            `yaml:"outputType" json:"outputType" bson:"outputType"`    // Must be one of JSON, FILE, NONE etc
	QueryParams             []*QueryParam      `yaml:"queryParams,omitempty" json:"queryParams,omitempty" bson:"queryParams"`
	InputSchema             *json.RawMessage   `yaml:"inputSchema,omitempty" json:"inputSchema,omitempty" bson:"inputSchema"`
	InputJsonSchema         *jsonschema.Schema `yaml:"inputJsonSchema,omitempty" json:"inputJsonSchema,omitempty" bson:"inputJsonSchema"`
	InputJsonSchemaEncoded  *string            `yaml:"inputJsonSchemaEncoded,omitempty" json:"inputJsonSchemaEncoded,omitempty" bson:"inputJsonSchemaEncoded"`
	InputJsonExample        *string            `yaml:"inputJsonExample,omitempty" json:"inputJsonExample,omitempty" bson:"inputJsonExample"`
	OutputSchema            *json.RawMessage   `yaml:"outputSchema,omitempty" json:"outputSchema,omitempty" bson:"outputSchema"`
	OutputJsonSchema        *jsonschema.Schema `yaml:"outputJsonSchema,omitempty" json:"outputJsonSchema,omitempty" bson:"outputJsonSchema"`
	OutputJsonSchemaEncoded *string            `yaml:"outputJsonSchemaEncoded,omitempty" json:"outputJsonSchemaEncoded,omitempty" bson:"outputJsonSchemaEncoded"`
	OutputJsonExample       *string            `yaml:"outputJsonExample,omitempty" json:"outputJsonExample,omitempty" bson:"outputJsonExample"`
	InputFileSchema         *json.RawMessage   `yaml:"inputFileSchema,omitempty" json:"inputFileSchema,omitempty" bson:"inputFileSchema"`
	InputFileSchemaEncoded  *string            `yaml:"inputFileSchemaEncoded,omitempty" json:"inputFileSchemaEncoded,omitempty" bson:"inputFileSchemaEncoded"`
	OutputFileSchema        *json.RawMessage   `yaml:"outputFileSchema,omitempty" json:"outputFileSchema,omitempty" bson:"outputFileSchema"`
	OutputFileSchemaEncoded *string            `yaml:"outputFileSchemaEncoded,omitempty" json:"outputFileSchemaEncoded,omitempty" bson:"outputFileSchemaEncoded"`
	ReadMeEncoded           *string            `yaml:"readmeEncoded,omitempty" json:"readmeEncoded,omitempty" bson:"readMeEncoded"`
	FunctionCallingEncoded  *string            `yaml:"functionCallingEncoded,omitempty" json:"functionCallingEncoded,omitempty" bson:"functionCallingEncoded"`
}

type QueryParam struct {
	Name        string `json:"name" yaml:"name"`
	Required    bool   `json:"required" yaml:"required"`
	Description string `json:"description" yaml:"description"`
	ExampleData string `json:"exampleData" yaml:"exampleData"`
}

type DeployedFunctionResponse struct {
	Function         string            `json:"function"`
	FunctionId       string            `json:"function_id"`
	DeployedDetails  []DeployedDetails `json:"deployedDetails"`
	CurrentVersion   int               `json:"currentVersion,omitempty"`
	DeployingVersion int               `json:"deployingVersion,omitempty"`
	New              bool              `json:"new"`
}

type DeployedDetails struct {
	Size        string
	Opid        string `json:"opid"`
	FunctionUrl string `json:"urlLocation"`
}

type ErrorDetails struct {
	ErrorId      string `json:"errorId"`
	ErrorMessage string `json:"errorMessage"`
}

type JellySpec struct {
	Name             string               `yaml:"name" json:"name" bson:"name"`
	JellySpecVersion string               `yaml:"$jellyspecversion" json:"$jellyspecversion,omitempty" bson:"jellyspecversion"`
	SchemaUrl        string               `yaml:"schemaUrl" json:"schemaUrl,omitempty" bson:"schemaUrl"`
	ShortName        string               `yaml:"shortname" json:"shortname" bson:"shortName"` //Less that 15 chars, no special. and over 5 checked below
	Runtime          string               `yaml:"runtime" json:"runtime" bson:"runtime"`
	EntryPoint       string               `yaml:"entrypoint" json:"entrypoint" bson:"entryPoint"`
	Description      string               `yaml:"description,omitempty" json:"description,omitempty" bson:"description"`
	Requirements     *FunctionRequirement `yaml:"requirements,omitempty" json:"requirements,omitempty" bson:"requirements"`
	Tags             []string             `yaml:"tags,omitempty" json:"tags,omitempty" bson:"tags"`
	Ai               bool                 `yaml:"ai,omitempty" json:"ai,omitempty" bson:"ai"`
	Overload         bool                 `yaml:"overload,omitempty" json:"overload,omitempty" bson:"overload"`
}

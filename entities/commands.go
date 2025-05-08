package entities

import "time"

type Options struct {
	User      UserCommands       `command:"user" description:"User related commands"`
	Secret    Secret             `command:"secret" description:"secret command"`
	Library   ListLibraryCommand `command:"library" description:"List library"`
	Deploy    DeployCommands     `command:"deploy" description:"Deploy related commands"`
	Publish   PublishCommands    `command:"publish" description:"Publish related commands"`
	Token     GetTokenCommand    `command:"token" description:"Setup a token in the .jellyfaas file"`
	Spec      CreateSpecCommand  `command:"spec" description:"Spec related commands"`
	BadBuilds BadBuildsCommand   `command:"builds" description:"Bad builds related commands"`
	Create    CreateCommand      `command:"create" description:"Create a new function"`
	Zip       ZipCommand         `command:"zip" description:"Zip a function"`
	Exists    Exists             `command:"exists" description:"Check if a function exists"`
	Base64    Base64Command      `command:"base64" description:"Base64 encode/decode a string"`
	Version   VersionCommand     `command:"version" short:"v" description:"Show the JellyFaaS CLI version"`
}

type VersionCommand struct{}

type Base64Command struct {
	Encode string `short:"e" long:"encode" description:"Encode a string" required:"false"`
	Decode string `short:"d" long:"decode" description:"Decode a string" required:"false"`
}

type Exists struct {
	Name string `short:"n" long:"name" description:"Name of the function" required:"true"`
}

type ZipCommand struct {
	Source    string `short:"s" long:"source" description:"Source of the function" required:"false" default:"."`
	Overwrite bool   `short:"o" long:"overwrite" description:"Overwrite the zip file" required:"false"`
	Deploy    bool   `short:"d" long:"deploy" description:"Deploy the function" required:"false"`
	Wait      bool   `short:"w" command:"wait" description:"Wait for a function to be ready" required:"false"`
}

type BadBuildsCommand struct {
	List  ListBadBuildsCommand  `command:"list" description:"List bad builds"`
	Clean CleanBadBuildsCommand `command:"clean" description:"Clean bad builds"`
}

type CreateCommand struct {
	Name        string `short:"n" long:"name" description:"Name of the function" required:"true"`
	Language    string `short:"l" long:"language" description:"Language of the function" required:"true"`
	Destination string `short:"d" long:"destination" description:"Destination of the function" required:"true"`
	Always      bool   `short:"a" long:"always" description:"Always create function, even if directory exists"`
}

type ListBadBuildsCommand struct{}

type CleanBadBuildsCommand struct {
	BuildId string `short:"b" long:"buildId" description:"Build ID" required:"true"`
}

type CreateSpecCommand struct {
	Name string `short:"j" long:"json" description:"JSON string to convert" required:"true"`
	Raw  bool   `short:"r" long:"raw" description:"Raw output, allowing you to pipe or pbcopy for example." required:"false"`
	Flat bool   `short:"f" long:"flat" description:"Flat output, allowing you to pipe or pbcopy for example." required:"false"`
}

type UserCommands struct {
	Create CreateUserCommand `command:"create" description:"Create a new user"`
	Delete DeleteUserCommand `command:"delete" description:"Delete a user"`
	List   ListUsersCommand  `command:"list" description:"List users"`
}

type Secret struct{}

type ListLibraryCommand struct {
	Details string `short:"d" long:"details" description:"Details of the library" required:"false"`
	ReadMe  bool   `short:"r" long:"readme" description:"View the Readme of the library" required:"false"`
}

type DeployCommands struct {
	Wait    bool   `short:"w" command:"wait" description:"Wait for a function to be ready" required:"false"`
	ZipFile string `short:"z" command:"zipfile" description:"Zip file to upload" required:"true"`
}

type PublishCommands struct {
	ID    string `short:"i" command:"id" description:"ID of the library item" required:"true"`
	State bool   `short:"s" command:"state" description:"State of the library item" required:"true"`
}

type CreateUserCommand struct {
	Email string `short:"e" long:"email" description:"Email of the user" required:"true"`
	Name  string `short:"n" long:"name" description:"Name of the user" required:"true"`
}

type DeleteUserCommand struct {
	Email string `short:"e" long:"email" description:"Email of the user" required:"true"`
}

type ListUsersCommand struct{}

type GetTokenCommand struct{}

type UserRequest struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Email string `json:"email" `
}

type UserResponse struct {
	Password string `json:"password"`
}

type APIKeyResponse struct {
	Apikey string `json:"apikey" bson:"apikey"`
}

type UserDetails struct {
	Name      string    `json:"name" omitempty:"true"`
	Email     string    `json:"email" omitempty:"true"`
	CreatedAt time.Time `json:"createdAt" omitempty:"true"`
	UpdatedAt time.Time `json:"updatedAt" omitempty:"true"`
}

type Entity struct {
	Entities []UserDetails `json:"entities" omitempty:"true"`
}

type TokenResponse struct {
	Token  string `json:"token"`
	Expiry string `json:"expiry"`
}

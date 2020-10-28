package model

import jwt "github.com/dgrijalva/jwt-go"

type ErrorResponse struct {
	Err string
}

type Exception struct {
	Message string `json:"message"`
}

type User struct {
	ID       uint
	Name     string
	Email    string `json:"Email:varchar(100);unique_index"`
	Password string `json:"Password"`
}

type Token struct {
	UserID uint
	Name   string
	Email  string
	*jwt.StandardClaims
}

type ManifestReq struct {
	ID      string     `json:"ID"`
	Name    string     `json:"Name"`
	Modules []Manifest `json:"Modules"`
}

type Manifest struct {
	Index     int      `json:"Index"`
	Name      string   `json:"Name"`
	Tag       string   `json:"Tag"`
	ImageID   string   `json:"ImageID"`
	ImageName string   `json:"ImageName"`
	Options   Option   `json:"Options"`
	Arguments Argument `json:"Arguments"`
}

type Argument []struct {
	Arg string `json:"opt"`
	Val bool   `json:"val"`
}

type Option []struct {
	Arg string `json:"arg"`
	Val bool   `json:"val"`
}

// type DatasourceConfig struct {
// 	Type     string `json:"Type"`
// 	Protocol string `json:"Protocol"`
// 	Port     string `json:"Port"`
// 	Param1   string `json:"Param1"`
// 	Param2   string `json:"Param2"`
// 	Param3   string `json:"Param3"`
// 	Param4   string `json:"Param4"`
// 	Param5   string `json:"Param5"`
// 	Param6   string `json:"Param6"`
// }
